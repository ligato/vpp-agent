// Copyright (c) 2020 Pantheon.tech
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
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	govppapi "go.fd.io/govpp/api"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/version"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

const failedReportFileName = "_failed-reports.txt"

func NewReportCommand(cli agentcli.Cli) *cobra.Command {
	var opts ReportOptions
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Create error report",
		Long: "Create report about running software stack (VPP-Agent, VPP, AgentCtl,...) " +
			"to allow quicker resolving of problems. The report will be a zip file containing " +
			"information grouped in multiple files",
		Example: `
# Default reporting (creates report file in current directory, whole reporting fails on subreport error)
{{.CommandPath}} report

# Reporting into custom directory ("/tmp")
{{.CommandPath}} report -o /tmp

# Reporting and ignoring errors from subreports (writing successful reports and 
# errors from failed subreports into zip report file)
{{.CommandPath}} report -i
`, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.OutputDirectory, "output-directory", "o", "",
		"Output directory (as absolute path) where report zip file will be written. "+
			"Default is current directory.")
	flags.BoolVarP(&opts.IgnoreErrors, "ignore-errors", "i", false,
		"Ignore subreport errors and create report zip file with all successfully retrieved/processed "+
			"information (the errors will be part of the report too)")
	return cmd
}

type ReportOptions struct {
	OutputDirectory string
	IgnoreErrors    bool
}

func runReport(cli agentcli.Cli, opts ReportOptions) error {
	// create report time and dependent variables
	reportTime := time.Now()
	reportName := fmt.Sprintf("agentctl-report--%s",
		strings.ReplaceAll(reportTime.UTC().Format("2006-01-02--15-04-05-.000"), ".", ""))

	// create temporal directory
	dirNamePattern := fmt.Sprintf("%v--*", reportName)
	dirName, err := ioutil.TempDir("", dirNamePattern)
	if err != nil {
		return fmt.Errorf("can't create tmp directory with name pattern %s due to %v", dirNamePattern, err)
	}
	defer os.RemoveAll(dirName)

	// create report files
	errors := packErrors(
		writeReportTo("_report.txt", dirName, writeMainReport, cli, reportTime),
		writeReportTo("software-versions.txt", dirName, writeAgentctlVersionReport, cli),
		writeReportTo("software-versions.txt", dirName, writeAgentVersionReport, cli),
		writeReportTo("software-versions.txt", dirName, writeVPPVersionReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareCPUReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareNumaReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPMainMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPNumaHeapMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPPhysMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPAPIMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPStatsMemoryReport, cli),
		writeReportTo("agent-status.txt", dirName, writeAgentStatusReport, cli),
		writeReportTo("agent-transaction-history.txt", dirName, writeAgentTxnHistoryReport, cli),
		writeReportTo("agent-NB-config.yaml", dirName, writeAgentNBConfigReport, cli),
		writeReportTo("agent-kvscheduler-NB-config-view.txt", dirName, writeKVschedulerNBConfigReport, cli),
		writeReportTo("agent-kvscheduler-SB-config-view.txt", dirName, writeKVschedulerSBConfigReport, cli),
		writeReportTo("agent-kvscheduler-cached-config-view.txt", dirName, writeKVschedulerCachedConfigReport, cli),
		writeReportTo("vpp-startup-config.txt", dirName, writeVPPStartupConfigReport, cli),
		writeReportTo("vpp-running-config(vpp-agent-SB-dump).yaml", dirName, writeVPPRunningConfigReport, cli),
		writeReportTo("vpp-event-log.txt", dirName, writeVPPEventLogReport, cli),
		writeReportTo("vpp-log.txt", dirName, writeVPPLogReport, cli),
		writeReportTo("vpp-statistics-interfaces.txt", dirName, writeVPPInterfaceStatsReport, cli),
		writeReportTo("vpp-statistics-errors.txt", dirName, writeVPPErrorStatsReport, cli),
		writeReportTo("vpp-api-trace.txt", dirName, writeVPPApiTraceReport, cli),
		writeReportTo("vpp-other-srv6.txt", dirName, writeVPPSRv6LocalsidReport, cli),
		writeReportTo("vpp-other-srv6.txt", dirName, writeVPPSRv6PolicyReport, cli),
		writeReportTo("vpp-other-srv6.txt", dirName, writeVPPSRv6SteeringReport, cli),
	)
	// summary errors from reports (actual errors are already written in reports,
	// user console and failedReportFileName file)
	if len(errors) > 0 {
		if !opts.IgnoreErrors {
			cli.Out().Write([]byte(fmt.Sprintf("%d subreport(s) failed.\n\nIf you want to ignore errors "+
				"from subreports and create report from the successfully retrieved/processed information then "+
				"add the --ignore-errors (-i) argument to the command (i.e. 'agentctl report -i')", len(errors))))
			return errors
		}
		cli.Out().Write([]byte(fmt.Sprintf("%d subreport(s) couldn't be fully or partially created "+
			"(full list with errors will be in packed zip file as file %s)\n\n", len(errors), failedReportFileName)))
	} else { //
		cli.Out().Write([]byte("All subreports were successfully created...\n\n"))
		// remove empty "failed report" file (ignoring remove failure because it means only one more
		// empty file in report zip file)
		os.Remove(filepath.Join(dirName, failedReportFileName))
	}

	// resolve zip file name
	simpleZipFileName := reportName + ".zip"
	zipFileName := filepath.Join(opts.OutputDirectory, simpleZipFileName)
	if opts.OutputDirectory == "" {
		zipFileName, err = filepath.Abs(simpleZipFileName)
		if err != nil {
			return fmt.Errorf("can't find out absolute path for output zip file due to: %v\n\n", err)
		}
	}

	// combine report files into one zip file
	cli.Out().Write([]byte("Creating report zip file... "))
	if err := createZipFile(zipFileName, dirName); err != nil {
		return fmt.Errorf("can't create zip file(%v) due to: %v", zipFileName, err)
	}
	cli.Out().Write([]byte(fmt.Sprintf("Done.\nReport file: %v\n", zipFileName)))

	return nil
}

func writeMainReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	// using template also for simple cases to be consistent with variable formatting (i.e. time formatting)
	format := `################# REPORT #################
report creation time:    {{epoch .ReportTime}}
report version:          {{.Version}} (=AGENTCTL version)

Subreport/file structure of this report:
    _report.txt (this file)
        Contains primary identification for the whole report (creation time, report version)
    _failed-reports.txt
        Contains all errors from all subreports. The errors are presented for user convenience at 3 places:
            1. showed to user in console while running the reporting command of agentctl
            2. in failed report file (_failed-reports.txt) - all errors at one place
            3. in subreport file - error in place where the retrieved information should be
    software-versions.txt
        Contains identification of the software stack used (agentctl/vpp-agent/vpp)
    hardware.txt
        Contains some information about the hardware that the software(vpp-agent/vpp) is running on
    agent-status.txt
        Contains status of vpp-agent and its plugins. Contains also boot time and up time of vpp-agent.
    agent-transaction-history.txt
        Contains transaction history of vpp-agent.
    agent-NB-config.yaml
        Contains vpp-agent's northbound(desired) configuration for VPP. The output is in compatible format 
        (yaml with correct structure) with agentctl configuration import fuctionality ('agentctl config update')
        so it can be used to setup the configuration from report target installation in local environment for 
        debugging or other purposes.
        This version doesn't contains data for custom 3rd party configuration models, but only data for 
        vpp-agent upstreamed configuration models. For full configuration data (but not in import compatible format)
        see agent-kvscheduler-NB-config-view.txt.
    agent-kvscheduler-NB-config-view.txt
        Contains vpp-agent's northbound(desired) configuration for VPP as seen by vpp-agent's KVScheduler component.
        The KVScheduler is the source of truth in VPP-Agent so it contains all data (even the 3rd party models not
        upstreamed into vpp-agent). The output is not compatible with agentctl configuration import functionality 
        (agentctl config update).
    agent-kvscheduler-SB-config-view.txt
        Contains vpp-agent's southbound configuration for VPP(actual configuration retrieved from VPP) as seen 
        by vpp-agent's KVScheduler component. This will actually retrieve new information from VPP.
    agent-kvscheduler-cached-config-view.txt
        Contains vpp-agent's cached northbound and southbound configuration for VPP as seen by vpp-agent's 
        KVScheduler component. This will not trigger any VPP data retrieval. It will show only cached information.
    vpp-startup-config.txt
        Contains startup configuration of VPP (retrieved from the VPP executable cmd line parameters)
    vpp-running-config(vpp-agent-SB-dump).yaml
        Contains running configuration of VPP. It is retrieved using vpp-agent.
    vpp-event-log.txt
        Contains event log from VPP. It contains events happening to VPP, but without additional information or errors.
    vpp-log.txt
        Container log of VPP.
    vpp-api-trace.txt
        Contains VPP API trace (formatted from VPP CLI command "api trace custom-dump").
    vpp-statistics-interfaces.txt
        Contains interface statistics from VPP.
    vpp-statistics-errors.txt
        Contains error statistics from VPP.
    vpp-other-*.txt
        Contains additional information from VPP by using vppctl commands.
`
	data := map[string]interface{}{
		"ReportTime": otherArgs[0].(time.Time).Unix(),
		"Version":    version.Version(),
	}
	if err := formatAsTemplate(w, format, data); err != nil {
		return err
	}
	return nil
}

func writeAgentctlVersionReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	format := `AGENTCTL:
    Version:     {{.Version}}

    Go version:  {{.GoVersion}}
    OS/Arch:     {{.OS}}/{{.Arch}}

    Build Info:
        Git commit: {{.GitCommit}}
        Git branch: {{.GitBranch}}
        User:       {{.BuildUser}}
        Host:       {{.BuildHost}}
        Built:      {{epoch .BuildTime}}

`
	data := map[string]interface{}{
		"Version":   version.Version(),
		"GitCommit": version.GitCommit(),
		"GitBranch": version.GitBranch(),
		"BuildUser": version.BuildUser(),
		"BuildHost": version.BuildHost(),
		"BuildTime": version.BuildTime(),
		"GoVersion": runtime.Version(),
		"OS":        runtime.GOOS,
		"Arch":      runtime.GOARCH,
	}
	if err := formatAsTemplate(w, format, data); err != nil {
		return err
	}
	return nil
}

func writeAgentStatusReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	const format = `State:       {{.AgentStatus.State}}
Started:     {{epoch .AgentStatus.StartTime}} ({{ago (epoch .AgentStatus.StartTime)}} ago)
Last change: {{ago (epoch .AgentStatus.LastChange)}}
Last update: {{ago (epoch .AgentStatus.LastUpdate)}}

PLUGINS
{{- range $name, $plugin := .PluginStatus}}
    {{$name}}: {{$plugin.State}}
{{- end}}
`
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subTaskActionName := "Retrieving agent status"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	status, err := cli.Client().Status(ctx)
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting status"), w, errorW, subTaskActionName)
	}

	if err := formatAsTemplate(w, format, status); err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "formatting"), w, errorW, subTaskActionName)
	}
	return nil
}

func writeAgentVersionReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	const format = `AGENT:
    App name:    {{.App}}
    Version:     {{.Version}}

    Go version:  {{.GoVersion}}
    OS/Arch:     {{.OS}}/{{.Arch}}

    Build Info:
        Git commit: {{.GitCommit}}
        Git branch: {{.GitBranch}}
        User:       {{.BuildUser}}
        Host:       {{.BuildHost}}
        Built:      {{epoch .BuildTime}}
`
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subTaskActionName := "Retrieving agent version"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	version, err := cli.Client().AgentVersion(ctx)
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting agent version"), w, errorW, subTaskActionName)
	}

	if err := formatAsTemplate(w, format, version); err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "formatting"), w, errorW, subTaskActionName)
	}
	return nil
}

func writeVPPRunningConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subTaskActionName := "Retrieving vpp running configuration"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	client, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting configuration client"), w, errorW, subTaskActionName)
	}
	resp, err := client.Dump(ctx, &configurator.DumpRequest{})
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting dump"), w, errorW, subTaskActionName)
	}

	if err := formatAsTemplate(w, "yaml", resp.Dump); err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "formatting"), w, errorW, subTaskActionName)
	}
	return nil
}

func writeAgentNBConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subTaskActionName := "Retrieving agent NB configuration"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	// TODO replace with new implementation for agentctl config get (https://github.com/ligato/vpp-agent/pull/1754)
	client, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting configuration client"),
			w, errorW, subTaskActionName)
	}
	resp, err := client.Get(ctx, &configurator.GetRequest{})
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting configuration"), w, errorW, subTaskActionName)
	}

	if err := formatAsTemplate(w, "yaml", resp.GetConfig()); err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "formatting"), w, errorW, subTaskActionName)
	}
	return nil
}

func writeKVschedulerNBConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeKVschedulerReport(
		"Retrieving agent kvscheduler NB configuration", "NB", nil, w, errorW, cli, otherArgs...)
}

func writeKVschedulerSBConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	ignoreModels := []string{ // not implemented SB retrieve
		"ligato.vpp.srv6.LocalSID",
		"ligato.vpp.srv6.Policy",
		"ligato.vpp.srv6.SRv6Global",
		"ligato.vpp.srv6.Steering",
	}
	return writeKVschedulerReport(
		"Retrieving agent kvscheduler SB configuration", "SB", ignoreModels, w, errorW, cli, otherArgs...)
}

func writeKVschedulerCachedConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeKVschedulerReport(
		"Retrieving agent kvscheduler cached configuration", "cached", nil, w, errorW, cli, otherArgs...)
}

func writeKVschedulerReport(subTaskActionName string, view string, ignoreModels []string,
	w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	// get key prefixes for all models
	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting model list"), w, errorW, subTaskActionName)
	}
	ignoreModelSet := make(map[string]struct{})
	for _, ignoreModel := range ignoreModels {
		ignoreModelSet[ignoreModel] = struct{}{}
	}
	var keyPrefixes []string
	for _, m := range allModels {
		if _, ignore := ignoreModelSet[m.ProtoName]; ignore {
			continue
		}
		keyPrefixes = append(keyPrefixes, m.KeyPrefix)
	}

	// retrieve KVScheduler data
	var (
		errs  Errors
		dumps []api.RecordedKVWithMetadata
	)
	for _, keyPrefix := range keyPrefixes {
		dump, err := cli.Client().SchedulerDump(ctx, types.SchedulerDumpOptions{
			KeyPrefix: keyPrefix,
			View:      view,
		})
		if err != nil {
			if strings.Contains(err.Error(), "no descriptor found matching the key prefix") {
				cli.Out().Write([]byte(fmt.Sprintf("Skipping key prefix %s due to: %v\n", keyPrefix, err)))
			} else {
				errs = append(errs, fmt.Errorf("Failed to get data for %s view and "+
					"key prefix %s due to: %v\n", view, keyPrefix, err))
			}
			continue
		}
		dumps = append(dumps, dump...)
	}

	// sort and print retrieved KVScheduler data
	sort.Slice(dumps, func(i, j int) bool {
		return dumps[i].Key < dumps[j].Key
	})
	printDumpTable(w, dumps)

	// error handling
	if len(errs) > 0 {
		return fileErrorPassing(cliOutputErrPassing(errs, "dumping kvscheduler data"), w, errorW, subTaskActionName)
	}
	return nil
}

func writeAgentTxnHistoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subTaskActionName := "Retrieving agent transaction history configuration"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	// get txn history
	txns, err := cli.Client().SchedulerHistory(ctx, types.SchedulerHistoryOptions{
		SeqNum: -1,
	})
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting scheduler txn history"),
			w, errorW, subTaskActionName)
	}

	// format and write it to output file
	// Note: not using one big template to print at least history summary in case of full txn log formatting fail
	w.Write([]byte("Agent transaction summary:\n"))
	var summaryBuf bytes.Buffer
	printHistoryTable(&summaryBuf, txns, true)
	w.Write([]byte(fmt.Sprintf("    %s\n",
		strings.ReplaceAll(stripTextColoring(summaryBuf.String()), "\n", "\n    "))))
	w.Write([]byte("Agent transaction log:\n"))
	var logBuf bytes.Buffer
	if err := formatAsTemplate(&logBuf, "{{.}}", txns); err != nil { // "log" format of history
		return fileErrorPassing(cliOutputErrPassing(err, "formatting"), w, errorW, subTaskActionName)
	}
	w.Write([]byte(fmt.Sprintf("    %s\n", strings.ReplaceAll(logBuf.String(), "\n", "\n    "))))

	return nil
}

func stripTextColoring(coloredText string) string {
	return regexp.MustCompile(`(?m)\x1B\[[0-9;]*m`).ReplaceAllString(coloredText, "")
}

func writeVPPInterfaceStatsReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	subTaskActionName := "Retrieving vpp interface statistics"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	// get interfaces statistics
	interfaceStats, err := cli.Client().VppGetInterfaceStats()
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting interface stats"),
			w, errorW, subTaskActionName)
	}

	// format and write it to output file
	printInterfaceStatsTable(w, interfaceStats)
	return nil
}

func writeVPPErrorStatsReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	subTaskActionName := "Retrieving vpp error statistics"
	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	// get error statistics
	errorStats, err := cli.Client().VppGetErrorStats()
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err, "getting error stats"), w, errorW, subTaskActionName)
	}

	// format and write it to output file
	printErrorStatsTable(w, errorStats)
	return nil
}

func writeVPPVersionReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	// NOTE: as task/info-specialized VPP-Agent API(REST/GRPC) should be preferred
	// (see writeVPPCLICommandReport docs) there is (plugins/govppmux/)vppcalls.VppCoreAPI containing some
	// information(not all), but it is not exposed using REST or GRPC.
	return writeVPPCLICommandReport("Retrieving vpp version", "show version verbose",
		w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("VPP:\n    %s\n",
				strings.ReplaceAll(cmdOutput, "\n", "\n    "))
		})
}

func writeVPPStartupConfigReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp startup config",
		"show version cmdline", w, errorW, cli)
}

func writeHardwareCPUReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp cpu information", "show cpu",
		w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("CPU (vppctl# %s):\n    %s\n\n",
				vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n    "))
		})
}

func writeHardwareNumaReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp numa information", "show buffers",
		w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("NUMA (indirect by viewing vpp buffer allocated for each numa node, "+
				"vppctl# %s):\n    %s\n\n", vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n    "))
		})
}

func writeHardwareVPPMainMemoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp main-heap memory information",
		"show memory main-heap verbose", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("MEMORY (only VPP related information available):\n"+
				"    vppctl# %s:\n    %s\n\n", vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n    "))
		})
}

func writeHardwareVPPNumaHeapMemoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp numa-heap memory information",
		"show memory numa-heaps", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("    vppctl# %s:\n    %s\n\n",
				vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n        "))
		})
}

func writeHardwareVPPPhysMemoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp phys memory information",
		"show physmem", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("    vppctl# %s:\n    %s\n\n",
				vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n        "))
		})
}

func writeHardwareVPPAPIMemoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp api-segment memory information",
		"show memory api-segment", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("    vppctl# %s:\n    %s\n\n",
				vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n        "))
		})
}

func writeHardwareVPPStatsMemoryReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp stats-segment memory information",
		"show memory stats-segment", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			return fmt.Sprintf("    vppctl# %s:\n    %s\n\n",
				vppCLICmd, strings.ReplaceAll(cmdOutput, "\n", "\n        "))
		})
}

func writeVPPApiTraceReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	var saveFileCmdOutput *string
	var errs Errors
	err := writeVPPCLICommandReport("Saving vpp api trace remotely to a file",
		"api trace save agentctl-report.api", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			saveFileCmdOutput = &cmdOutput
			return fmt.Sprintf("vppctl# %s:\n%s\n\n", vppCLICmd, cmdOutput) // default formatting
		})
	if err != nil {
		errs = append(errs, err)
	}

	// retrieve file location on remote machine
	// Example output: "API trace saved to /tmp/agentctl-report.api"
	fileLocation := "/tmp/agentctl-report.api"
	expectedFormattingPrefix := "API trace saved to"
	if strings.HasPrefix(*saveFileCmdOutput, expectedFormattingPrefix) {
		fileLocation = strings.TrimSpace(strings.TrimPrefix(*saveFileCmdOutput, expectedFormattingPrefix))
	}

	if err := writeVPPCLICommandReport("Retrieving vpp api trace from saved remote file",
		fmt.Sprintf("api trace custom-dump %s", fileLocation), w, errorW, cli); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func writeVPPEventLogReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	// retrieve (and write to report) vpp clock information
	var clockOutput *string
	var errs Errors
	err := writeVPPCLICommandReport("Retrieving vpp start time information(for event-log)",
		"show clock verbose", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			clockOutput = &cmdOutput
			return fmt.Sprintf("vppctl# %s (for precise conversion between vpp running time "+
				"in seconds and real time):\n%s\n\n", vppCLICmd, cmdOutput)
		})
	if err != nil {
		errs = append(errs, err)
	}

	// get VPP startup time
	vppStartUpTime, err := vppStartupTime(*clockOutput)
	addRealTime := err == nil

	// write event-log report (+ add into each line real date/time computed from VPP start timestamp)
	err = writeVPPCLICommandReport("Retrieving vpp event-log information",
		"show event-logger all", w, errorW, cli, func(vppCLICmd, cmdOutput string) string { // formatting output
			if addRealTime {
				// Example line from output: "     6.952401056: api-msg: trace_plugin_msg_ids"
				var reVPPTimeStamp = regexp.MustCompile(`(?m)^\s*(\d*.\d*):`)
				var sb strings.Builder
				for _, line := range strings.Split(cmdOutput, "\n") {
					lineVPPTimeStrSlice := reVPPTimeStamp.FindStringSubmatch(line)
					if len(lineVPPTimeStrSlice) == 0 {
						sb.WriteString(line) // error => forget conversion for this line
						continue
					}
					lineVPPTimeStr := lineVPPTimeStrSlice[0]
					if len(lineVPPTimeStr) < 1 {
						sb.WriteString(line) // error => forget conversion for this line
						continue
					}
					lineVPPTime, err := strconv.ParseFloat(
						strings.TrimSpace(lineVPPTimeStr[:len(lineVPPTimeStr)-1]), 32)
					if err != nil {
						sb.WriteString(line) // error => forget conversion for this line
						continue
					}
					lineRealTime := (*vppStartUpTime).Add(time.Duration(int(lineVPPTime*1_000_000)) * time.Microsecond)
					sb.WriteString(strings.Replace(line, lineVPPTimeStr, fmt.Sprintf("%s(%s +-1s)",
						lineVPPTimeStr, lineRealTime.UTC().Format(time.RFC1123)), 1) + "\n")
				}
				return fmt.Sprintf("vppctl# %s:\n%s\n\n", vppCLICmd, sb.String())
			}
			return fmt.Sprintf("vppctl# %s:\n%s\n\n", vppCLICmd, cmdOutput) // default formatting
		})
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}

