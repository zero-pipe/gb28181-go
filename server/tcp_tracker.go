package server

import (
	"net"
	"sync"
)

// tcpPeer tracks which device registered over a given TCP connection so a
// disconnect can be reported immediately instead of waiting for heartbeat
// timeout (9.1.1: NAT-inside devices keep a TCP registration connection; a
// broken TCP connection means the device is offline).
type tcpPeerRegistry struct {
	mu      sync.Mutex
	byAddr  map[string]string // "ip:port" -> deviceID
	byID    map[string]string // deviceID -> "ip:port"
	onClose func(deviceID string)
}

func newTCPPeerRegistry(onClose func(deviceID string)) *tcpPeerRegistry {
	return &tcpPeerRegistry{
		byAddr:  make(map[string]string),
		byID:    make(map[string]string),
		onClose: onClose,
	}
}

func (r *tcpPeerRegistry) bind(deviceID, addr string) {
	if deviceID == "" || addr == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if prev, ok := r.byID[deviceID]; ok && prev != addr {
		delete(r.byAddr, prev)
	}
	r.byAddr[addr] = deviceID
	r.byID[deviceID] = addr
}

func (r *tcpPeerRegistry) unbind(deviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if addr, ok := r.byID[deviceID]; ok {
		delete(r.byAddr, addr)
		delete(r.byID, deviceID)
	}
}

// closed returns the device that owned addr, if any, and forgets it.
func (r *tcpPeerRegistry) closed(addr string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	deviceID, ok := r.byAddr[addr]
	if !ok {
		return ""
	}
	delete(r.byAddr, addr)
	delete(r.byID, deviceID)
	return deviceID
}

// trackingListener wraps a net.Listener and reports closed TCP connections.
type trackingListener struct {
	net.Listener
	notify func(remoteAddr string)
}

func (l *trackingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &trackingConn{Conn: conn, notify: l.notify}, nil
}

// trackingConn reports its remote address exactly once when the connection
// stops being readable (EOF/reset) or is closed locally.
type trackingConn struct {
	net.Conn
	notify func(remoteAddr string)
	once   sync.Once
}

func (c *trackingConn) close() {
	c.once.Do(func() {
		if c.notify != nil {
			c.notify(c.Conn.RemoteAddr().String())
		}
	})
}

func (c *trackingConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if err != nil {
		c.close()
	}
	return n, err
}

func (c *trackingConn) Close() error {
	c.close()
	return c.Conn.Close()
}
