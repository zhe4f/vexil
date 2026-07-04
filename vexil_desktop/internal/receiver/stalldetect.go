package receiver

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

func StallDetector(ctx context.Context, totalBytes *atomic.Int64, doneCh <-chan struct{}, compCh <-chan struct{},
	stallCh chan<- error, stallInitial time.Duration, stallInterval time.Duration, stallThreshold int) {
	
	ticker := time.NewTicker(stallInterval)
	defer ticker.Stop()

	last := totalBytes.Load()
	stallCount := 0

	initialDeadline := time.After(stallInitial)
	initialDone := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-doneCh:
			return
		case <-compCh:
			return
		case <-initialDeadline:
			initialDone = true
		case <-ticker.C:
			// 再次检查取消
			select {
			case <-ctx.Done():
				return
			default:
			}

			if !initialDone {
				last = totalBytes.Load()
				continue
			}

			current := totalBytes.Load()
			if current == last {
				stallCount++
				if stallCount >= stallThreshold {
					select {
					case stallCh <- fmt.Errorf("transfer stalled (no progress for %v)",
						time.Duration(stallThreshold)*stallInterval):
					default:
					}
					return
				}
			} else {
				stallCount = 0
			}
			last = current
		}
	}
}