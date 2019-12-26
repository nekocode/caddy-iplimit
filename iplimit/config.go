package iplimit

import (
	"github.com/caddyserver/caddy"
	"strconv"
	"time"
)

// Config specifies configuration parsed for Caddyfile
type Config struct {
	MaxAmount int
	MaxAge    time.Duration
}

func parseConfig(c *caddy.Controller) (Config, error) {
	var config = Config{}
	var err error

	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 2:
			config.MaxAmount, err = strconv.Atoi(args[0])
			if err != nil {
				return config, err
			}
			config.MaxAge, err = time.ParseDuration(args[1])
			if err != nil {
				return config, err
			}
		default:
			return config, c.ArgErr()
		}
	}
	return config, nil
}
