package servers

import "os"

func restAddress() string {
	host := os.Getenv("REST_HOST")
	// We use plain port here to get access to heroku's dynamically assigned port.
	port := os.Getenv("REST_PORT")
	address := host + ":" + port
	return address
}

func grpcAddress() string {
	host := os.Getenv("GRPC_HOST")
	port := os.Getenv("GRPC_PORT")
	address := host + ":" + port
	return address
}
