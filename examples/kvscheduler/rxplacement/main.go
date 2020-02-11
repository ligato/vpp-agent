//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"log"
	"time"

	"go.ligato.io/cn-infra/v2/agent"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs_api "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

/*
	* VPP1 (configured by this example):
		- startup config:
			unix {
				interactive
				cli-listen 0.0.0.0:5002
				cli-no-pager
				coredump-size unlimited
				full-coredump
				poll-sleep-usec 50
			}
			api-trace {
				on
			}
			socksvr {
				socket-name "/tmp/vpp1.sock"
			}
			statseg {
				default
				per-node-counters on
			}
			cpu {
				workers 2
			}

	* VPP2 (configured manually to connect with myMemif):
		- startup config:
			unix {
				interactive
				cli-listen 0.0.0.0:5003
				cli-no-pager
				coredump-size unlimited
				full-coredump
				poll-sleep-usec 50
			}
			api-trace {
				on
			}
			socksvr {
				socket-name "/tmp/vpp2.sock"
			}
			statseg {
				default
				per-node-counters on
			}
			cpu {
				workers 2
			}

		- configuration to be applied via CLI:
			$ create interface memif id 0 slave rx-queues 5 tx-queues 5
			$ set int state memif0/0 up
			$ set int ip address memif0/0 192.168.1.2/24
*/
func main() {
	//vpp_ifplugin.DefaultPlugin.PublishStatistics = &Publisher{}
	ep := &ExamplePlugin{
		KVScheduler:  &kvs.DefaultPlugin,
		VPPIfPlugin:  &vpp_ifplugin.DefaultPlugin,
		Orchestrator: &orchestrator.DefaultPlugin,
	}

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	KVScheduler  *kvs.Scheduler
	VPPIfPlugin  *vpp_ifplugin.IfPlugin
	Orchestrator *orchestrator.Plugin
}

/*
type Publisher struct {
}

func (p *Publisher) Put(key string, data proto.Message, opts ...datasync.PutOption) error {
	fmt.Printf("Publishing key=%s, data=%+v\n", key, data)
	return nil
}
*/

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "rx-placement-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	ch := make(chan *kvscheduler.BaseValueStatus, 100)
	p.KVScheduler.WatchValueStatus(ch, nil)
	go watchValueStatus(ch)
	go testLocalClientWithScheduler(p.KVScheduler)
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

func watchValueStatus(ch <-chan *kvscheduler.BaseValueStatus) {
	for {
		select {
		case status := <-ch:
			fmt.Printf("Value status change: %v\n", status.String())
		}
	}
}

func testLocalClientWithScheduler(kvscheduler kvs_api.KVScheduler) {
	// initial resync
	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC WITH MEMIF ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		VppInterface(myMemif).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change #1
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE ===")

	myMemif.RxModes[0].Mode = vpp_interfaces.Interface_RxMode_INTERRUPT // change default
	myMemif.RxModes = append(myMemif.RxModes, &vpp_interfaces.Interface_RxMode{
		Queue: 3,
		Mode:  vpp_interfaces.Interface_RxMode_POLLING,
	})
	myMemif.RxPlacements = append(myMemif.RxPlacements, &vpp_interfaces.Interface_RxPlacement{
		Queue:      3,
		MainThread: true,
		Worker:     100, // ignored
	})

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.Put().
		VppInterface(myMemif).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change #2
	time.Sleep(time.Second * 20)
	fmt.Println("=== CHANGE ===")

	myMemif.GetMemif().RxQueues = 5
	myMemif.GetMemif().TxQueues = 5
	myMemif.RxPlacements = append(myMemif.RxPlacements, &vpp_interfaces.Interface_RxPlacement{
		Queue:      4,
		MainThread: true,
	})

	/* Re-create will fail - that is expected and it is due to the link-state key
	   being updated AFTER the transaction, not during. The subsequent retry/notification
	   should fix all the errors.
	*/

	txn3 := localclient.DataChangeRequest("example")
	err = txn3.Put().
		VppInterface(myMemif). // re-create
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}

var (
	myMemif = &vpp_interfaces.Interface{
		Name:        "my-memif",
		Type:        vpp_interfaces.Interface_MEMIF,
		Enabled:     true,
		IpAddresses: []string{"192.168.1.1/24"},

		RxPlacements: []*vpp_interfaces.Interface_RxPlacement{
			{
				Queue:  0,
				Worker: 0,
			},
			{
				Queue:      1,
				MainThread: true,
			},
			{
				Queue:  2,
				Worker: 1,
			},
		},

		RxModes: []*vpp_interfaces.Interface_RxMode{
			{
				DefaultMode: true,
				Mode:        vpp_interfaces.Interface_RxMode_POLLING,
			},
			{
				Queue: 1,
				Mode:  vpp_interfaces.Interface_RxMode_INTERRUPT,
			},
			{
				Queue: 2,
				Mode:  vpp_interfaces.Interface_RxMode_INTERRUPT,
			},
		},

		Link: &vpp_interfaces.Interface_Memif{
			Memif: &vpp_interfaces.MemifLink{
				Mode:     vpp_interfaces.MemifLink_ETHERNET,
				Master:   true,
				Id:       0,
				RxQueues: 4,
				TxQueues: 4,
			},
		},
	}
)
