package iplimit

import (
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

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
		return LimitListener(cln, config.Max)
	})
	return nil
}
