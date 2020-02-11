package commands

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

var (
	// RootName defines default name used for the root command.
	RootName = "agentctl"
)

// NewAgentCli creates new AgentCli with opts and configures log output to error stream.
func NewAgentCli(opts ...cli.AgentCliOption) *cli.AgentCli {
	agentCli, err := cli.NewAgentCli(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logrus.SetOutput(agentCli.Err())
	logging.DefaultLogger.SetOutput(agentCli.Err())
	return agentCli
}

// NewRootCommand is helper for default initialization process for root command.
// Returs cobra command which is ready to be executed.
func NewRootCommand(agentCli *cli.AgentCli) (*cobra.Command, error) {
	root := NewRoot(agentCli)
	cmd, err := root.PrepareCommand()
	if err != nil {
		return nil, err
	}
	if err := root.Initialize(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// NewRoot returns new Root using RootName for name.
func NewRoot(agentCli *cli.AgentCli) *Root {
	return NewRootNamed(RootName, agentCli)
}

// AddBaseCommands adds all base commands to cmd.
func AddBaseCommands(cmd *cobra.Command, cli cli.Cli) {
	cmd.AddCommand(
		NewModelCommand(cli),
		NewLogCommand(cli),
		NewImportCommand(cli),
		NewVppCommand(cli),
		NewDumpCommand(cli),
		NewKvdbCommand(cli),
		NewGenerateCommand(cli),
		NewStatusCommand(cli),
		NewValuesCommand(cli),
		NewServiceCommand(cli),
		NewMetricsCommand(cli),
	)
}