func vppStartupTime(vppClockCmdOutput string) (*time.Time, error) {
	// vppctl# show clock verbose
	// Example output: "Time now 705.541644, reftime 705.541644, error 0.000000, clocks/sec 2591995355.407570, Wed, 11 Nov 2020 14:32:39 GMT"

	// get real date/time from command output
	var reDateTime = regexp.MustCompile(`[^\s,]*,[^,]*\z`)
	cmdTime, err := time.Parse(time.RFC1123, reDateTime.FindString(strings.ReplaceAll(vppClockCmdOutput, "\n", "")))
	if err != nil {
		return nil, fmt.Errorf("can't parse ref time form VPP "+
			"show clock command due to: %v (cmd output=%s)", err, vppClockCmdOutput)
	}

	// get VPP time (seconds from VPP start)
	var reReftime = regexp.MustCompile(`reftime\W*(\d*.\d*)\D`)
	strSubmatch := reReftime.FindStringSubmatch(vppClockCmdOutput)
	if len(strSubmatch) < 2 {
		return nil, fmt.Errorf("can't find reftime in vpp clock cmd output %v", vppClockCmdOutput)
	}
	vppTimeAtCmdTime, err := strconv.ParseFloat(strSubmatch[1], 32)
	if err != nil {
		return nil, fmt.Errorf("can't parse reftime string(%v) to float due to: %v", strSubmatch[1], err)
	}

	// compute VPP startup time
	vppStartUpTime := cmdTime.Add(-time.Duration(int(vppTimeAtCmdTime*1_000_000)) * time.Microsecond)
	return &vppStartUpTime, nil
}

func writeVPPLogReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp log information",
		"show logging", w, errorW, cli)
}

func writeVPPSRv6LocalsidReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp SRv6 localsid information",
		"show sr localsid", w, errorW, cli)
}

func writeVPPSRv6PolicyReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp SRv6 policies information",
		"show sr policies", w, errorW, cli)
}

func writeVPPSRv6SteeringReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp SRv6 steering information",
		"show sr steering-policies", w, errorW, cli)
}

