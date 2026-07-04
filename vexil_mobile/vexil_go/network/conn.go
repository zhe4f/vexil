package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"vexil_go/protocol"
)

type FramedConn struct {
	conn        net.Conn
	writeMu     sync.Mutex
	readBufSize int
}

func NewFramedConn(conn net.Conn) *FramedConn {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
	}
	return &FramedConn{conn: conn}
}

func NewFramedConnWithBuf(conn net.Conn, readBufSize, writeBufSize int) *FramedConn {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
		tcp.SetReadBuffer(readBufSize)
		tcp.SetWriteBuffer(writeBufSize)
	}
	return &FramedConn{conn: conn, readBufSize: readBufSize}
}

func (c *FramedConn) SendFromMemory(f *protocol.Frame) error {
	data := f.Encode()
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if _, err := c.conn.Write(lenBuf); err != nil {
		return err
	}
	_, err := c.conn.Write(data)
	return err
}

func (c *FramedConn) ReadFrame(readTimeout time.Duration) (*protocol.Frame, error) {
	lenBuf := make([]byte, 4)
	if _, err := c.readFull(lenBuf, readTimeout); err != nil {
		return nil, err
	}
	totalLen := binary.BigEndian.Uint32(lenBuf)
	if totalLen < protocol.HeaderSize || totalLen > protocol.MaxPayloadLen+protocol.HeaderSize {
		return nil, fmt.Errorf("invalid frame length: %d", totalLen)
	}
	data := make([]byte, totalLen)
	if _, err := c.readFull(data, readTimeout); err != nil {
		return nil, err
	}
	return protocol.DecodeHeader(data)
}

func (c *FramedConn) readFull(buf []byte, readTimeout time.Duration) (int, error) {
	total := 0
	for total < len(buf) {
		c.conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := c.conn.Read(buf[total:])
		total += n
		if err != nil && total >= len(buf) {
			return total, nil
		}
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (c *FramedConn) RawConn() net.Conn { return c.conn }
func (c *FramedConn) Close() error       { return c.conn.Close() }