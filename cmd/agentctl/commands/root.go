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

package commands

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/debug"
)

// NewRootNamed returns new Root named with name.
func NewRootNamed(name string, agentCli *cli.AgentCli) *Root {
	var (
		opts    *cli.ClientOptions
		flags   *pflag.FlagSet
		helpCmd *cobra.Command
	)
	cmd := &cobra.Command{
		Use:                   fmt.Sprintf("%s [options]", name),
		Short:                 "A CLI app for managing Ligato agents",
		SilenceUsage:          true,
		SilenceErrors:         true,
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return ShowHelp(agentCli.Err())(cmd, args)
			}
			return fmt.Errorf("%[1]s: '%[2]s' is not a %[1]s command.\nSee '%[1]s --help'", name, args[0])
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.Debugf("running command: %q", cmd.CommandPath())
			// TODO: isSupported?
			return nil
		},
		Version: fmt.Sprintf("%s, commit %s", agent.BuildVersion, agent.CommitHash),
	}

	opts, flags, helpCmd = SetupRootCommand(cmd)

	flags.BoolP("version", "v", false, "Print version info and quit")
	flags.BoolVarP(&opts.Debug, "debug", "D", false, "Enable debug mode")
	flags.StringVarP(&opts.LogLevel, "log-level", "l", "", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)

	cmd.SetHelpCommand(helpCmd)
	cmd.SetOutput(agentCli.Out())

	AddBaseCommands(cmd, agentCli)

	DisableFlagsInUseLine(cmd)

	return newRoot(cmd, agentCli, opts, flags)
}

// PrepareCommand handles global flags and Initialize should be
// called before executing returned cobra command.
func (root *Root) PrepareCommand() (*cobra.Command, error) {
	cmd, args, err := root.HandleGlobalFlags()
	if err != nil {
		return nil, fmt.Errorf("handle global flags failed: %v", err)
	}
	if debug.IsEnabledFor("flags") {
		fmt.Printf("flag.Args() = %v\n", args)
		cmd.DebugFlags()
	}
	cmd.SetArgs(args)
	return cmd, nil
}

// Root encapsulates a top-level cobra command (either agentctl or custom one).
type Root struct {
	cmd      *cobra.Command
	agentCli *cli.AgentCli
	opts     *cli.ClientOptions
	flags    *pflag.FlagSet
	args     []string
}

func newRoot(cmd *cobra.Command, agentCli *cli.AgentCli, opts *cli.ClientOptions, flags *pflag.FlagSet) *Root {
	return &Root{cmd, agentCli, opts, flags, os.Args[1:]}
}

// HandleGlobalFlags takes care of parsing global flags defined on the
// command, it returns the underlying cobra command and the args it
// will be called with (or an error).
//
// On success the caller is responsible for calling Initialize()
// before calling `Execute` on the returned command.
func (root *Root) HandleGlobalFlags() (*cobra.Command, []string, error) {
	cmd := root.cmd
	flags := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	flags.SetInterspersed(false)

	// We need the single parse to see both sets of flags.
	flags.AddFlagSet(cmd.Flags())
	flags.AddFlagSet(cmd.PersistentFlags())
	// Now parse the global flags, up to (but not including) the
	// first command. The result will be that all the remaining
	// arguments are in `flags.Args()`.
	if err := flags.Parse(root.args); err != nil {
		// Our FlagErrorFunc uses the cli, make sure it is initialized
		if err := root.Initialize(); err != nil {
			return nil, nil, err
		}
		return nil, nil, cmd.FlagErrorFunc()(cmd, err)
	}

	return cmd, flags.Args(), nil
}

// Initialize finalises global option parsing and initializes the agentctl client.
func (root *Root) Initialize(ops ...cli.InitializeOpt) error {
	return root.agentCli.Initialize(root.opts, ops...)
}

