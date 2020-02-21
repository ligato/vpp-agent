package e2e

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/vishvananda/netns"

	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
)

const (
	msImage       = "busybox"
	msImageTag    = "1.31"
	msStopTimeout = 3 // seconds
	msLabelKey    = "e2e.test.ms"
	msNamePrefix  = "e2e-test-"
)

var (
	linuxPingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)

type microservice struct {
	t            *testing.T
	name         string
	dockerClient *docker.Client
	container    *docker.Container
	nsCalls      nslinuxcalls.NetworkNamespaceAPI
}

func createMicroservice(t *testing.T, msName string, dockerClient *docker.Client, nsCalls nslinuxcalls.NetworkNamespaceAPI) *microservice {
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: msNamePrefix + msName,
		Config: &docker.Config{
			Env:    []string{"MICROSERVICE_LABEL=" + msNamePrefix + msName},
			Image:  msImage + ":" + msImageTag,
			Cmd:    []string{"tail", "-f", "/dev/null"},
			Labels: map[string]string{msLabelKey: msName},
		},
		HostConfig: &docker.HostConfig{
			// networking configured via VPP in E2E tests
			NetworkMode: "none",
		},
	})
	if err != nil {
		t.Fatalf("failed to create microservice '%s': %v", msName, err)
	}
	err = dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		t.Fatalf("failed to start microservice '%s': %v", msName, err)
	}
	container, err = dockerClient.InspectContainer(container.ID)
	if err != nil {
		t.Fatalf("failed to inspect microservice '%s': %v", msName, err)
	}
	return &microservice{
		t:            t,
		name:         msName,
		container:    container,
		dockerClient: dockerClient,
		nsCalls:      nsCalls,
	}
}

func resetMicroservices(t *testing.T, dockerClient *docker.Client) {
	// pull image for microservices
	err := dockerClient.PullImage(docker.PullImageOptions{
		Repository: msImage,
		Tag:        msImageTag,
	}, docker.AuthConfiguration{})
	if err != nil {
		t.Fatalf("failed to pull image '%s:%s' for microservices: %v", msImage, msImageTag, err)
	}

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

func (ms *microservice) stop() error {
	err := ms.dockerClient.StopContainer(ms.container.ID, msStopTimeout)
	if err != nil {
		return err
	}
	return ms.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ms.container.ID,
		Force: true,
	})
}

// exec allows to execute command **inside** the microservice - i.e. not just
// inside the network namespace of the microservice, but inside the container
// as a whole.
func (ms *microservice) exec(cmdName string, args ...string) (output string, err error) {
	execCtx, err := ms.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		Cmd:          append([]string{cmdName}, args...),
		Container:    ms.container.ID,
	})
	if err != nil {
		ms.t.Fatalf("failed to create docker exec instance for ping: %v", err)
	}

	var stdout bytes.Buffer
	err = ms.dockerClient.StartExec(execCtx.ID, docker.StartExecOptions{
		OutputStream: &stdout,
	})
	return stdout.String(), err
}

// enterNetNs enters the **network** namespace of the microservice (other namespaces
// remain unchanged). Leave using the returned callback.
func (ms *microservice) enterNetNs() (exitNetNs func()) {
	origns, err := netns.Get()
	if err != nil {
		ms.t.Fatalf("failed to obtain current network namespace: %v", err)
	}
	nsHandle, err := ms.nsCalls.GetNamespaceFromPid(ms.container.State.Pid)
	if err != nil {
		ms.t.Fatalf("failed to obtain handle for network namespace of microservice '%s': %v",
			ms.name, err)
	}
	defer nsHandle.Close()

	runtime.LockOSThread()
	err = ms.nsCalls.SetNamespace(nsHandle)
	if err != nil {
		ms.t.Fatalf("failed to enter network namespace of microservice '%s': %v",
			ms.name, err)
	}
	return func() {
		err = ms.nsCalls.SetNamespace(origns)
		if err != nil {
			ms.t.Fatalf("failed to return back to the original network namespace: %v", err)
		}
		origns.Close()
		runtime.UnlockOSThread()
	}
}

// ping <destAddress> from inside of the microservice.
func (ms *microservice) ping(destAddress string, allowedLoss ...int) error {
	ms.t.Helper()

	stdout, err := ms.exec("ping", "-w", "4", destAddress)
	if err != nil {
		return err
	}

	matches := linuxPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	ms.t.Logf("Linux ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	maxLoss := 49 // by default at least half of the packets should ge through
	if len(allowedLoss) > 0 {
		maxLoss = allowedLoss[0]
	}
	if sent == 0 || loss > maxLoss {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

func parsePingOutput(output string, matches []string) (sent int, recv int, loss int, err error) {
	if len(matches) != 4 {
		err = fmt.Errorf("unexpected output from ping: %s", output)
		return
	}
	sent, err = strconv.Atoi(matches[1])
	if err != nil {
		err = fmt.Errorf("failed to parse the sent packet count: %v", err)
		return
	}
	recv, err = strconv.Atoi(matches[2])
	if err != nil {
		err = fmt.Errorf("failed to parse the received packet count: %v", err)
		return
	}
	loss, err = strconv.Atoi(matches[3])
	if err != nil {
		err = fmt.Errorf("failed to parse the loss percentage: %v", err)
		return
	}
	return
}
