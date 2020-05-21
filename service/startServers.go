package service

import (
	"github.com/joho/godotenv"
	"github.com/peake100/stalkgateway-go/service/servers"
)

func StartServers() *servers.Monitor {
	// Load .env file. We are only using .env files in development, so we can suppress
	// errors here -- it won't affect production.
	_ = godotenv.Load()

	monitor := servers.NewServiceMonitor()
	monitor.StartServers()
	return monitor
}
