package iplimit

import (
	"github.com/caddyserver/caddy"
	"strconv"
)

// Config specifies configuration parsed for Caddyfile
type Config struct {
	Max int
}

func parseConfig(c *caddy.Controller) (Config, error) {
	var config = Config{}
	var err error

	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 1:
			config.Max, err = strconv.Atoi(args[0])
			if err != nil {
				return config, err
			}
		default:
			return config, c.ArgErr()
		}
	}
	return config, nil
}
