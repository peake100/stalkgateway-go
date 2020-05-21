package servers

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// We'll use this to run the services, with a bonus that we can use it during tests
// as well.

type Monitor struct {
	osExitSignal chan os.Signal

	// Shutdown signal that comes from outside the monitor
	shutDownMaster chan interface{}

	// shutdown signal to be received by the rest service
	shutdownRest chan interface{}

	// signal to be sent by the rest gateway when it is complete
	restShutdownComplete chan interface{}

	// a channel for the rest gateway to send errors across to the monitor
	restErrs chan error

	// a waitgroup for programs outside the monitor to block on until shutdown is
	// complete.
	shutdownComplete *sync.WaitGroup

	// hold the shutdown context so we don't hang on a shutdown.
	shutdownCtx context.Context

	// a list of all errors sent to the monitor from the rest gateway.
	restErrList []error

	// STATE INFO
	shutdownInProgress bool
	restDone           bool
}

// Start monitoring the servers
func (monitor *Monitor) StartServers() {
	monitor.shutdownCtx = context.Background()
	go RunRest(monitor)
	go monitor.monitorServers()
}

// Sends shutdown signals. Forces shutdown after 10 seconds.
func (monitor *Monitor) ShutdownServers() {
	monitor.shutDownMaster <- nil
}

// BLocks until servers are shut down.
func (monitor *Monitor) WaitOnShutdown() {
	monitor.shutdownComplete.Wait()
}

func (monitor *Monitor) ErrorsEncountered() bool {
	return len(monitor.restErrList) != 0
}

func (monitor *Monitor) processEvent() (timeout bool) {
	// Wait for events
	select {
	case <-monitor.shutdownCtx.Done():
		log.Println("server shutdown timed out - exiting")
		return true
	case <-monitor.osExitSignal:
		log.Println("received exit signal")
		monitor.ShutdownServers()
	case <-monitor.shutDownMaster:
		log.Println("received shutdown order")
		monitor.shutdownRest <- nil
	case err := <-monitor.restErrs:
		monitor.restErrList = append(monitor.restErrList, err)
		log.Println("error from rest server:", err)
		monitor.ShutdownServers()
	case <-monitor.restShutdownComplete:
		log.Println("rest server shutdown complete")
		monitor.restDone = true
	}

	return false
}

func (monitor *Monitor) monitorServers() {
	defer close(monitor.shutDownMaster)
	defer close(monitor.shutdownRest)
	defer close(monitor.restShutdownComplete)
	defer close(monitor.restErrs)

	// Block until one of the servers encounters a fatal error or a shutdown signal
	// is received.

	for {
		timeout := monitor.processEvent()

		// When we first receive a shutdown event, lets create a timeout context.
		if !monitor.shutdownInProgress {
			log.Println("starting 10 second shutdown timeout")
			ctx := context.Background()
			ctx, _ = context.WithTimeout(ctx, time.Second*10)
			monitor.shutdownCtx = ctx
			monitor.shutdownInProgress = true
		}

		if timeout || monitor.restDone {
			break
		}
	}

	log.Println("shutdown complete")
	// Signal to outside waiters that the servers are shutdown
	monitor.shutdownComplete.Done()
}

func NewServiceMonitor() *Monitor {
	monitor := &Monitor{
		osExitSignal: make(chan os.Signal),
		// The master shutdown signal is sent and received from the same select block,
		// so it needs a buffer
		shutDownMaster: make(chan interface{}, 2),
		shutdownRest:   make(chan interface{}, 1),

		restErrList: make([]error, 0),

		restErrs: make(chan error, 1),

		restShutdownComplete: make(chan interface{}, 1),
		shutdownComplete:     new(sync.WaitGroup),
	}

	signal.Notify(monitor.osExitSignal, os.Interrupt, syscall.SIGTERM)
	monitor.shutdownComplete.Add(1)

	return monitor
}
