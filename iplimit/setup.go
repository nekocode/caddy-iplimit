package iplimit

import (
	"container/list"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

// IPLimit represents a middleware instance
type IPLimit struct {
	Next   httpserver.Handler
	Config Config
	IPPool map[string]int64
}

// Init initializes the plugin
func init() {
	caddy.RegisterPlugin("iplimit", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	config, err := parseConfig(c)
	if err != nil {
		return err
	}

	// Create new middleware
	newMiddleWare := func(next httpserver.Handler) httpserver.Handler {
		return &IPLimit{
			Next:   next,
			Config: config,
			IPPool: make(map[string]int64),
		}
	}
	// Add middleware
	cfg := httpserver.GetConfig(c)
	cfg.AddMiddleware(newMiddleWare)

	return nil
}

func (ipl IPLimit) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	clientIP, err := getClientIP(r, true)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	ipKey := clientIP.String()
	now := time.Now().Unix()

	// Check pool size
	if len(ipl.IPPool) < ipl.Config.MaxAmount {
		ipl.IPPool[ipKey] = now
		return ipl.Next.ServeHTTP(w, r)
	}

	// Clear inactive IPs
	inactiveIPs := list.New()
	for key := range ipl.IPPool {
		timestamp := ipl.IPPool[key]
		if (timestamp + int64(ipl.Config.MaxAge.Seconds())) <= now {
			inactiveIPs.PushBack(key)
		}
	}
	for i := inactiveIPs.Front(); i != nil; i = i.Next() {
		delete(ipl.IPPool, i.Value.(string))
	}

	// Check pool size again
	if len(ipl.IPPool) < ipl.Config.MaxAmount {
		ipl.IPPool[ipKey] = now
		return ipl.Next.ServeHTTP(w, r)
	}

	// If client IP is in pool
	_, ok := ipl.IPPool[ipKey]
	if ok {
		ipl.IPPool[ipKey] = now
		return ipl.Next.ServeHTTP(w, r)
	}

	return http.StatusTooManyRequests, nil
}

func getClientIP(r *http.Request, strict bool) (net.IP, error) {
	var ip string

	// Use the client ip from the 'X-Forwarded-For' header, if available.
	if fwdFor := r.Header.Get("X-Forwarded-For"); fwdFor != "" && !strict {
		ips := strings.Split(fwdFor, ", ")
		ip = ips[0]
	} else {
		// Otherwise, get the client ip from the request remote address.
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			if serr, ok := err.(*net.AddrError); ok && serr.Err == "missing port in address" { // It's not critical try parse
				ip = r.RemoteAddr
			} else {
				log.Printf("Error when SplitHostPort: %v", serr.Err)
				return nil, err
			}
		}
	}

	// Parse the ip address string into a net.IP.
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.New("unable to parse address")
	}

	return parsedIP, nil
}
