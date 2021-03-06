package main

import (
	"github.com/caddyserver/caddy/caddy/caddymain"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"

	_ "caddy/iplimit"
)

func main() {
	httpserver.RegisterDevDirective("iplimit", "index")
	caddymain.EnableTelemetry = false
	caddymain.Run()
}
