package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"vexil/internal/config"
)

const udpMagic = "VEXIL_DISCOVER"

type udpMessage struct {
	Magic    string `json:"magic"`
	Cmd      string `json:"cmd"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	IP       string `json:"ip"`
}

type UDPDiscovery struct {
	conn     *net.UDPConn
	stopCh   chan struct{}
	stopOnce sync.Once
	hostname string
}

func NewUDPDiscovery() *UDPDiscovery {
	return NewUDPDiscoveryWithName("")
}

func NewUDPDiscoveryWithName(name string) *UDPDiscovery {
	if name == "" {
		name, _ = os.Hostname()
	}
	return &UDPDiscovery{
		stopCh:   make(chan struct{}),
		hostname: name,
	}
}

func (d *UDPDiscovery) Name() string { return "udp" }

func (d *UDPDiscovery) Start(port int) error {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: config.UDPBroadcastPort,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("udp listen: %w", err)
	}
	d.conn = conn

	localIP := getLocalIP()

	go func() {
		buf := make([]byte, 1500)
		for {
			select {
			case <-d.stopCh:
				return
			default:
			}

			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, remote, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			var msg udpMessage
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}
			if msg.Magic != udpMagic || msg.Cmd != "query" {
				continue
			}

			resp := udpMessage{
				Magic:    udpMagic,
				Cmd:      "response",
				Hostname: d.hostname,
				Port:     port,
				IP:       localIP,
			}
			respBytes, _ := json.Marshal(resp)

			respAddr := &net.UDPAddr{
				IP:   remote.IP,
				Port: remote.Port,
			}
			if _, err := conn.WriteToUDP(respBytes, respAddr); err != nil {
				fmt.Fprintf(os.Stderr, "UDP 响应失败 (%s:%d): %v\n", remote.IP, remote.Port, err)
			}
		}
	}()

	return nil
}

func (d *UDPDiscovery) Discover(timeout time.Duration) (<-chan Device, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("udp listen: %w", err)
	}

	query := udpMessage{
		Magic: udpMagic,
		Cmd:   "query",
	}
	queryBytes, _ := json.Marshal(query)

	broadcastAddrs := getBroadcastAddrs()
	for _, addr := range broadcastAddrs {
		if _, err := conn.WriteToUDP(queryBytes, &net.UDPAddr{
			IP:   addr,
			Port: config.UDPBroadcastPort,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "UDP 广播失败 (%s): %v\n", addr, err)
		}
	}

	deviceCh := make(chan Device, 10)
	go func() {
		defer conn.Close()
		defer close(deviceCh)

		buf := make([]byte, 1500)
		deadline := time.After(timeout)

		for {
			select {
			case <-deadline:
				return
			default:
			}

			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					select {
					case <-deadline:
						return
					default:
						continue
					}
				}
				return
			}

			var msg udpMessage
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}
			if msg.Magic != udpMagic || msg.Cmd != "response" {
				continue
			}
			if msg.IP == "" {
				continue
			}

			deviceCh <- Device{
				Name:     msg.Hostname,
				IP:       msg.IP,
				Port:     msg.Port,
				Source:   "udp",
				LastSeen: time.Now(),
			}
		}
	}()

	return deviceCh, nil
}

func (d *UDPDiscovery) Stop() error {
	d.stopOnce.Do(func() {
		close(d.stopCh)
		if d.conn != nil {
			d.conn.Close()
		}
	})
	return nil
}

func getBroadcastAddrs() []net.IP {
	var addrs []net.IP
	seen := make(map[string]bool)

	interfaces, err := net.Interfaces()
	if err != nil {
		addrs = append(addrs, net.IPv4bcast)
		return addrs
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}
		addrsList, _ := iface.Addrs()
		for _, a := range addrsList {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				bcast := make(net.IP, 4)
				for i := 0; i < 4; i++ {
					bcast[i] = ipNet.IP.To4()[i] | ^ipNet.Mask[i]
				}
				key := bcast.String()
				if !seen[key] {
					seen[key] = true
					addrs = append(addrs, bcast)
				}
			}
		}
	}

	if len(addrs) == 0 {
		addrs = append(addrs, net.IPv4bcast)
	}
	return addrs
}