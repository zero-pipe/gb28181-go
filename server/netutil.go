package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// GuessLocalIP returns the first non-loopback IPv4 address.
func GuessLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String()
		}
	}
	return ""
}

func detectLocalIPForRemote(remoteIP string) string {
	remoteIP = strings.TrimSpace(remoteIP)
	if remoteIP == "" {
		return ""
	}
	conn, err := net.DialTimeout("udp", net.JoinHostPort(remoteIP, "9"), time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()
	if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok && ua.IP != nil {
		if v4 := ua.IP.To4(); v4 != nil {
			return v4.String()
		}
	}
	return ""
}

func (s *Server) resolveInviteLocalIP(peer Peer) string {
	candidates := []string{s.localIP, peer.LocalIP, peer.SDPIP}
	for _, ip := range candidates {
		ip = strings.TrimSpace(ip)
		if ip != "" && ip != "0.0.0.0" && ip != "127.0.0.1" {
			return ip
		}
	}
	if peer.IP != "" {
		if ip := detectLocalIPForRemote(peer.IP); ip != "" {
			return ip
		}
	}
	return GuessLocalIP()
}

func (s *Server) requirePeerAddr(peer Peer) error {
	if peer.DeviceID == "" || peer.IP == "" || peer.Port <= 0 {
		return fmt.Errorf("invalid peer: need DeviceID/IP/Port")
	}
	return nil
}
