package discovery

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// Device 表示发现到的设备
type Device struct {
	Name     string    `json:"name"`      // 主机名
	IP       string    `json:"ip"`        // IPv4 地址
	Port     int       `json:"port"`      // 监听端口
	Source   string    `json:"source"`    // 发现来源: "udp" / "mdns"
	LastSeen time.Time `json:"last_seen"` // 最后活跃时间
}

// Discoverer 发现器接口
type Discoverer interface {
	// Start 启动发现/广播 (receiver 侧)
	Start(port int) error
	// Discover 主动发现设备 (sender 侧)，返回设备通道
	Discover(timeout time.Duration) (<-chan Device, error)
	// Stop 停止
	Stop() error
	// Name 发现器名称
	Name() string
}

// MergeDevices 合并多个发现源的设备列表，去重（按 IP:Port）
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
					// 如果新设备名不是 localhost 且旧的是，或者时间更新，则覆盖
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

// getLocalIP 获取首选本地 IPv4，排除回环和链路本地地址
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	// 回退：遍历网卡，排除回环和链路本地
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