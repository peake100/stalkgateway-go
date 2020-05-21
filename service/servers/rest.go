package servers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/peake100/stalkgateway-go/protogen/stalk_proto"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func RunRest(
	monitor *Monitor,
) {
	grpcServerEndpoint := grpcAddress()
	restAddress := restAddress()

	log.Printf("serving rest gateway on '%v'\n", restAddress)

	// Get a cancelable context for the server, and start a goroutine to listen for
	// a shutdown command and execute it.
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		<-monitor.shutdownRest
		log.Println("rest shutdown request received")
		cancelFunc()
		monitor.restShutdownComplete <- nil
	}()

	// Create a grpc dialer. Because all of our GRPC requests are getting routed through
	// Nginx, we can use the dame dialer for all of our endpoints.
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := stalkproto.RegisterStalkForecasterHandlerFromEndpoint(
		ctx, mux, grpcServerEndpoint, opts,
	)
	if err != nil {
		err = xerrors.Errorf(
			"error registering forecaster REST endpoints: %v", err,
		)
		monitor.restErrs <- err
	}

	err = stalkproto.RegisterStalkReporterHandlerFromEndpoint(
		ctx, mux, grpcServerEndpoint, opts,
	)
	if err != nil {
		err = xerrors.Errorf(
			"error registering reporter REST endpoints: %v", err,
		)
		monitor.restErrs <- err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoints)
	err = http.ListenAndServe(restAddress, mux)
	if err != nil {
		err = xerrors.Errorf("error serving REST endpoints: %v", err)
		monitor.restErrs <- err
	}

	log.Println("rest process exiting")
}
