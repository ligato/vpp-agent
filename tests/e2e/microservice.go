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
	msDefaultImage = "busybox:1.31"
	msStopTimeout  = 1 // seconds
	msLabelKey     = "e2e.test.ms"
	msNamePrefix   = "e2e-test-ms-"
)

var (
	linuxPingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)

type microservice struct {
	ctx *TestCtx

	name string

	dockerClient *docker.Client
	container    *docker.Container
	nsCalls      nslinuxcalls.NetworkNamespaceAPI
}

func createMicroservice(ctx *TestCtx, msName string, dockerClient *docker.Client, nsCalls nslinuxcalls.NetworkNamespaceAPI) (*microservice, error) {
	msLabel := msNamePrefix + msName

	opts := docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    msLabel,
		Config: &docker.Config{
			Image: msDefaultImage,
			Labels: map[string]string{
				msLabelKey: msName,
			},
			Env: []string{"MICROSERVICE_LABEL=" + msLabel},
			Cmd: []string{"sleep", "600"},
		},
		HostConfig: &docker.HostConfig{
			// networking configured via VPP in E2E tests
			NetworkMode: "none",
		},
	}

	container, err := dockerClient.CreateContainer(opts)
	if err != nil {
		return nil, fmt.Errorf("create container '%s' error: %w", msName, err)
	}

	err = dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("start container '%s' error: %w", msName, err)
	}
	container, err = dockerClient.InspectContainerWithOptions(docker.InspectContainerOptions{
		Context: ctx.ctx,
		ID:      container.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("inspect container '%s' error: %w", msName, err)
	}

	return &microservice{
		ctx:          ctx,
		name:         msName,
		container:    container,
		dockerClient: dockerClient,
		nsCalls:      nsCalls,
	}, nil
}

func resetMicroservices(t *testing.T, dockerClient *docker.Client) {
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

// execCmd allows to execute command **inside** the microservice - i.e. not just
// inside the network namespace of the microservice, but inside the container
// as a whole.
func (ms *microservice) execCmd(cmdName string, args ...string) (output string, err error) {
	execCtx, err := ms.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		Cmd:          append([]string{cmdName}, args...),
		Container:    ms.container.ID,
	})
	if err != nil {
		ms.ctx.t.Fatalf("failed to create docker exec instance: %v", err)
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

type pingOptions struct {
	allowedLoss int    // percentage of allowed loss for success
	outIface    string // outgoing interface name
	maxTimeout  int    // timeout in seconds before ping exits
	count       int    // number of pings
}

func newPingOpts(opts ...pingOpt) *pingOptions {
	popts := &pingOptions{
		allowedLoss: 49, // by default at least half of the packets should get through
		maxTimeout:  4,
	}
	popts.init(opts...)
	return popts
}

func (ping *pingOptions) init(opts ...pingOpt) {
	for _, o := range opts {
		o(ping)
	}
}

func (ping *pingOptions) args() []string {
	var args []string
	if ping.maxTimeout > 0 {
		args = append(args, "-w", fmt.Sprint(ping.maxTimeout))
	}
	if ping.count > 0 {
		args = append(args, "-c", fmt.Sprint(ping.count))
	}
	if ping.outIface != "" {
		args = append(args, "-I", ping.outIface)
	}
	return args
}

type pingOpt func(opts *pingOptions)

func pingWithAllowedLoss(maxLoss int) pingOpt {
	return func(opts *pingOptions) {
		opts.allowedLoss = maxLoss
	}
}

func pingWithOutInterface(iface string) pingOpt {
	return func(opts *pingOptions) {
		opts.outIface = iface
	}
}

// ping <destAddress> from inside of the microservice.
func (ms *microservice) ping(destAddress string, opts ...pingOpt) error {
	ms.ctx.t.Helper()

	ping := newPingOpts(opts...)
	args := append(ping.args(), destAddress)

	stdout, err := ms.execCmd("ping", args...)
	if err != nil {
		return err
	}

	matches := linuxPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	ms.ctx.logger.Printf("Linux ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss > ping.allowedLoss {
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
