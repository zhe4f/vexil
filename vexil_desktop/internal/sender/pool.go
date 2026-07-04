package sender

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"vexil/internal/network"
)

type ConnPool struct {
	host             string
	port             int
	conns            []*network.FramedConn
	mu               sync.Mutex
	dialTimeout      time.Duration
	dialRetries      int
	dialRetryBase    time.Duration
	connEstablishGap time.Duration
	tlsConfig        *tls.Config
}

func NewConnPool(host string, port int, dialTimeout time.Duration,
	dialRetries int, dialRetryBase time.Duration,
	connEstablishGap time.Duration, tlsConfig *tls.Config) *ConnPool {
	return &ConnPool{
		host:             host,
		port:             port,
		dialTimeout:      dialTimeout,
		dialRetries:      dialRetries,
		dialRetryBase:    dialRetryBase,
		connEstablishGap: connEstablishGap,
		tlsConfig:        tlsConfig,
	}
}

func (p *ConnPool) DialOne(readBufSize, writeBufSize int) (*network.FramedConn, error) {
	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	var rawConn net.Conn
	var err error

	for retry := 0; retry < p.dialRetries; retry++ {
		if p.tlsConfig != nil {
			rawConn, err = tls.DialWithDialer(
				&net.Dialer{Timeout: p.dialTimeout},
				"tcp", addr, p.tlsConfig,
			)
		} else {
			rawConn, err = net.DialTimeout("tcp", addr, p.dialTimeout)
		}

		if err == nil {
			return network.NewFramedConnWithBuf(rawConn, readBufSize, writeBufSize), nil
		}
		if retry < p.dialRetries-1 {
			waitTime := time.Duration(retry+1) * p.dialRetryBase
			fmt.Printf("  连接失败 (第%d次重试): %v, %v后重试...\n", retry+1, err, waitTime)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("dial failed after %d retries: %w", p.dialRetries, err)
}

func (p *ConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.conns {
		c.Close()
	}
	p.conns = nil
}