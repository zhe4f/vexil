package sender

import (
	"context"
	"sync"
	"sync/atomic"

	"vexil_go/protocol"
)

type Window struct {
	allChunks    []chunkMeta
	windowSize   int64
	numConns     int
	nextChunkIdx atomic.Int64
	inFlight     atomic.Int64
	sentBytes    atomic.Int64
	startBytes   atomic.Int64
	started      bool
	chunkChs     []chan chunkMeta
	doneCh       chan struct{}
	notifyCh     chan struct{}
	skipRanges   []protocol.AckRange
	skipMu       sync.RWMutex
	infoCh       chan<- chunkInfo
	closeOnce    sync.Once
}

func NewWindow(allChunks []chunkMeta, windowSize int64, numConns int, infoCh chan<- chunkInfo) *Window {
	chs := make([]chan chunkMeta, numConns)
	for i := 0; i < numConns; i++ {
		chs[i] = make(chan chunkMeta, 1)
	}
	return &Window{
		allChunks:  allChunks,
		windowSize: windowSize,
		numConns:   numConns,
		chunkChs:   chs,
		doneCh:     make(chan struct{}),
		notifyCh:   make(chan struct{}, 1),
		infoCh:     infoCh,
	}
}

func (w *Window) ChunkCh(workerIdx int) chan chunkMeta {
	return w.chunkChs[workerIdx]
}

func (w *Window) Done() chan struct{} { return w.doneCh }

func (w *Window) SentBytes() int64 {
	return w.sentBytes.Load()
}

func (w *Window) StartBytes() int64 {
	return w.startBytes.Load()
}

func (w *Window) Notify() {
	select {
	case w.notifyCh <- struct{}{}:
	default:
	}
}

func (w *Window) UpdateAcked(ackedOffset int64) {
	w.Notify()
}

func (w *Window) SetResumeState(ranges []protocol.AckRange, existing []protocol.BlockHash) {
	w.skipMu.Lock()
	w.skipRanges = ranges

	for _, b := range existing {
		w.infoCh <- chunkInfo{offset: b.Offset, hash: b.Hash}
	}

	var baseBytes int64
	for _, b := range existing {
		for _, c := range w.allChunks {
			if c.offset == b.Offset {
				baseBytes += c.size
				break
			}
		}
	}
	w.sentBytes.Store(baseBytes)

	if len(ranges) > 0 {
		for i, chunk := range w.allChunks {
			if chunk.offset >= ranges[0].Offset {
				w.nextChunkIdx.Store(int64(i))
				break
			}
		}
	}
	w.skipMu.Unlock()
}

func (w *Window) shouldSkip(offset int64) bool {
	w.skipMu.RLock()
	defer w.skipMu.RUnlock()

	if len(w.skipRanges) == 0 {
		return false
	}

	for _, r := range w.skipRanges {
		if offset >= r.Offset && offset < r.Offset+r.Length {
			return false
		}
	}
	return true
}

func (w *Window) Run(ctx context.Context) {
	if !w.started {
		w.startBytes.Store(w.sentBytes.Load())
		w.started = true
	}

	go func() {
		<-ctx.Done()
		w.closeOnce.Do(func() {
			for _, ch := range w.chunkChs {
				close(ch)
			}
		})
	}()

	totalChunks := int64(len(w.allChunks))
	round := int64(0)

	for {
		idx := w.nextChunkIdx.Load()
		inFlight := w.inFlight.Load()

		if inFlight < w.windowSize && idx < totalChunks {
			chunk := w.allChunks[idx]

			if w.shouldSkip(chunk.offset) {
				w.nextChunkIdx.Store(idx + 1)
				round++
				continue
			}

			w.inFlight.Add(chunk.size)
			w.nextChunkIdx.Store(idx + 1)

			workerIdx := int(round % int64(w.numConns))
			round++

			select {
			case w.chunkChs[workerIdx] <- chunk:
			case <-ctx.Done():
				w.closeOnce.Do(func() {
					for _, ch := range w.chunkChs {
						close(ch)
					}
				})
				return
			}
			continue
		}

		if idx >= totalChunks && inFlight <= 0 {
			w.closeOnce.Do(func() {
				for _, ch := range w.chunkChs {
					close(ch)
				}
			})
			return
		}

		select {
		case <-ctx.Done():
			w.closeOnce.Do(func() {
				for _, ch := range w.chunkChs {
					close(ch)
				}
			})
			return
		case <-w.notifyCh:
		}
	}
}