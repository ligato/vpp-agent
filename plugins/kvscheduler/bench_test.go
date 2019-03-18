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

package kvscheduler_test

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"testing"

	"github.com/ligato/cn-infra/logging"
	_ "github.com/ligato/cn-infra/logging/logrus" // for setting default registry

	mock_ifplugin "github.com/ligato/vpp-agent/examples/kvscheduler/mock_plugins/ifplugin"
	"github.com/ligato/vpp-agent/examples/kvscheduler/mock_plugins/ifplugin/model"
	mock_l2plugin "github.com/ligato/vpp-agent/examples/kvscheduler/mock_plugins/l2plugin"
	"github.com/ligato/vpp-agent/examples/kvscheduler/mock_plugins/l2plugin/model"
	"github.com/ligato/vpp-agent/pkg/models"
	. "github.com/ligato/vpp-agent/plugins/kvscheduler"
	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

/*
------------------------
 KVScheduler benchmarks
------------------------
- starts scheduler together with mocked ifplugin and l2plugin
- configures 1/10/100/1000 number of interfaces with bridge domain

How to run:
  - build test binary	`go test -c`
  - run all benchmarks:	`./kvscheduler.test -test.run=XXX -test.bench=.`
  - with CPU profile:	`./kvscheduler.test -test.run=XXX -test.bench=. -test.cpuprofile=cpu.out`
    - analyze profile: `go tool pprof cpu.out`
  - with mem profile:	`./kvscheduler.test -test.run=XXX -test.bench=. -memprofile mem.out`
    - analyze profile: `go tool pprof -alloc_space mem.out`
  - with trace profile:	`./kvscheduler.test -test.run=XXX -test.bench=. -trace trace.out`
    - analyze profile: `go tool trace -http=:6060 trace.out`

*/

func BenchmarkScale(b *testing.B) {
	benchmarkScale(b, true)
}

func BenchmarkScaleWithoutSimulation(b *testing.B) {
	benchmarkScale(b, false)
}


func benchmarkScale(b *testing.B, withSimulation bool) {
	for _, n := range [...]int{1, 10, 100, 1000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runScale(n, withSimulation)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// result should be saved to global variable to prevent compiler optimization
var seqNum uint64

func runScale(n int, withSimulation bool) error {
	c := setupScale()
	defer teardownScale(c)

	// run non-resync transaction against empty SB
	txn := c.scheduler.StartNBTransaction()

	// create single bridge domain
	valBd := &mock_l2.BridgeDomain{
		Name: fmt.Sprintf("bd-%d", 1),
	}
	// create n interfaces for bridge domain
	for i := 0; i < n; i++ {
		valIface := &mock_interfaces.Interface{
			Name: fmt.Sprintf("iface-%d", i),
			Type: mock_interfaces.Interface_LOOPBACK,
		}
		valBd.Interfaces = append(valBd.Interfaces, &mock_l2.BridgeDomain_Interface{
			Name: valIface.Name,
		})
		txn.SetValue(models.Key(valIface), valIface)
	}
	txn.SetValue(models.Key(valBd), valBd)

	testCtx := context.Background()
	if withSimulation {
		testCtx = WithSimulation(testCtx)
	}
	seq, err := txn.Commit(WithDescription(testCtx, "benchmarking scale"))
	if err != nil {
		return err
	}

	seqNum = seq

	return err
}

func setupScale() *runCtx {
	// prepare run context
	c := newRunCtx()
	if err := c.Init(); err != nil {
		panic(err)
	}
	return c
}

func teardownScale(c *runCtx) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}

type runCtx struct {
	scheduler *Scheduler
	IfPlugin  *mock_ifplugin.IfPlugin
	L2Plugin  *mock_l2plugin.L2Plugin
}

func newRunCtx() *runCtx {
	c := &runCtx{}
	c.scheduler = NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	c.IfPlugin = mock_ifplugin.NewPlugin(mock_ifplugin.UseDeps(
		func(deps *mock_ifplugin.Deps) {
			deps.KVScheduler = c.scheduler
		}))
	c.L2Plugin = mock_l2plugin.NewPlugin(mock_l2plugin.UseDeps(
		func(deps *mock_l2plugin.Deps) {
			deps.KVScheduler = c.scheduler
		}))
	return c
}

func (c *runCtx) Init() error {
	if err := c.scheduler.Init(); err != nil {
		return err
	}
	if err := c.IfPlugin.Init(); err != nil {
		return err
	}
	if err := c.L2Plugin.Init(); err != nil {
		return err
	}
	return nil
}

func (c *runCtx) Close() error {
	if err := c.L2Plugin.Close(); err != nil {
		return err
	}
	if err := c.IfPlugin.Close(); err != nil {
		return err
	}
	if err := c.scheduler.Close(); err != nil {
		return err
	}
	logging.DefaultRegistry.ClearRegistry()
	return nil
}

var scaleFlag = flag.Int("scale", 10, "number of items for scale test")

func TestScale(t *testing.T) {
	if err := runScale(*scaleFlag, true); err != nil {
		t.Fatal(err)
	}
}
