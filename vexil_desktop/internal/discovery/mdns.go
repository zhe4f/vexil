package discovery

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	mdnsServiceName = "_vexil._tcp"
	mdnsDomain      = "local."
)

type MDNSDiscovery struct {
	server   *mdns.Server
	stopCh   chan struct{}
	stopOnce sync.Once
	hostname string
}

func NewMDNSDiscovery() *MDNSDiscovery {
	return NewMDNSDiscoveryWithName("")
}

func NewMDNSDiscoveryWithName(name string) *MDNSDiscovery {
	if name == "" {
		name, _ = os.Hostname()
	}
	return &MDNSDiscovery{
		stopCh:   make(chan struct{}),
		hostname: name,
	}
}

func (d *MDNSDiscovery) Name() string { return "mdns" }

func (d *MDNSDiscovery) Start(port int) error {
	hostname := d.hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	localIP := getLocalIP()
	if localIP == "" {
		return fmt.Errorf("无法获取本地 IP")
	}

	info := []string{
		fmt.Sprintf("hostname=%s", hostname),
		fmt.Sprintf("port=%d", port),
	}

	service, err := mdns.NewMDNSService(
		hostname,
		mdnsServiceName,
		mdnsDomain,
		"",
		port,
		[]net.IP{net.ParseIP(localIP)},
		info,
	)
	if err != nil {
		return fmt.Errorf("创建 mDNS 服务失败: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{
		Zone: service,
	})
	if err != nil {
		return fmt.Errorf("启动 mDNS 服务失败: %w", err)
	}
	d.server = server
	return nil
}

func (d *MDNSDiscovery) Discover(timeout time.Duration) (<-chan Device, error) {
	deviceCh := make(chan Device, 10)

	go func() {
		defer close(deviceCh)

		entriesCh := make(chan *mdns.ServiceEntry, 10)
		go func() {
			mdns.Lookup(mdnsServiceName, entriesCh)
			close(entriesCh)
		}()

		deadline := time.After(timeout)

		for {
			select {
			case entry, ok := <-entriesCh:
				if !ok {
					return
				}
				if entry.AddrV4 == nil {
					continue
				}
				hostname := entry.Name
				if idx := strings.Index(hostname, "._"); idx >= 0 {
					hostname = hostname[:idx]
				}
				deviceCh <- Device{
					Name:     hostname,
					IP:       entry.AddrV4.String(),
					Port:     entry.Port,
					Source:   "mdns",
					LastSeen: time.Now(),
				}
			case <-deadline:
				return
			}
		}
	}()

	return deviceCh, nil
}

func (d *MDNSDiscovery) Stop() error {
	d.stopOnce.Do(func() {
		close(d.stopCh)
		if d.server != nil {
			d.server.Shutdown()
		}
	})
	return nil
}