package sender

import (
	"context"
	"crypto/sha256"
	"os"
	"sync"
	"time"

	"vexil_go/network"
	"vexil_go/protocol"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 16*1024*1024+1024)
		return &buf
	},
}

func (s *Sender) streamWorker(ctx context.Context, idx int, c *network.FramedConn, files []FileInfo, openFiles []*os.File,
	w *Window, wg *sync.WaitGroup, infoCh chan<- chunkInfo) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-w.ChunkCh(idx):
			if !ok {
				return
			}

			bufPtr := bufferPool.Get().(*[]byte)
			buf := (*bufPtr)[:chunk.size]

			n, err := readAtCrossFile(files, openFiles, chunk.offset, buf)
			if err != nil || n <= 0 {
				bufferPool.Put(bufPtr)
				w.inFlight.Add(-chunk.size)
				if w.inFlight.Load() < 0 {
					w.inFlight.Store(0)
				}
				w.Notify()
				continue
			}

			hash := sha256.Sum256(buf[:n])
			c.SendFromMemory(&protocol.Frame{
				Type:    protocol.MSG_DATA,
				Offset:  uint64(chunk.offset),
				Payload: buf[:n],
			})

			s.startOnce.Do(func() {
				s.startTime = time.Now()
			})

			infoCh <- chunkInfo{offset: chunk.offset, hash: hash}
			bufferPool.Put(bufPtr)

			w.inFlight.Add(-chunk.size)
			if w.inFlight.Load() < 0 {
				w.inFlight.Store(0)
			}
			w.sentBytes.Add(chunk.size)
			w.Notify()
		}
	}
}

func readAtCrossFile(files []FileInfo, openFiles []*os.File, offset int64, buf []byte) (int, error) {
	totalRead := 0
	remaining := len(buf)
	currentOffset := offset

	for remaining > 0 {
		fileIdx, localOff := locateFile(files, currentOffset)
		if fileIdx < 0 {
			return totalRead, nil
		}

		fileEnd := files[fileIdx].Size
		availInFile := fileEnd - localOff
		toRead := remaining
		if int64(toRead) > availInFile {
			toRead = int(availInFile)
		}

		n, err := openFiles[fileIdx].ReadAt(buf[totalRead:totalRead+toRead], localOff)
		if err != nil && n <= 0 {
			return totalRead, err
		}
		totalRead += n
		currentOffset += int64(n)
		remaining -= n

		if n < toRead {
			continue
		}
	}

	return totalRead, nil
}

func locateFile(files []FileInfo, compactOffset int64) (int, int64) {
	var cumulative int64
	for i, f := range files {
		if compactOffset < cumulative+f.Size {
			return i, compactOffset - cumulative
		}
		cumulative += f.Size
	}
	return -1, 0
}