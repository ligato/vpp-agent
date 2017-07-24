package main

import (
	"os"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	vpp_flavor "github.com/ligato/vpp-agent/flavours/vpp"
)

// runs statically linked binary of Agent Core Plugins (called "vpp_flavor") with ETCD & Kafka connectors
func main() {

	f := vpp_flavor.Flavour{}
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, f.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logging.DebugLevel)
}
