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
	"strings"
	"time"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"

	"github.com/spf13/cobra"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/version"
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
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.OutputDirectory, "output-directory", "o", "",
		"Output directory (as absolute path) where report zip file will be written. "+
			"Default is current directory.")
	return cmd
}

type ReportOptions struct {
	OutputDirectory string
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
	errCounter := errorCounter( // TODO add other reports
		writeReportTo("_report.txt", dirName, writeReportHeader, cli, reportTime), // TODO add info about other report files(what is there)
		writeReportTo("software-versions.txt", dirName, writeAgentctlVersionReport, cli),
		writeReportTo("software-versions.txt", dirName, writeAgentVersionReport, cli),
		writeReportTo("software-versions.txt", dirName, writeVPPVersionReport, cli),
		writeReportTo("agent-status.txt", dirName, writeAgentStatusReport, cli),
		writeReportTo("vpp-startup-config.txt", dirName, writeVPPStartupConfigReport, cli),
		writeReportTo("vpp-running-config(vpp-agent-SB-dump).yaml", dirName, writeVPPRunningConfigReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareCPUReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareNumaReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPMainMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPNumaHeapMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPPhysMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPAPIMemoryReport, cli),
		writeReportTo("hardware.txt", dirName, writeHardwareVPPStatsMemoryReport, cli),
		writeReportTo("vpp-event-log.txt", dirName, writeVPPEventLogReport, cli),
		writeReportTo("vpp-log.txt", dirName, writeVPPLogReport, cli),
		writeReportTo("agent-transaction-history.txt", dirName, writeAgentTxnHistoryReport, cli),
	)
	if errCounter > 0 {
		cli.Out().Write([]byte(fmt.Sprintf("\ncan't write %d subreport(s) "+
			"(full list with errors will be in packed zip file as file %s)\n\n", errCounter, failedReportFileName)))
	} else { //
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
	cli.Out().Write([]byte(fmt.Sprintf("Done.\n\nReport file: %v\n", zipFileName)))

	return nil
}

func writeReportHeader(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	// using template also for simple cases to be consistent with variable formatting (i.e. time formatting)
	format := `################# REPORT #################
report creation time:    {{epoch .ReportTime}}
report version:          {{.Version}} (=AGENTCTL version)

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

func writeVPPVersionReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	// NOTE: task/info-specialized VPP-Agent API(REST/GRPC) should be preferred (in compare to direct VPP
	// CLI commands) due to solved compatibility with different VPP version
	// => (plugins/govppmux/)vppcalls.VppCoreAPI contains some information(not all), but it is not exposed
	// using REST or GRPC.
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

func writeVPPEventLogReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp event-log information",
		"show event-logger all", w, errorW, cli)
}

func writeVPPLogReport(w io.Writer, errorW io.Writer, cli agentcli.Cli, otherArgs ...interface{}) error {
	return writeVPPCLICommandReport("Retrieving vpp log information",
		"show logging", w, errorW, cli)
}

// TODO remove comments for development
// show errors -> only error counts -> govpp instead?

//vpp# show memory api-segment
//API segment
//total: 16.00M, used: 1.50M, free: 14.49M, trimmable: 14.49M
//free chunks 2 free fastbin blks 0
//max total allocated 16.00M
//vpp# show memory stats-segment
//Stats segment
//total: 31.99M, used: 783.62K, free: 31.23M, trimmable: 30.59M
//free chunks 5 free fastbin blks 0
//max total allocated 31.99M
//vpp# show memeory main-heap
//show: unknown input `memeory main-heap'
//vpp# show memory main-heap
//Thread 0 vpp_main
//  virtual memory start 0x7ffaae1f0000, size 2123903k, 530975 pages, page size 4k
//    numa 0: 11056 pages, 44224k
//    not mapped: 247873 pages, 991492k
//    unknown: 272046 pages, 1088184k
//  total: 2.03G, used: 1.39G, free: 641.25M, trimmable: 63.83K
//vpp# show memory numa-heaps
//Numa 0 uses the main heap...

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

// TODO possible vpp things for output
//vppcli "show clock" >> $vpp_info
//vppcli "show version verbose" >> $vpp_info
//vppcli "show plugins" >> $vpp_info
//vppcli "show cpu" >> $vpp_info
//vppcli "show version cmdline" >> $vpp_info
//vppcli "show threads" >> $vpp_info
//vppcli "show physmem" >> $vpp_info
//vppcli "show memory main-heap verbose" >> $vpp_info
//vppcli "show memory api-segment verbose" >> $vpp_info
//vppcli "show memory stats-segment verbose" >> $vpp_info
//vppcli "show api histogram" >> $vpp_info
//vppcli "show api ring-stats" >> $vpp_info
//vppcli "show api trace-status" >> $vpp_info
//vppcli "api trace status" >> $vpp_info
//vppcli "show api clients" >> $vpp_info
//vppcli "show unix files" >> $vpp_info
//vppcli "show unix errors" >> $vpp_info
//vppcli "show event-logger" >> $vpp_info
//vppcli "show ip fib summary" >> $vpp_info

//runCmd("show node counters")
//runCmd("show runtime")
//runCmd("show buffers")
//runCmd("show memory")
//runCmd("show ip fib")
//runCmd("show ip6 fib")

func errorCounter(errors ...error) int {
	counter := 0
	for _, err := range errors {
		if err != nil {
			counter++
		}
	}
	return counter
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
