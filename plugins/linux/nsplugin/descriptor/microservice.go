// Copyright (c) 2018 Cisco and/or its affiliates.
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

package descriptor

import (
	"context"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	prototypes "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"

	nsmodel "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
)

const (
	// MicroserviceDescriptorName is the name of the descriptor for microservices.
	MicroserviceDescriptorName = "microservice"

	// docker API keywords
	dockerTypeContainer = "container"
	dockerStateRunning  = "running"
	dockerActionStart   = "start"
	dockerActionStop    = "stop"
)

// MicroserviceDescriptor watches Docker and notifies KVScheduler about newly
// started and stopped microservices.
type MicroserviceDescriptor struct {
	// input arguments
	log         logging.Logger
	kvscheduler kvs.KVScheduler

	// map microservice label -> time of the last creation
	createTime map[string]time.Time

	// lock used to serialize access to microservice state data
	msStateLock sync.Mutex

	// conditional variable to check if microservice state data are in-sync
	// with the docker
	msStateInSync     bool
	msStateInSyncCond *sync.Cond

	// docker client - used to convert microservice label into the PID and
	// ID of the container
	dockerClient *docker.Client
	// microservice label -> microservice state data
	microServiceByLabel map[string]*Microservice
	// microservice container ID -> microservice state data
	microServiceByID map[string]*Microservice

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Microservice is used to store PID and ID of the container running a given
// microservice.
type Microservice struct {
	Label string
	PID   int
	ID    string
}

// microserviceCtx contains all data required to handle microservice changes.
type microserviceCtx struct {
	created       []string
	since         string
	lastInspected int64
}

// NewMicroserviceDescriptor creates a new instance of the descriptor for microservices.
func NewMicroserviceDescriptor(kvscheduler kvs.KVScheduler, log logging.PluginLogger) (*MicroserviceDescriptor, error) {
	var err error

	descriptor := &MicroserviceDescriptor{
		log:                 log.NewLogger("ms-descriptor"),
		kvscheduler:         kvscheduler,
		createTime:          make(map[string]time.Time),
		microServiceByLabel: make(map[string]*Microservice),
		microServiceByID:    make(map[string]*Microservice),
	}
	descriptor.msStateInSyncCond = sync.NewCond(&descriptor.msStateLock)
	descriptor.ctx, descriptor.cancel = context.WithCancel(context.Background())

	// Docker client
	descriptor.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return nil, errors.Errorf("failed to get docker client instance from the environment variables: %v", err)
	}
	log.Debugf("Using docker client endpoint: %+v", descriptor.dockerClient.Endpoint())

	return descriptor, nil
}

// GetDescriptor returns descriptor suitable for registration with the KVScheduler.
func (d *MicroserviceDescriptor) GetDescriptor() *kvs.KVDescriptor {
	return &kvs.KVDescriptor{
		Name:        MicroserviceDescriptorName,
		KeySelector: d.IsMicroserviceKey,
		Retrieve:    d.Retrieve,
	}
}

// IsMicroserviceKey returns true for key identifying microservices.
func (d *MicroserviceDescriptor) IsMicroserviceKey(key string) bool {
	return strings.HasPrefix(key, nsmodel.MicroserviceKeyPrefix)
}

// Retrieve returns key with empty value for every currently existing microservice.
func (d *MicroserviceDescriptor) Retrieve(correlate []kvs.KVWithMetadata) (values []kvs.KVWithMetadata, err error) {
	// wait until microservice state data are in-sync with the docker
	d.msStateLock.Lock()
	if !d.msStateInSync {
		d.msStateInSyncCond.Wait()
	}
	defer d.msStateLock.Unlock()

	for msLabel := range d.microServiceByLabel {
		values = append(values, kvs.KVWithMetadata{
			Key:    nsmodel.MicroserviceKey(msLabel),
			Value:  &prototypes.Empty{},
			Origin: kvs.FromSB,
		})
	}

	return values, nil
}

// StartTracker starts microservice tracker,
func (d *MicroserviceDescriptor) StartTracker() {
	go d.trackMicroservices(d.ctx)
}

// StopTracker stops microservice tracker,
func (d *MicroserviceDescriptor) StopTracker() {
	d.cancel()
	d.wg.Wait()
}

// GetMicroserviceStateData returns state data for the given microservice.
func (d *MicroserviceDescriptor) GetMicroserviceStateData(msLabel string) (ms *Microservice, found bool) {
	d.msStateLock.Lock()
	if !d.msStateInSync {
		d.msStateInSyncCond.Wait()
	}
	defer d.msStateLock.Unlock()

	ms, found = d.microServiceByLabel[msLabel]
	return ms, found
}

// detectMicroservice inspects container to see if it is a microservice.
// If microservice is detected, processNewMicroservice() is called to process it.
func (d *MicroserviceDescriptor) detectMicroservice(container *docker.Container) {
	// Search for the microservice label.
	var label string
	for _, env := range container.Config.Env {
		if strings.HasPrefix(env, servicelabel.MicroserviceLabelEnvVar+"=") {
			label = env[len(servicelabel.MicroserviceLabelEnvVar)+1:]
			if label != "" {
				d.log.Debugf("detected container as microservice: Name=%v ID=%v Created=%v State.StartedAt=%v", container.Name, container.ID, container.Created, container.State.StartedAt)
				last := d.createTime[label]
				if last.After(container.Created) {
					d.log.Debugf("ignoring older container created at %v as microservice: %+v", last, container)
					continue
				}
				d.createTime[label] = container.Created
				d.processNewMicroservice(label, container.ID, container.State.Pid)
			}
		}
	}
}

