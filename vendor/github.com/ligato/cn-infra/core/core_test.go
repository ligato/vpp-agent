// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/onsi/gomega"
)

func TestEmptyAgent(t *testing.T) {
	gomega.RegisterTestingT(t)

	agent := NewAgent(Inject(), WithTimeout(1*time.Second))
	gomega.Expect(agent).NotTo(gomega.BeNil())
	err := agent.Start()
	gomega.Expect(err).To(gomega.BeNil())
	err = agent.Stop()
	gomega.Expect(err).To(gomega.BeNil())
}

func TestEventLoopWithInterrupt(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugins := []*TestPlugin{{}, {}, {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeFalse())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeFalse())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	time.Sleep(100 * time.Millisecond)
	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeTrue())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeTrue())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}
	close(closeCh)

	select {
	case errCh := <-errCh:
		gomega.Expect(errCh).To(gomega.BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		gomega.Expect(p.Closed()).To(gomega.BeTrue())
	}
}

func TestEventLoopFailInit(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugins := []*TestPlugin{{}, {}, NewTestPlugin(true, false, false)}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeFalse())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeFalse())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	select {
	case errCh := <-errCh:
		gomega.Expect(errCh).NotTo(gomega.BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeTrue())
		// initialization failed of a plugin failed, afterInit was not called
		gomega.Expect(p.AfterInitialized()).To(gomega.BeFalse())
		gomega.Expect(p.Closed()).To(gomega.BeTrue())
	}
	close(closeCh)

}

func TestEventLoopAfterInitFailed(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugins := []*TestPlugin{{}, NewTestPlugin(false, true, false), {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeFalse())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeFalse())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	select {
	case errCh := <-errCh:
		gomega.Expect(errCh).NotTo(gomega.BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeTrue())
		gomega.Expect(p.Closed()).To(gomega.BeTrue())
	}
	close(closeCh)

	gomega.Expect(plugins[0].AfterInitialized()).To(gomega.BeTrue())
	gomega.Expect(plugins[1].AfterInitialized()).To(gomega.BeTrue())
	// afterInit of the second plugin failed thus the third was not afterInitialized
	gomega.Expect(plugins[2].AfterInitialized()).To(gomega.BeFalse())

}

func TestEventLoopCloseFailed(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugins := []*TestPlugin{NewTestPlugin(false, false, true), {}, {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeFalse())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeFalse())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	time.Sleep(100 * time.Millisecond)
	for _, p := range plugins {
		gomega.Expect(p.Initialized()).To(gomega.BeTrue())
		gomega.Expect(p.AfterInitialized()).To(gomega.BeTrue())
		gomega.Expect(p.Closed()).To(gomega.BeFalse())
	}

	close(closeCh)

	select {
	case errCh := <-errCh:
		gomega.Expect(errCh).NotTo(gomega.BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		gomega.Expect(p.Closed()).To(gomega.BeTrue())
	}

}

func TestPluginApi(t *testing.T) {
	gomega.RegisterTestingT(t)
	const plName = "Name"
	named := NamedPlugin{plName, &TestPlugin{}}

	strRep := named.String()
	gomega.Expect(strRep).To(gomega.BeEquivalentTo(plName))
}

type TestPlugin struct {
	failInit      bool
	failAfterInit bool
	failClose     bool

	sync.Mutex
	initCalled      bool
	afterInitCalled bool
	closeCalled     bool
}

func NewTestPlugin(failInit, failAfterInit, failClose bool) *TestPlugin {
	return &TestPlugin{failInit: failInit, failAfterInit: failAfterInit, failClose: failClose}
}

func (p *TestPlugin) Init() error {
	p.Lock()
	defer p.Unlock()
	p.initCalled = true
	if p.failInit {
		return fmt.Errorf("Init failed")
	}
	return nil
}
func (p *TestPlugin) AfterInit() error {
	p.Lock()
	defer p.Unlock()
	p.afterInitCalled = true
	if p.failAfterInit {
		return fmt.Errorf("AfterInit failed")
	}
	return nil
}
func (p *TestPlugin) Close() error {
	p.Lock()
	defer p.Unlock()
	p.closeCalled = true
	if p.failClose {
		return fmt.Errorf("Close failed")
	}
	return nil
}

func (p *TestPlugin) Initialized() bool {
	p.Lock()
	defer p.Unlock()
	return p.initCalled
}

func (p *TestPlugin) AfterInitialized() bool {
	p.Lock()
	defer p.Unlock()
	return p.afterInitCalled
}

func (p *TestPlugin) Closed() bool {
	p.Lock()
	defer p.Unlock()
	return p.closeCalled
}
