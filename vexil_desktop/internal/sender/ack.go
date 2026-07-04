package sender

import (
	"context"
	"fmt"
	"time"

	"vexil/internal/network"
	"vexil/internal/protocol"
)

type AckHandler struct {
	conn        *network.FramedConn
	window      *Window
	doneCh      chan struct{}
	err         error
	readTimeout time.Duration
}

func NewAckHandler(ctx context.Context, conn *network.FramedConn, w *Window, readTimeout time.Duration) *AckHandler {
	h := &AckHandler{
		conn:        conn,
		window:      w,
		doneCh:      make(chan struct{}),
		readTimeout: readTimeout,
	}
	go h.loop(ctx)
	return h
}

func (h *AckHandler) Done() chan struct{} { return h.doneCh }
func (h *AckHandler) Err() error          { return h.err }

func (h *AckHandler) loop(ctx context.Context) {
	defer close(h.doneCh)

	totalSize := int64(0)
	for _, c := range h.window.allChunks {
		totalSize += c.size
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		frame, err := h.conn.ReadFrame(h.readTimeout)
		if err != nil {
			// 检查是否被取消
			select {
			case <-ctx.Done():
				return
			default:
			}
			h.err = fmt.Errorf("ACK 处理器: 连接断开: %w", err)
			return
		}

		switch frame.Type {
		case protocol.MSG_ACK_RANGE:
			ack := protocol.DecodeAckRange(frame)
			h.window.UpdateAcked(ack.Offset + ack.Length)

			if ack.Offset+ack.Length >= totalSize {
				return
			}

		case protocol.MSG_RESUME_RANGE:
			ranges, existing := protocol.DecodeResumeRangeEx(frame)
			if len(ranges) > 0 {
				h.window.SetResumeState(ranges, existing)
				h.window.Notify()
			}
		}
	}
}