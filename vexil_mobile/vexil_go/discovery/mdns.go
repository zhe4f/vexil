package discovery

import (
	"fmt"
	"os"
	"sync"
	"time"
	"net"
	"strings"

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
	hostname, _ := os.Hostname()
	return &MDNSDiscovery{
		stopCh:   make(chan struct{}),
		hostname: hostname,
	}
}

func NewMDNSDiscoveryWithName(name string) *MDNSDiscovery {
	if name == "" {
		hostname, _ := os.Hostname()
		name = hostname
	}
	return &MDNSDiscovery{
		stopCh:   make(chan struct{}),
		hostname: name,
	}
}

func (d *MDNSDiscovery) Name() string { return "mdns" }

// Start 作为 receiver，注册 mDNS 服务
func (d *MDNSDiscovery) Start(port int) error {
	localIP := getLocalIP()
	if localIP == "" {
		return fmt.Errorf("无法获取本地 IP")
	}

	service, err := mdns.NewMDNSService(
		d.hostname,        // 服务实例名 = 自定义设备名
		mdnsServiceName,
		mdnsDomain,
		"",
		port,
		[]net.IP{net.ParseIP(localIP)},
		[]string{},
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

// Discover 作为 sender，查询 mDNS 服务
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