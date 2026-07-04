package receiver

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"vexil_go/network"
	"vexil_go/protocol"
)

type ConnPool struct {
	ln            net.Listener
	mu            sync.Mutex
	PeerIP        string
	acceptTimeout time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
}

func NewConnPool(port int, acceptTimeout time.Duration, tlsConfig *tls.Config) (*ConnPool, error) {
	rawLn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("监听端口 %d 失败: %w", port, err)
	}

	var ln net.Listener
	if tlsConfig != nil {
		ln = tls.NewListener(rawLn, tlsConfig)
		fmt.Printf("  [TLS] 已启用 TLS 加密监听\n")
	} else {
		ln = rawLn
	}

	return &ConnPool{
		ln:            ln,
		acceptTimeout: acceptTimeout,
		stopCh:        make(chan struct{}),
	}, nil
}

func (p *ConnPool) AcceptOne(readTimeout time.Duration) (*network.FramedConn, string, error) {
	type result struct {
		conn net.Conn
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		c, e := p.ln.Accept()
		ch <- result{c, e}
	}()

	var rawConn net.Conn
	select {
	case r := <-ch:
		if r.err != nil {
			return nil, "", r.err
		}
		rawConn = r.conn
	case <-time.After(p.acceptTimeout):
		return nil, "", fmt.Errorf("accept timeout")
	case <-p.stopCh:
		return nil, "", fmt.Errorf("stopped")
	}

	fc := network.NewFramedConn(rawConn)
	host, _, _ := net.SplitHostPort(rawConn.RemoteAddr().String())

	frame, err := fc.ReadFrame(readTimeout)
	if err != nil {
		fc.Close()
		return nil, "", fmt.Errorf("read CONN_ROLE: %w", err)
	}
	if frame.Type != protocol.MSG_CONN_ROLE {
		fc.Close()
		return nil, "", fmt.Errorf("expected CONN_ROLE, got %s", protocol.MessageName(frame.Type))
	}
	role, err := protocol.DecodeConnRole(frame)
	if err != nil {
		fc.Close()
		return nil, "", fmt.Errorf("decode CONN_ROLE: %w", err)
	}

	p.mu.Lock()
	if p.PeerIP == "" {
		p.PeerIP = host
	}
	p.mu.Unlock()

	return fc, role.Role, nil
}

func (p *ConnPool) Close() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
	p.ln.Close()
}