// writeVPPCLICommandReport writes to file a report based purely on one VPP CLI command output.
// Before using this function think about using some task/info-specialized VPP-Agent API(REST/GRPC) because
// it solves compatibility issues with different VPP versions. This method don't care whether the given
// vppCLICmd is actually a valid VPP CLI command on the version of VPP to which it will connect to. In case
// of incompatibility, this subreport will fail (the whole report will be created, just like other subreports
// if they don't fail on their own).
func writeVPPCLICommandReport(subTaskActionName string, vppCLICmd string, w io.Writer, errorW io.Writer,
	cli agentcli.Cli, otherArgs ...interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cliOutputDefer, cliOutputErrPassing := subTaskCliOutputs(cli, subTaskActionName)
	defer cliOutputDefer(cli)

	cmdOutput, err := cli.Client().VppRunCli(ctx, vppCLICmd)
	if err != nil {
		return fileErrorPassing(cliOutputErrPassing(err), w, errorW, subTaskActionName)
	}
	formattedOutput := fmt.Sprintf("vppctl# %s\n%s\n", vppCLICmd, cmdOutput)
	if len(otherArgs) > 0 {
		formattedOutput = otherArgs[0].(func(string, string) string)(vppCLICmd, cmdOutput)
	}
	fmt.Fprintf(w, formattedOutput)
	return nil
}

func printErrorStatsTable(out io.Writer, errorStats *govppapi.ErrorStats) {
	table := tablewriter.NewWriter(out)
	header := []string{
		"Statistics name", "Error counter",
	}
	table.SetHeader(header)
	table.SetRowLine(false)
	table.SetAutoWrapText(false)

	for _, errorStat := range errorStats.Errors {
		var valSum uint64 = 0
		// errorStat.Values are per worker counters
		for _, val := range errorStat.Values {
			valSum += val
		}
		row := []string{
			errorStat.CounterName,
			fmt.Sprint(valSum),
		}
		table.Append(row)
	}
	table.Render()
}

func printInterfaceStatsTable(out io.Writer, interfaceStats *govppapi.InterfaceStats) {
	table := tablewriter.NewWriter(out)
	header := []string{
		"Index", "Name", "Rx", "Tx", "Rx errors", "Tx errors", "Drops", "Rx unicast/multicast/broadcast",
		"Tx unicast/multicast/broadcast", "Other",
	}
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	for _, interfaceStat := range interfaceStats.Interfaces {
		row := []string{
			fmt.Sprint(interfaceStat.InterfaceIndex),
			fmt.Sprint(interfaceStat.InterfaceName),
			fmt.Sprintf("%d packets (%d bytes)", interfaceStat.Rx.Packets, interfaceStat.Rx.Bytes),
			fmt.Sprintf("%d packets (%d bytes)", interfaceStat.Tx.Packets, interfaceStat.Tx.Bytes),
			fmt.Sprint(interfaceStat.RxErrors),
			fmt.Sprint(interfaceStat.TxErrors),
			fmt.Sprint(interfaceStat.Drops),
			fmt.Sprintf("%d packets (%d bytes)\n%d packets (%d bytes)\n%d packets (%d bytes)",
				interfaceStat.RxUnicast.Packets, interfaceStat.RxUnicast.Bytes,
				interfaceStat.RxMulticast.Packets, interfaceStat.RxMulticast.Bytes,
				interfaceStat.RxBroadcast.Packets, interfaceStat.RxBroadcast.Bytes,
			),
			fmt.Sprintf("%d packets (%d bytes)\n%d packets (%d bytes)\n%d packets (%d bytes)",
				interfaceStat.TxUnicast.Packets, interfaceStat.TxUnicast.Bytes,
				interfaceStat.TxMulticast.Packets, interfaceStat.TxMulticast.Bytes,
				interfaceStat.TxBroadcast.Packets, interfaceStat.TxBroadcast.Bytes,
			),
			fmt.Sprintf("Punts: %d\n"+
				"IPv4: %d\n"+
				"IPv6: %d\n"+
				"RxNoBuf: %d\n"+
				"RxMiss: %d\n"+
				"Mpls: %d",
				interfaceStat.Punts,
				interfaceStat.IP4,
				interfaceStat.IP6,
				interfaceStat.RxNoBuf,
				interfaceStat.RxMiss,
				interfaceStat.Mpls),
		}
		table.Append(row)
	}
	table.Render()
}

func packErrors(errors ...error) Errors {
	var errs Errors
	for _, err := range errors {
		if err != nil {
			if alreadyPackedError, isPacked := err.(Errors); isPacked {
				errs = append(errs, alreadyPackedError...)
				continue
			}
			errs = append(errs, err)
		}
	}
	return errs
}

