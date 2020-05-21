package main

import (
	"github.com/peake100/stalkgateway-go/service"
	"os"
)

func main() {
	// start all server processes
	monitor := service.StartServers()

	// Block until there is a fatal error or a shutdown signal is sent
	monitor.WaitOnShutdown()

	// If there was an error, exit with a code of 1
	if monitor.ErrorsEncountered() {
		os.Exit(1)
	}

	// Otherwise exit with a code of 0
	os.Exit(0)
}
