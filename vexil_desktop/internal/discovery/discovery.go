package discovery

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

type Device struct {
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	Source   string    `json:"source"`
	LastSeen time.Time `json:"last_seen"`
}

type Discoverer interface {
	Start(port int) error
	Discover(timeout time.Duration) (<-chan Device, error)
	Stop() error
	Name() string
}

func MergeDevices(channels ...<-chan Device) []Device {
	seen := make(map[string]*Device)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan Device) {
			defer wg.Done()
			for dev := range c {
				mu.Lock()
				key := fmt.Sprintf("%s:%d", dev.IP, dev.Port)
				if existing, ok := seen[key]; ok {
					if (dev.Name != "localhost" && existing.Name == "localhost") ||
						dev.LastSeen.After(existing.LastSeen) {
						*existing = dev
					}
				} else {
					devCopy := dev
					seen[key] = &devCopy
				}
				mu.Unlock()
			}
		}(ch)
	}
	wg.Wait()

	devices := make([]Device, 0, len(seen))
	for _, d := range seen {
		devices = append(devices, *d)
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})
	return devices
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				if ipNet.IP[0] == 169 && ipNet.IP[1] == 254 {
					continue
				}
				return ipNet.IP.String()
			}
		}
	}
	return ""
}