// SetupRootCommand setups cobra command and returns CLI client options, flags and help command.
func SetupRootCommand(rootCmd *cobra.Command) (*cli.ClientOptions, *pflag.FlagSet, *cobra.Command) {
	opts := cli.NewClientOptions()

	opts.InstallFlags(rootCmd.PersistentFlags())

	cobra.AddTemplateFunc("add", func(a, b int) int { return a + b })
	cobra.AddTemplateFunc("cmdExample", cmdExample)
	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
	cobra.AddTemplateFunc("wrappedGlobalFlagUsages", wrappedGlobalFlagUsages)

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetFlagErrorFunc(FlagErrorFunc)
	rootCmd.SetHelpCommand(helpCommand)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	_ = rootCmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")
	rootCmd.PersistentFlags().Lookup("help").Hidden = true

	return opts, rootCmd.Flags(), helpCommand
}

// FlagErrorFunc returns status error when err is not nil.
// It includes usage string and error message.
func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	var usage string
	if cmd.HasSubCommands() {
		usage = "\n\n" + cmd.UsageString()
	}
	return StatusError{
		Status:     fmt.Sprintf("%s\nSee '%s --help'.%s", err, cmd.CommandPath(), usage),
		StatusCode: 125,
	}
}

// ShowHelp shows the command help.
func ShowHelp(out io.Writer) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SetOutput(out)
		cmd.HelpFunc()(cmd, args)
		return nil
	}
}

// VisitAll will traverse all commands from the root.
// This is different from the VisitAll of cobra.Command where only parents
// are checked.
func VisitAll(root *cobra.Command, fn func(*cobra.Command)) {
	for _, cmd := range root.Commands() {
		VisitAll(cmd, fn)
	}
	fn(root)
}

// DisableFlagsInUseLine sets the DisableFlagsInUseLine flag on all
// commands within the tree rooted at cmd.
func DisableFlagsInUseLine(cmd *cobra.Command) {
	VisitAll(cmd, func(ccmd *cobra.Command) {
		// do not add a `[flags]` to the end of the usage line.
		ccmd.DisableFlagsInUseLine = true
	})
}

var helpCommand = &cobra.Command{
	Use:               "help [command]",
	Short:             "Help about the command",
	PersistentPreRun:  func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(c *cobra.Command, args []string) error {
		cmd, args, e := c.Root().Find(args)
		if cmd == nil || e != nil || len(args) > 0 {
			return errors.Errorf("unknown help topic: %v", strings.Join(args, " "))
		}
		helpFunc := cmd.HelpFunc()
		helpFunc(cmd, args)
		return nil
	},
}

var usageTemplate = `Usage:
{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasAvailableSubCommands}}	{{ .CommandPath}}{{- if .HasAvailableFlags}} [options]{{end}} COMMAND{{ end}}

{{if ne .Long ""}}{{ .Long  | trimRightSpace }}{{else }}{{ .Short | trimRightSpace }}

{{- end}}
{{- if gt .Aliases 0}}

ALIASES
  {{.NameAndAliases}}

{{- end}}
{{- if .HasExample}}

EXAMPLES
{{ cmdExample . | trimRightSpace}}

{{- end}}
{{- if .HasAvailableSubCommands }}

COMMANDS

{{- range .Commands }}{{- if .IsAvailableCommand}}
  {{rpad .Name (add .NamePadding 1)}}{{.Short}}
{{- end}}{{- end}}

{{- end}}
{{- if .HasLocalFlags}}

OPTIONS:
{{ wrappedFlagUsages . | trimRightSpace}}

{{- end}}
{{- if .HasInheritedFlags}}

GLOBALS:
{{ wrappedGlobalFlagUsages . | trimRightSpace}}

{{- end}}
{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{- if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

func cmdExample(cmd *cobra.Command) string {
	t := template.New("example")
	template.Must(t.Parse(cmd.Example))
	var b bytes.Buffer
	if err := t.Execute(&b, cmd); err != nil {
		panic(err)
	}
	return b.String()
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.LocalFlags().FlagUsagesWrapped(width - 1)
}

func wrappedGlobalFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.InheritedFlags().FlagUsagesWrapped(width - 1)
}
