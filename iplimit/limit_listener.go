package iplimit

import (
	"container/list"
	"errors"
	"net"
	"sync"
)

// LimitListener returns a Listener that accepts at most n simultaneous
// connections from the provided Listener.
func LimitListener(l net.Listener, n int) net.Listener {
	return &limitListener{
		Listener: l,
		limit:    n,
		ipPool:   make(map[string]int),
	}
}

type limitListener struct {
	net.Listener
	limit     int
	ipPool    map[string]int
	closeOnce sync.Once
}

func (l *limitListener) Accept() (net.Conn, error) {
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}

		if l.putConn(c) {
			return &limitListenerConn{Conn: c, release: l.removeConn}, nil
		}
		c.Close()
	}
}

func (l *limitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() {})
	return err
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func(c net.Conn)
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(func() { l.release(l) })
	return err
}

func (l *limitListener) cleanPool() {
	inactiveIPs := list.New()
	for key := range l.ipPool {
		count := l.ipPool[key]
		if count <= 0 {
			inactiveIPs.PushBack(key)
		}
	}
	for i := inactiveIPs.Front(); i != nil; i = i.Next() {
		delete(l.ipPool, i.Value.(string))
	}
}

func (l *limitListener) putConn(c net.Conn) bool {
	ip, err := getIP(c)
	if err != nil {
		return false
	}
	ipKey := ip.String()

	_, ok := l.ipPool[ipKey]
	if ok {
		l.ipPool[ipKey]++
		return true
	}

	// Check size of pool
	if len(l.ipPool) < l.limit {
		l.ipPool[ipKey] = 1
		return true
	}
	l.cleanPool()
	// Check again
	if len(l.ipPool) < l.limit {
		l.ipPool[ipKey] = 1
		return true
	}

	return false
}

func (l *limitListener) removeConn(c net.Conn) {
	ip, err := getIP(c)
	if err != nil {
		return
	}
	ipKey := ip.String()

	_, ok := l.ipPool[ipKey]
	if ok {
		l.ipPool[ipKey]--
	}
}

func getIP(c net.Conn) (net.IP, error) {
	remoteAddr := c.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		if serr, ok := err.(*net.AddrError); ok && serr.Err == "missing port in address" { // It's not critical try parse
			ip = remoteAddr
		} else {
			return nil, err
		}
	}

	// Parse the ip address string into a net.IP.
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.New("unable to parse address")
	}

	return parsedIP, nil
}
