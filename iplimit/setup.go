package iplimit

import (
	"net"
	"os"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

// IPLimit represents a middleware instance
type IPLimit struct {
	Next httpserver.Handler
}

type listener struct {
	net.Listener
}

// Initializes the plugin
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

	// Add middlewares
	siteConfig := httpserver.GetConfig(c)
	siteConfig.AddListenerMiddleware(func(cln caddy.Listener) caddy.Listener {
		ln := LimitListener(cln, config.Max)
		return &listener{ln}
	})
	return nil
}

func (l *listener) File() (*os.File, error) {
	return nil, nil
}
