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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bennyscetbun/jsongo"
	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/version"
)

var (
	theInputFile = flag.String("input-file", "", "Input file with VPP API in JSON format.")
	theInputDir  = flag.String("input-dir", "/usr/share/vpp/api", "Input directory with VPP API files in JSON format.")
	theOutputDir = flag.String("output-dir", ".", "Output directory where package folders will be generated.")

	includeAPIVer      = flag.Bool("include-apiver", true, "Include APIVersion constant for each module.")
	includeServices    = flag.Bool("include-services", true, "Include RPC service api and client implementation.")
	includeComments    = flag.Bool("include-comments", false, "Include JSON API source in comments for each object.")
	includeBinapiNames = flag.Bool("include-binapi-names", false, "Include binary API names in struct tag.")

	continueOnError = flag.Bool("continue-onerror", false, "Continue with next file on error.")
	debugMode       = flag.Bool("debug", os.Getenv("GOVPP_DEBUG") != "", "Enable debug mode.")

	printVersion = flag.Bool("version", false, "Prints current version and exits.")
)

func main() {
	flag.Parse()

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.NArg() > 0 {
		switch cmd := flag.Arg(0); cmd {
		case "version":
			fmt.Fprintln(os.Stdout, version.Verbose())
			os.Exit(0)

		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
			flag.Usage()
			os.Exit(2)
		}
	}

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("debug mode enabled")
	}

	if err := run(*theInputFile, *theInputDir, *theOutputDir, *continueOnError); err != nil {
		logrus.Errorln("binapi-generator:", err)
		os.Exit(1)
	}
}

func run(inputFile, inputDir string, outputDir string, continueErr bool) error {
	if inputFile == "" && inputDir == "" {
		return fmt.Errorf("input-file or input-dir must be specified")
	}

	if inputFile != "" {
		// process one input file
		if err := generateFromFile(inputFile, outputDir); err != nil {
			return fmt.Errorf("code generation from %s failed: %v\n", inputFile, err)
		}
	} else {
		// process all files in specified directory
		dir, err := filepath.Abs(inputDir)
		if err != nil {
			return fmt.Errorf("invalid input directory: %v\n", err)
		}
		files, err := getInputFiles(inputDir, 1)
		if err != nil {
			return fmt.Errorf("problem getting files from input directory: %v\n", err)
		} else if len(files) == 0 {
			return fmt.Errorf("no input files found in input directory: %v\n", dir)
		}
		for _, file := range files {
			if err := generateFromFile(file, outputDir); err != nil {
				if continueErr {
					logrus.Warnf("code generation from %s failed: %v (error ignored)\n", file, err)
					continue
				} else {
					return fmt.Errorf("code generation from %s failed: %v\n", file, err)
				}
			}
		}
	}

	return nil
}

// getInputFiles returns all input files located in specified directory
func getInputFiles(inputDir string, deep int) (files []string, err error) {
	entries, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s failed: %v", inputDir, err)
	}
	for _, e := range entries {
		if e.IsDir() && deep > 0 {
			nestedDir := filepath.Join(inputDir, e.Name())
			if nested, err := getInputFiles(nestedDir, deep-1); err != nil {
				return nil, err
			} else {
				files = append(files, nested...)
			}
		} else if strings.HasSuffix(e.Name(), inputFileExt) {
			files = append(files, filepath.Join(inputDir, e.Name()))
		}
	}
	return files, nil
}

func parseInputJSON(inputData []byte) (*jsongo.JSONNode, error) {
	jsonRoot := new(jsongo.JSONNode)
	if err := json.Unmarshal(inputData, jsonRoot); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON failed: %v", err)
	}
	return jsonRoot, nil
}

// generateFromFile generates Go package from one input JSON file
func generateFromFile(inputFile, outputDir string) error {
	// create generator context
	ctx, err := newContext(inputFile, outputDir)
	if err != nil {
		return err
	}

	logf("------------------------------------------------------------")
	logf("module: %s", ctx.moduleName)
	logf(" - input: %s", ctx.inputFile)
	logf(" - output: %s", ctx.outputFile)
	logf("------------------------------------------------------------")

	// prepare options
	ctx.includeAPIVersion = *includeAPIVer
	ctx.includeComments = *includeComments
	ctx.includeBinapiNames = *includeBinapiNames
	ctx.includeServices = *includeServices

	// read API definition from input file
	ctx.inputData, err = ioutil.ReadFile(ctx.inputFile)
	if err != nil {
		return fmt.Errorf("reading input file %s failed: %v", ctx.inputFile, err)
	}
	// parse JSON data into objects
	jsonRoot, err := parseInputJSON(ctx.inputData)
	if err != nil {
		return fmt.Errorf("parsing JSON input failed: %v", err)
	}
	ctx.packageData, err = parsePackage(ctx, jsonRoot)
	if err != nil {
		return fmt.Errorf("parsing package %s failed: %v", ctx.packageName, err)
	}

	// generate Go package code
	var buf bytes.Buffer
	if err := generatePackage(ctx, &buf); err != nil {
		return fmt.Errorf("generating code for package %s failed: %v", ctx.packageName, err)
	}

	// create output directory
	packageDir := filepath.Dir(ctx.outputFile)
	if err := os.MkdirAll(packageDir, 0775); err != nil {
		return fmt.Errorf("creating output dir %s failed: %v", packageDir, err)
	}
	// write generated code to output file
	if err := ioutil.WriteFile(ctx.outputFile, buf.Bytes(), 0666); err != nil {
		return fmt.Errorf("writing to output file %s failed: %v", ctx.outputFile, err)
	}

	// go format the output file (fail probably means the output is not compilable)
	cmd := exec.Command("gofmt", "-w", ctx.outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gofmt failed: %v\n%s", err, string(output))
	}

	return nil
}

func logf(f string, v ...interface{}) {
	if *debugMode {
		logrus.Debugf(f, v...)
	}
}
