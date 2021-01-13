package e2e

import (
	"fmt"
	"regexp"
	"runtime"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/vishvananda/netns"

	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
)

const (
	msDefaultImage = "busybox:1.31"
	msStopTimeout  = 1 // seconds
	msLabelKey     = "e2e.test.ms"
	msNamePrefix   = "e2e-test-ms-"
)

var (
	linuxPingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)

type microservice struct {
	*Container

	name    string
	nsCalls nslinuxcalls.NetworkNamespaceAPI
}

func createMicroservice(ctx *TestCtx, msName string, dockerClient *docker.Client, nsCalls nslinuxcalls.NetworkNamespaceAPI) (*microservice, error) {
	ms := &microservice{
		Container: &Container{
			ctx:         ctx,
			logIdentity: "Microservice " + msName,
			stopTimeout: msStopTimeout,
		},
		name:    msName,
		nsCalls: nsCalls,
	}

	msLabel := msNamePrefix + msName
	opts := &docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    msLabel,
		Config: &docker.Config{
			Image: msDefaultImage,
			Labels: map[string]string{
				msLabelKey: msName,
			},
			Env: []string{"MICROSERVICE_LABEL=" + msLabel},
			Cmd: []string{"tail", "-f", "/dev/null"},
		},
		HostConfig: &docker.HostConfig{
			// networking configured via VPP in E2E tests
			NetworkMode: "none",
		},
	}

	_, err := ms.create(opts, true)
	if err != nil {
		return nil, fmt.Errorf("create container '%s' error: %w", msName, err)
	}

	err = ms.start()
	if err != nil {
		return nil, fmt.Errorf("start container '%s' error: %w", msName, err)
	}

	return ms, nil
}

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

// enterNetNs enters the **network** namespace of the microservice (other namespaces
// remain unchanged). Leave using the returned callback.
func (ms *microservice) enterNetNs() (exitNetNs func()) {
	origns, err := netns.Get()
	if err != nil {
		ms.ctx.t.Fatalf("failed to obtain current network namespace: %v", err)
	}
	nsHandle, err := ms.nsCalls.GetNamespaceFromPid(ms.container.State.Pid)
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
