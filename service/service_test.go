package service

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"github.com/peake100/stalkgateway-go/service/servers"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
)

type TestBasicPredictionSuite struct {
	suite.Suite
	monitor    *servers.Monitor

	grpcAddress string
}

func (suite *TestBasicPredictionSuite) SetupSuite() {
	envVars := map[string]string{
		"GRPC_HOST": "localhost",
		"GRPC_PORT": "50051",
		"REST_HOST": "localhost",
		"PORT":      "8080",
	}
	for key, value := range envVars {
		err := os.Setenv(key, value)
		if err != nil {
			suite.FailNow("error setting env vars", err)
		}
	}

	log.Println("starting servers")
	suite.monitor = StartServers()

	log.Println("servers started")
}

func (suite *TestBasicPredictionSuite) TearDownSuite() {
	defer suite.monitor.WaitOnShutdown()
	defer suite.monitor.ShutdownServers()
}

func (suite *TestBasicPredictionSuite) TestDummy() {

}

func TestBasicPrediction(t *testing.T) {
	suite.Run(t, new(TestBasicPredictionSuite))
}