// processNewMicroservice is triggered every time a new microservice gets freshly started. All pending interfaces are moved
// to its namespace.
func (d *MicroserviceDescriptor) processNewMicroservice(microserviceLabel string, id string, pid int) {
	d.msStateLock.Lock()
	defer d.msStateLock.Unlock()

	ms, restarted := d.microServiceByLabel[microserviceLabel]
	if restarted {
		d.processTerminatedMicroservice(ms.ID)
		d.log.WithFields(logging.Fields{"label": microserviceLabel, "new-pid": pid, "new-id": id}).
			Warn("Microservice has been restarted")
	} else {
		d.log.WithFields(logging.Fields{"label": microserviceLabel, "pid": pid, "id": id}).
			Debug("Discovered new microservice")
	}

	ms = &Microservice{Label: microserviceLabel, PID: pid, ID: id}
	d.microServiceByLabel[microserviceLabel] = ms
	d.microServiceByID[id] = ms

	// Notify scheduler about new microservice
	if d.msStateInSync {
		d.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
			Key:      nsmodel.MicroserviceKey(ms.Label),
			Value:    &prototypes.Empty{},
			Metadata: nil,
		})
	}
}

// processTerminatedMicroservice is triggered every time a known microservice
// has terminated. All associated interfaces become obsolete and are thus removed.
func (d *MicroserviceDescriptor) processTerminatedMicroservice(id string) {
	ms, exists := d.microServiceByID[id]
	if !exists {
		d.log.WithFields(logging.Fields{"id": id}).
			Warn("Detected removal of an unknown microservice")
		return
	}
	d.log.WithFields(logging.Fields{"label": ms.Label, "pid": ms.PID, "id": ms.ID}).
		Debug("Microservice has terminated")

	delete(d.microServiceByLabel, ms.Label)
	delete(d.microServiceByID, ms.ID)

	// Notify scheduler about terminated microservice
	if d.msStateInSync {
		d.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
			Key:      nsmodel.MicroserviceKey(ms.Label),
			Value:    nil,
			Metadata: nil,
		})
	}
}

// setStateInSync sets internal state to "in sync" and signals the state transition.
func (d *MicroserviceDescriptor) setStateInSync() {
	d.msStateLock.Lock()
	d.msStateInSync = true
	d.msStateLock.Unlock()
	d.msStateInSyncCond.Broadcast()
}

// processStartedContainer processes a started Docker container - inspects whether it is a microservice.
// If it is, notifies scheduler about a new microservice.
func (d *MicroserviceDescriptor) processStartedContainer(id string) {
	container, err := d.dockerClient.InspectContainer(id)
	if err != nil {
		d.log.Warnf("Error by inspecting container %s: %v", id, err)
		return
	}
	d.detectMicroservice(container)
}

// processStoppedContainer processes a stopped Docker container - if it is a microservice,
// notifies scheduler about its termination.
func (d *MicroserviceDescriptor) processStoppedContainer(id string) {
	d.msStateLock.Lock()
	defer d.msStateLock.Unlock()

	if _, found := d.microServiceByID[id]; found {
		d.processTerminatedMicroservice(id)
	}
}

// trackMicroservices is running in the background and maintains a map of microservice labels to container info.
func (d *MicroserviceDescriptor) trackMicroservices(ctx context.Context) {
	d.wg.Add(1)
	defer func() {
		d.wg.Done()
		d.log.Debugf("Microservice tracking ended")
	}()

	// subscribe to Docker events
	listener := make(chan *docker.APIEvents, 10)
	err := d.dockerClient.AddEventListener(listener)
	if err != nil {
		d.log.Warnf("Failed to add Docker event listener: %v", err)
		d.setStateInSync() // empty set of microservices is considered
		return
	}

	// list currently running containers
	listOpts := docker.ListContainersOptions{
		All: true,
	}
	containers, err := d.dockerClient.ListContainers(listOpts)
	if err != nil {
		d.log.Warnf("Failed to list Docker containers: %v", err)
		d.setStateInSync() // empty set of microservices is considered
		return
	}
	for _, container := range containers {
		if container.State == dockerStateRunning {
			details, err := d.dockerClient.InspectContainer(container.ID)
			if err != nil {
				d.log.Warnf("Error by inspecting container %s: %v", container.ID, err)
				continue
			}
			d.detectMicroservice(details)
		}
	}

	// mark state data as in-sync
	d.setStateInSync()

	// process Docker events
	for {
		select {
		case ev, ok := <-listener:
			if !ok {
				return
			}
			if ev.Type == dockerTypeContainer {
				if ev.Action == dockerActionStart {
					d.processStartedContainer(ev.Actor.ID)
				}
				if ev.Action == dockerActionStop {
					d.processStoppedContainer(ev.Actor.ID)
				}
			}
		case <-d.ctx.Done():
			return
		}
	}
}
