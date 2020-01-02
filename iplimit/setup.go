package iplimit

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	cmap "github.com/orcaman/concurrent-map"
)

type ipLimitMiddleware struct {
	next   httpserver.Handler
	config *Config
}

var restartConfigs map[string]*Config = make(map[string]*Config)

// Initializes the plugin
func init() {
	caddy.RegisterPlugin("iplimit", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(controller *caddy.Controller) error {
	config, err := parseConfig(controller)
	if err != nil {
		return err
	}
	vhost := config.Addr.VHost()

	// Hook restart
	controller.OnRestart(func() error {
		delete(restartConfigs, vhost)
		restartConfigs[vhost] = &config
		return nil
	})

	// Create or retrieve IP pool
	oldConfig, ok := restartConfigs[vhost]
	if ok {
		config.IPPool = oldConfig.IPPool
	} else {
		ipPool := cmap.New()
		config.IPPool = &ipPool
	}

	// Add middlewares
	config.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &ipLimitMiddleware{
			next:   next,
			config: &config,
		}
	})
	return nil
}

func (ipl ipLimitMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	clientIP, err := getClientIP(r, true)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	ip := clientIP.String()

	if !ipl.config.addIP(ip) {
		return http.StatusTooManyRequests, nil
	}
	rlt, err := ipl.next.ServeHTTP(w, r)
	ipl.config.removeIP(ip)
	return rlt, err
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
