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
)

var (
	inputFile          = flag.String("input-file", "", "Input file with VPP API in JSON format.")
	inputDir           = flag.String("input-dir", ".", "Input directory with VPP API files in JSON format.")
	outputDir          = flag.String("output-dir", ".", "Output directory where package folders will be generated.")
	includeAPIVer      = flag.Bool("include-apiver", true, "Include APIVersion constant for each module.")
	includeComments    = flag.Bool("include-comments", false, "Include JSON API source in comments for each object.")
	includeBinapiNames = flag.Bool("include-binapi-names", false, "Include binary API names in struct tag.")
	includeServices    = flag.Bool("include-services", false, "Include service interface with client implementation.")
	continueOnError    = flag.Bool("continue-onerror", false, "Continue with next file on error.")
	debug              = flag.Bool("debug", debugMode, "Enable debug mode.")
)

var debugMode = os.Getenv("DEBUG_BINAPI_GENERATOR") != ""

func main() {
	flag.Parse()
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if *inputFile == "" && *inputDir == "" {
		fmt.Fprintln(os.Stderr, "ERROR: input-file or input-dir must be specified")
		os.Exit(1)
	}

	if *inputFile != "" {
		// process one input file
		if err := generateFromFile(*inputFile, *outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: code generation from %s failed: %v\n", *inputFile, err)
			os.Exit(1)
		}
	} else {
		// process all files in specified directory
		dir, err := filepath.Abs(*inputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid input directory: %v\n", err)
			os.Exit(1)
		}
		files, err := getInputFiles(*inputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: problem getting files from input directory: %v\n", err)
			os.Exit(1)
		} else if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: no input files found in input directory: %v\n", dir)
			os.Exit(1)
		}
		for _, file := range files {
			if err := generateFromFile(file, *outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: code generation from %s failed: %v\n", file, err)
				if *continueOnError {
					continue
				}
				os.Exit(1)
			}
		}
	}
}

// getInputFiles returns all input files located in specified directory
func getInputFiles(inputDir string) (res []string, err error) {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s failed: %v", inputDir, err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), inputFileExt) {
			res = append(res, filepath.Join(inputDir, f.Name()))
		}
	}
	return res, nil
}

// generateFromFile generates Go package from one input JSON file
func generateFromFile(inputFile, outputDir string) error {
	logf("generating from file: %s", inputFile)
	logf("------------------------------------------------------------")
	defer logf("------------------------------------------------------------")

	ctx, err := getContext(inputFile, outputDir)
	if err != nil {
		return err
	}

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
	jsonRoot := new(jsongo.JSONNode)
	if err := json.Unmarshal(ctx.inputData, jsonRoot); err != nil {
		return fmt.Errorf("unmarshalling JSON failed: %v", err)
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

	// count number of lines in generated output file
	cmd = exec.Command("wc", "-l", ctx.outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		logf("wc command failed: %v\n%s", err, string(output))
	} else {
		logf("number of generated lines: %s", output)
	}

	return nil
}

func logf(f string, v ...interface{}) {
	if *debug {
		logrus.Debugf(f, v...)
	}
}
