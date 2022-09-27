package e2e

import (
	"runtime"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
	"github.com/vishvananda/netns"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
)

const (
	msImage       = "busybox:1.31"
	msLabelKey    = "e2e.test.ms"
	MsNamePrefix  = "e2e-test-ms-"
	msStopTimeout = 1 // seconds
)

// Microservice represents running microservice
type Microservice struct {
	ComponentRuntime
	Pinger
	Diger

	ctx     *TestCtx
	name    string
	nsCalls nslinuxcalls.NetworkNamespaceAPI
}

// NewMicroservice creates and starts new microservice container
func NewMicroservice(
	ctx *TestCtx,
	msName string,
	nsCalls nslinuxcalls.NetworkNamespaceAPI,
	optMods ...MicroserviceOptModifier,
) (*Microservice, error) {
	// compute options
	opts := DefaultMicroserviceOpt(ctx, msName)
	for _, mod := range optMods {
		mod(opts)
	}

	// create struct for ETCD server
	ms := &Microservice{
		ComponentRuntime: opts.Runtime,
		ctx:              ctx,
		name:             msName,
		nsCalls:          nsCalls,
	}

	// Note: if runtime doesn't implement Pinger/Diger interface and test use it, then compilation
	// will be ok but runtime will throw "panic: runtime error: invalid memory address or nil pointer
	// dereference" when referencing Ping/Dig function
	if pinger, ok := opts.Runtime.(Pinger); ok {
		ms.Pinger = pinger
	}
	if diger, ok := opts.Runtime.(Diger); ok {
		ms.Diger = diger
	}

	// get runtime specific options and start microservice in runtime environment
	startOpts, err := opts.RuntimeStartOptions(ctx, opts)
	if err != nil {
		return nil, errors.Errorf("can't get microservice %s start option for runtime due to: %v", msName, err)
	}
	err = ms.Start(startOpts)
	if err != nil {
		return nil, errors.Errorf("can't start microservice %s due to: %v", msName, err)
	}
	return ms, nil
}

func (ms *Microservice) Stop(options ...interface{}) error {
	if err := ms.ComponentRuntime.Stop(options); err != nil {
		// not additionally cleaning up after attempting to stop test topology component because
		// it would lock access to further inspection of this component (i.e. why it won't stop)
		return err
	}
	// cleanup
	delete(ms.ctx.microservices, ms.name)
	return nil
}

// MicroserviceStartOptionsForContainerRuntime translates MicroserviceOpt to options for ComponentRuntime.Start(option)
// method implemented by ContainerRuntime
func MicroserviceStartOptionsForContainerRuntime(ctx *TestCtx, options interface{}) (interface{}, error) {
	opts, ok := options.(*MicroserviceOpt)
	if !ok {
		return nil, errors.Errorf("expected MicroserviceOpt but got %+v", options)
	}

	msLabel := MsNamePrefix + opts.Name
	createOpts := &docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    msLabel,
		Config: &docker.Config{
			Image: msImage,
			Labels: map[string]string{
				msLabelKey: opts.Name,
			},
			//Entrypoint:
			Env: []string{"MICROSERVICE_LABEL=" + msLabel},
			Cmd: []string{"tail", "-f", "/dev/null"},
		},
		HostConfig: &docker.HostConfig{
			// networking configured via VPP in E2E tests
			NetworkMode: "none",
		},
	}

	if opts.ContainerOptsHook != nil {
		opts.ContainerOptsHook(createOpts)
	}

	return &ContainerStartOptions{
		ContainerOptions: createOpts,
		Pull:             true,
	}, nil
}

// TODO this is runtime specific -> integrate it into runtime concept
func removeDanglingMicroservices(t *testing.T, dockerClient *docker.Client) {
	// remove any running microservices prior to starting a new test
	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": {msLabelKey},
		},
	})
	if err != nil {
		t.Fatalf("failed to list existing microservices: %v", err)
	}
	for _, container := range containers {
		err = dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			t.Fatalf("failed to remove existing microservices: %v", err)
		} else {
			t.Logf("removed existing microservice: %s", container.Labels[msLabelKey])
		}
	}
}

// TODO this is runtime specific -> integrate it into runtime concept
// enterNetNs enters the **network** namespace of the microservice (other namespaces
// remain unchanged). Leave using the returned callback.
func (ms *Microservice) enterNetNs() (exitNetNs func()) {
	origns, err := netns.Get()
	if err != nil {
		ms.ctx.t.Fatalf("failed to obtain current network namespace: %v", err)
	}
	nsHandle, err := ms.nsCalls.GetNamespaceFromPid(ms.PID())
	if err != nil {
		ms.ctx.t.Fatalf("failed to obtain handle for network namespace of microservice '%s': %v",
			ms.name, err)
	}
	defer nsHandle.Close()

	runtime.LockOSThread()
	err = ms.nsCalls.SetNamespace(nsHandle)
	if err != nil {
		ms.ctx.t.Fatalf("failed to enter network namespace of microservice '%s': %v",
			ms.name, err)
	}
	return func() {
		err = ms.nsCalls.SetNamespace(origns)
		if err != nil {
			ms.ctx.t.Fatalf("failed to return back to the original network namespace: %v", err)
		}
		origns.Close()
		runtime.UnlockOSThread()
	}
}
