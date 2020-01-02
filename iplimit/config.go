package iplimit

import (
	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	cmap "github.com/orcaman/concurrent-map"
)

// Config specifies configuration parsed for Caddyfile
type Config struct {
	*httpserver.SiteConfig
	Max    int
	IPPool *cmap.ConcurrentMap
}

func parseConfig(controller *caddy.Controller) (Config, error) {
	config := Config{
		SiteConfig: httpserver.GetConfig(controller),
	}
	var err error

	for controller.Next() {
		args := controller.RemainingArgs()

		switch len(args) {
		case 1:
			config.Max, err = strconv.Atoi(args[0])
			if err != nil {
				return config, err
			}
		default:
			return config, controller.ArgErr()
		}
	}
	return config, nil
}