func subTaskCliOutputs(cli agentcli.Cli, subTaskActionName string) (func(agentcli.Cli), func(error, ...string) error) {
	cli.Out().Write([]byte(fmt.Sprintf("%s... ", subTaskActionName)))
	subTaskResult := "Done."
	pSubTaskResult := &subTaskResult
	return func(cli agentcli.Cli) {
			cli.Out().Write([]byte(fmt.Sprintf("%s\n", *pSubTaskResult)))
		}, func(err error, failedActions ...string) error {
			if len(failedActions) > 0 {
				err = fmt.Errorf("%s failed due to:\n%v", failedActions[0], err)
			}
			subTaskResult = fmt.Sprintf("Error... (this error will be part of the report zip file, "+
				"see %s):\n%v\nSkipping this subreport.\n", failedReportFileName, err)
			pSubTaskResult = &subTaskResult
			return err
		}
}

func fileErrorPassing(err error, w io.Writer, errorW io.Writer, subTaskActionName string) error {
	errorStr := fmt.Sprintf("%s... failed due to:\n%v\n", subTaskActionName, err)
	w.Write([]byte(fmt.Sprintf("<<\n%s>>\n", errorStr)))
	errorW.Write([]byte(fmt.Sprintf("%s\n%s\n", strings.Repeat("#", 70), errorStr)))
	return err
}

func writeReportTo(fileName string, dirName string,
	writeFunc func(io.Writer, io.Writer, agentcli.Cli, ...interface{}) error,
	cli agentcli.Cli, otherWriteFuncArgs ...interface{}) (err error) {
	// open file (and close it in the end)
	f, err := os.OpenFile(filepath.Join(dirName, fileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		err = fmt.Errorf("can't open file %v due to: %v", filepath.Join(dirName, fileName), err)
		return
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("can't close file %v due to: %v", filepath.Join(dirName, fileName), closeErr)
		}
	}()

	// open error file (and close it in the end)
	errorFile, err := os.OpenFile(filepath.Join(dirName, failedReportFileName),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		err = fmt.Errorf("can't open error file %v due to: %v",
			filepath.Join(dirName, failedReportFileName), err)
		return
	}
	defer func() {
		if closeErr := errorFile.Close(); closeErr != nil {
			err = fmt.Errorf("can't close error file %v due to: %v",
				filepath.Join(dirName, failedReportFileName), closeErr)
		}
	}()

	// append some report to file
	err = writeFunc(f, errorFile, cli, otherWriteFuncArgs...)
	return
}

// createZipFile compresses content of directory dirName into a single zip archive file named filename.
// Both arguments should be absolut path file/directory names. The directory content excludes subdirectories.
func createZipFile(zipFileName string, dirName string) (err error) {
	// create zip writer
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return fmt.Errorf("can't create empty zip file(%v) due to: %v", zipFileName, err)
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			err = fmt.Errorf("can't close zip file %v due to: %v", zipFileName, closeErr)
		}
	}()
	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			err = fmt.Errorf("can't close zip file writer for zip file %v due to: %v", zipFileName, closeErr)
		}
	}()

	// Add files to zip
	dirItems, err := ioutil.ReadDir(dirName)
	if err != nil {
		return fmt.Errorf("can't read report directory(%v) due to: %v", dirName, err)
	}
	for _, dirItem := range dirItems {
		if !dirItem.IsDir() {
			if err = addFileToZip(zipWriter, filepath.Join(dirName, dirItem.Name())); err != nil {
				return fmt.Errorf("can't add file dirItem.Name() to report zip file due to: %v", err)
			}
		}
	}
	return nil
}

// addFileToZip adds file to zip file by using zip.Writer. The file name should be a absolute path.
func addFileToZip(zipWriter *zip.Writer, filename string) error {
	// open file for addition
	fileToZip, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("can't open file %v due to: %v", filename, err)
	}
	defer func() {
		if closeErr := fileToZip.Close(); closeErr != nil {
			err = fmt.Errorf("can't close zip file %v opened "+
				"for file appending due to: %v", filename, closeErr)
		}
	}()

	// get information from file for addition
	info, err := fileToZip.Stat()
	if err != nil {
		return fmt.Errorf("can't get information about file (%v) "+
			"that should be added to zip file due to: %v", filename, err)
	}

	// add file to zip file
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("can't create zip file info header for file %v due to: %v", filename, err)
	}
	header.Method = zip.Deflate // enables compression
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("can't create zip header for file %v due to: %v", filename, err)
	}
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return fmt.Errorf("can't copy content of file %v to zip file due to: %v", filename, err)
	}
	return nil
}
