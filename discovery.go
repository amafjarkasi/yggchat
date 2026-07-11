package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type DiscoveryManager struct {
	ygg        *YggManager
	localPort  string
	quit       chan struct{}
}

func NewDiscoveryManager(ygg *YggManager, listenAddrs []string) *DiscoveryManager {
	// Find first TCP listener port
	port := "9000"
	for _, l := range listenAddrs {
		if strings.HasPrefix(l, "tcp://") {
			parts := strings.Split(l, ":")
			if len(parts) >= 3 {
				port = parts[len(parts)-1]
				break
			}
		}
	}

	return &DiscoveryManager{
		ygg:       ygg,
		localPort: port,
		quit:      make(chan struct{}),
	}
}

func (d *DiscoveryManager) Start() {
	go d.broadcastLoop()
	go d.listenLoop()
}

func (d *DiscoveryManager) Stop() {
	close(d.quit)
}

func (d *DiscoveryManager) broadcastLoop() {
	ticker := time.NewTicker(DiscoveryBeaconSec * time.Second)
	defer ticker.Stop()

	// Multicast to local subnet
	multicastAddr, err := net.ResolveUDPAddr("udp", "224.0.0.50:9999")
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, multicastAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		select {
		case <-d.quit:
			return
		case <-ticker.C:
			ips := getLocalIPs()
			for _, ip := range ips {
				uri := fmt.Sprintf("tcp://%s:%s", ip, d.localPort)
				_, _ = conn.Write([]byte(uri))
			}
		}
	}
}

func (d *DiscoveryManager) listenLoop() {
	multicastAddr, err := net.ResolveUDPAddr("udp", "224.0.0.50:9999")
	if err != nil {
		return
	}

	conn, err := net.ListenMulticastUDP("udp", nil, multicastAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		select {
		case <-d.quit:
			return
		default:
			_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			uri := string(buf[:n])
			if strings.HasPrefix(uri, "tcp://") {
				// Prevent self-peering
				isSelf := false
				ips := getLocalIPs()
				for _, ip := range ips {
					if strings.Contains(uri, ip+":"+d.localPort) {
						isSelf = true
						break
					}
				}

				if !isSelf {
					_ = d.ygg.AddPeer(uri)
				}
			}
		}
	}
}

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	if len(ips) == 0 {
		ips = append(ips, "127.0.0.1")
	}
	return ips
}
