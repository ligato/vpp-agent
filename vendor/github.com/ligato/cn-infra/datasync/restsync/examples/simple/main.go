package main

import (
	"os"
	"os/signal"
	"time"

	agent "github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logroot"
	"gitlab.cisco.com/ctao/vnf-agent/agent/flavors/generic"
	vpp_flavor "gitlab.cisco.com/ctao/vnf-agent/agent/flavors/vpp"
)

// runs HTTP/REST endpoint for standard VPP plugins (without ETCD connectivity)
func main() {
	agent := agent.NewAgent(log.StandardLogger(), 15*time.Second, append(
		generic.HTTPAndLogPlugins(), //start without ETCD & Kafka
		vpp_flavor.DefaultVppPlugins()...)...)

	err := agent.Start()
	if err != nil {
		log.Errorf("Agent start failed, error '%+v'", err)
		os.Exit(1)
	}
	defer func() {
		err := agent.Stop()
		if err != nil {
			log.Errorf("Agent stop error '%+v'", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
		log.Println("Interrupt received, returning.")
		return
	}
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logging.DebugLevel)
}
