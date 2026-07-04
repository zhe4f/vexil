package receiver

import (
	"sync"
	"sync/atomic"
	"time"
	"context"

	"vexil_go/network"
	"vexil_go/protocol"
)

func AckFlusher(ctx context.Context, c *network.FramedConn, totalBytes *atomic.Int64, totalSize int64,
    doneCh chan struct{}, compCh chan struct{}, wg *sync.WaitGroup, ackInterval time.Duration) {
    defer func() {
        wg.Done()
    }()
    ticker := time.NewTicker(ackInterval)
    defer ticker.Stop()

    lastSent := int64(0)
    doneClosed := false

	for {
		select {
		case <-ctx.Done():
        	return	
		case <-doneCh:
			doneClosed = true
			current := totalBytes.Load()
			if current > lastSent {
				ack := protocol.AckRange{Offset: 0, Length: totalSize}
				frame, _ := protocol.EncodeAckRange(ack)
				c.SendFromMemory(frame)
				lastSent = current
			}
			// 连接已关闭则退出
    		return

		case <-compCh:
			return

		case <-ticker.C:
			if doneClosed {
				current := totalBytes.Load()
				if current > lastSent {
					ack := protocol.AckRange{Offset: 0, Length: totalSize}
					frame, _ := protocol.EncodeAckRange(ack)
					c.SendFromMemory(frame)
					lastSent = current
				}
			} else {
				current := totalBytes.Load()
				if current > lastSent {
					ack := protocol.AckRange{Offset: 0, Length: current}
					frame, _ := protocol.EncodeAckRange(ack)
					c.SendFromMemory(frame)
					lastSent = current
				}
			}
		}
	}
}