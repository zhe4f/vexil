package receiver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"vexil/internal/network"
	"vexil/internal/protocol"
)

func (r *Receiver) RecvWorker(ctx context.Context, idx int, c *network.FramedConn, manifest []FileOffsetInfo,
	resumePath string, totalSize int64,
	doneCh chan struct{}, compCh chan struct{}, compOnce *sync.Once,
	wg *sync.WaitGroup, doneOnce *sync.Once,
	writeErrCh chan *WriteError) {

	defer func() {
        wg.Done()
    }()

	lastSave := time.Now()

	for {
		// 检查取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		frame, err := c.ReadFrame(r.tcfg.ReadTimeout)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			r.drainAndCheckComplete(c, doneCh, compCh, doneOnce, compOnce, r.tcfg.DrainRetries, r.tcfg.DrainRetryTimeout)
			return
		}

		switch frame.Type {
		case protocol.MSG_DATA:
			compactOffset := int64(frame.Offset)
			payloadLen := int64(len(frame.Payload))

			r.startOnce.Do(func() {
				r.startTime = time.Now()
			})

			_, loaded := r.receivedSet.LoadOrStore(compactOffset, true)
			if loaded {
				continue
			}

			remaining := frame.Payload
			writeOffset := compactOffset
			var writeErr error
			for len(remaining) > 0 {
				f, locOff, fileEnd := locateFileByOffset(manifest, writeOffset)
				if f == nil {
					writeErr = fmt.Errorf("无法定位文件偏移 %d", writeOffset)
					break
				}
				avail := fileEnd - writeOffset
				toWrite := len(remaining)
				if int64(toWrite) > avail {
					toWrite = int(avail)
				}
				if toWrite == 0 {
					break
				}
				if _, err := f.WriteAt(remaining[:toWrite], locOff); err != nil {
					writeErr = fmt.Errorf("WriteAt offset=%d len=%d: %w", locOff, toWrite, err)
					break
				}
				remaining = remaining[toWrite:]
				writeOffset += int64(toWrite)
			}

			if writeErr != nil {
				we := &WriteError{
					Offset:  compactOffset,
					Size:    payloadLen,
					FileIdx: -1,
					Err:     writeErr,
				}
				select {
				case writeErrCh <- we:
				default:
					fmt.Fprintf(os.Stderr, "写入错误通道已满，丢弃: %v\n", we)
				}
				return
			}

			chunkHash := sha256.Sum256(frame.Payload)
			r.addChunk(compactOffset, payloadLen, chunkHash)

			r.totalBytes.Add(payloadLen)
			fmt.Printf("  [进度] totalBytes=%d/%d\n", r.totalBytes.Load(), totalSize)

			now := time.Now()
			if now.Sub(lastSave) >= r.tcfg.ResumeSaveInterval {
				r.saveResumeState(resumePath, totalSize, r.tcfg.MaxChunk)
				lastSave = now
			}

			if r.totalBytes.Load() >= totalSize && r.isComplete(totalSize, r.tcfg.MaxChunk) {
				doneOnce.Do(func() {
					close(doneCh)
				})
			}

		case protocol.MSG_COMPLETE:
			if !r.isComplete(totalSize, r.tcfg.MaxChunk) || r.totalBytes.Load() < totalSize {
				fmt.Printf("  [警告] COMPLETE 到达但数据不完整 (received=%d/%d, chunks=%v)\n",
					r.totalBytes.Load(), totalSize, r.isComplete(totalSize, r.tcfg.MaxChunk))
			}

			r.saveResumeState(resumePath, totalSize, r.tcfg.MaxChunk)

			expectedSum := string(frame.Payload)
			if expectedSum != "" {
				actual := r.computeFinalHash()
				if actual != expectedSum {
					fmt.Printf("  [校验] ❌ SHA-256 不匹配\n  expected: %s\n  actual:   %s\n", expectedSum, actual)
					os.Remove(resumePath)
				} else {
					fmt.Println("  [校验] ✅ SHA-256 匹配")
					os.Remove(resumePath)
				}
			} else {
				fmt.Println("  [校验] ⚠️  发送端未提供校验和，跳过校验")
				os.Remove(resumePath)
			}

			doneOnce.Do(func() {
				close(doneCh)
			})
			compOnce.Do(func() {
				close(compCh)
			})
			return
		}
	}
}

func (r *Receiver) drainAndCheckComplete(c *network.FramedConn, doneCh chan struct{}, compCh chan struct{},
	doneOnce *sync.Once, compOnce *sync.Once, drainRetries int, drainRetryTimeout time.Duration) {

	rawConn := c.RawConn()

	for i := 0; i < drainRetries; i++ {
		if tcp, ok := rawConn.(*net.TCPConn); ok {
			tcp.SetReadDeadline(time.Now().Add(drainRetryTimeout))
		}

		frame, err := c.ReadFrame(drainRetryTimeout)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		if frame.Type == protocol.MSG_COMPLETE {
			fmt.Println("  [恢复] 在缓冲区中发现遗漏的 COMPLETE 帧")

			expectedSum := string(frame.Payload)
			if expectedSum != "" {
				actual := r.computeFinalHash()
				if actual != expectedSum {
					fmt.Printf("  [校验] ❌ SHA-256 不匹配\n  expected: %s\n  actual:   %s\n", expectedSum, actual)
				} else {
					fmt.Println("  [校验] ✅ SHA-256 匹配")
				}
			}

			doneOnce.Do(func() {
				close(doneCh)
			})
			compOnce.Do(func() {
				close(compCh)
			})
			return
		}

		if frame.Type == protocol.MSG_DATA {
			fmt.Printf("  [恢复] 在缓冲区中发现遗漏的 DATA 帧 (offset=%d, len=%d)\n",
				frame.Offset, len(frame.Payload))
		}
	}

	if tcp, ok := rawConn.(*net.TCPConn); ok {
		tcp.SetReadDeadline(time.Time{})
	}
}

func (r *Receiver) saveResumeState(resumePath string, totalSize int64, maxChunk int64) {
	r.collectMu.RLock()
	entries := make([]chunkEntry, len(r.collected))
	copy(entries, r.collected)
	r.collectMu.RUnlock()

	offsets := make([]int64, len(entries))
	hashes := make([][32]byte, len(entries))
	sizes := make([]int64, len(entries))
	for i, entry := range entries {
		offsets[i] = entry.offset
		hashes[i] = entry.hash
		sizes[i] = entry.size
	}

	state := &ResumeState{
		FileSize:        totalSize,
		TotalChunks:     uint32(len(entries)),
		ContiguousBytes: r.findFirstMissing(offsets, sizes, totalSize, maxChunk),
		ChunkHashes:     hashes,
		ChunkOffsets:    offsets,
		ChunkSizes:      sizes,
	}
	SaveResume(resumePath, state)
}

func (r *Receiver) restoreAllHashes(offsets []int64, hashes [][32]byte, sizes []int64) {
	r.collectMu.Lock()
	defer r.collectMu.Unlock()

	r.collected = r.collected[:0]
	r.receivedSet.Range(func(key, value interface{}) bool {
		r.receivedSet.Delete(key)
		return true
	})

	for i := 0; i < len(hashes) && i < len(offsets); i++ {
		sz := r.tcfg.MaxChunk
		if i < len(sizes) && sizes[i] > 0 {
			sz = sizes[i]
		}
		r.collected = append(r.collected, chunkEntry{offset: offsets[i], size: sz, hash: hashes[i]})
		r.receivedSet.Store(offsets[i], true)
	}
}