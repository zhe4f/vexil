package tracker

import (
	"sync"
	"time"
)

type Progress struct {
	Percent   float64
	SpeedMBps float64
	Sent      int64
	Total     int64
	ETA       time.Duration
}

type Callback func(p Progress)

type speedSample struct {
	bytes    int64
	duration float64
}

type Tracker struct {
	total        int64
	getBytes     func() int64
	startTime    time.Time
	callback     Callback
	doneCh       chan struct{}
	stopCh       chan struct{}
	stopped      sync.WaitGroup
	lastBytes    int64
	lastTime     time.Time
	samples      []speedSample
	totalBytes   float64
	totalSeconds float64
	maxSamples   int
	interval     time.Duration
}

func New(total int64, numChunks int, chunkSize int, getAcked func() int, getBytes func() int64,
	doneCh chan struct{}, cb Callback, interval time.Duration) *Tracker {
	t := &Tracker{
		total:      total,
		getBytes:   getBytes,
		startTime:  time.Now(),
		callback:   cb,
		doneCh:     doneCh,
		stopCh:     make(chan struct{}),
		maxSamples: 6,
		interval:   interval,
	}
	t.stopped.Add(1)
	go t.loop()
	return t
}

// Stop 停止 tracker，等待 loop 退出
func (t *Tracker) Stop() {
    close(t.stopCh)
    t.stopped.Wait()
}

func (t *Tracker) loop() {
	defer func() {
        t.stopped.Done()
    }()
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	time.Sleep(t.interval)

	t.lastBytes = t.getBytes()
	t.lastTime = time.Now()

	for {
		select {
		case <-t.stopCh:
			return

		case <-t.doneCh:
			if t.callback != nil {
				t.callback(Progress{
					Percent:   100,
					SpeedMBps: 0,
					Sent:      t.total,
					Total:     t.total,
					ETA:       0,
				})
			}
			return

		case <-ticker.C:
			if t.callback == nil {
				continue
			}

			currentBytes := t.getBytes()
			if currentBytes > t.total {
				currentBytes = t.total
			}

			now := time.Now()

			deltaBytes := currentBytes - t.lastBytes
			deltaTime := now.Sub(t.lastTime).Seconds()

			if deltaTime > 0 && deltaBytes > 0 {
				t.addSample(speedSample{
					bytes:    deltaBytes,
					duration: deltaTime,
				})
			}

			t.lastBytes = currentBytes
			t.lastTime = now

			speed := t.smoothSpeed()

			pct := float64(currentBytes) / float64(t.total) * 100
			if pct > 100 {
				pct = 100
			}

			var eta time.Duration
			if speed > 0 && currentBytes < t.total {
				remainingBytes := t.total - currentBytes
				etaSeconds := float64(remainingBytes) / (speed * 1024 * 1024)
				eta = time.Duration(etaSeconds) * time.Second
				if eta%time.Second != 0 {
					eta = eta.Truncate(time.Second) + time.Second
				}
			}

			t.callback(Progress{
				Percent:   pct,
				SpeedMBps: speed,
				Sent:      currentBytes,
				Total:     t.total,
				ETA:       eta,
			})
		}
	}
}

func (t *Tracker) addSample(s speedSample) {
	t.samples = append(t.samples, s)
	t.totalBytes += float64(s.bytes)
	t.totalSeconds += s.duration

	for len(t.samples) > t.maxSamples {
		old := t.samples[0]
		t.totalBytes -= float64(old.bytes)
		t.totalSeconds -= old.duration
		t.samples = t.samples[1:]
	}
}

func (t *Tracker) smoothSpeed() float64 {
	if len(t.samples) == 0 || t.totalSeconds <= 0 {
		return 0
	}
	return (t.totalBytes / t.totalSeconds) / 1024 / 1024
}