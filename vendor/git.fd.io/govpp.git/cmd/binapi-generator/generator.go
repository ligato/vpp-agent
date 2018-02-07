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

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bennyscetbun/jsongo"
)

// MessageType represents the type of a VPP message.
type messageType int

const (
	requestMessage messageType = iota // VPP request message
	replyMessage                      // VPP reply message
	eventMessage                      // VPP event message
	otherMessage                      // other VPP message
)

const (
	apiImportPath = "git.fd.io/govpp.git/api" // import path of the govpp API
	inputFileExt  = ".json"                   // filename extension of files that should be processed as the input
)

// context is a structure storing details of a particular code generation task
type context struct {
	inputFile   string            // file with input JSON data
	inputData   []byte            // contents of the input file
	inputBuff   *bytes.Buffer     // contents of the input file currently being read
	inputLine   int               // currently processed line in the input file
	outputFile  string            // file with output data
	packageName string            // name of the Go package being generated
	packageDir  string            // directory where the package source files are located
	types       map[string]string // map of the VPP typedef names to generated Go typedef names
}

func main() {
	inputFile := flag.String("input-file", "", "Input JSON file.")
	inputDir := flag.String("input-dir", ".", "Input directory with JSON files.")
	outputDir := flag.String("output-dir", ".", "Output directory where package folders will be generated.")
	flag.Parse()

	if *inputFile == "" && *inputDir == "" {
		fmt.Fprintln(os.Stderr, "ERROR: input-file or input-dir must be specified")
		os.Exit(1)
	}

	var err, tmpErr error
	if *inputFile != "" {
		// process one input file
		err = generateFromFile(*inputFile, *outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: code generation from %s failed: %v\n", *inputFile, err)
		}
	} else {
		// process all files in specified directory
		files, err := getInputFiles(*inputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: code generation failed: %v\n", err)
		}
		for _, file := range files {
			tmpErr = generateFromFile(file, *outputDir)
			if tmpErr != nil {
				fmt.Fprintf(os.Stderr, "ERROR: code generation from %s failed: %v\n", file, err)
				err = tmpErr // remember that the error occurred
			}
		}
	}
	if err != nil {
		os.Exit(1)
	}
}

// getInputFiles returns all input files located in specified directory
func getInputFiles(inputDir string) ([]string, error) {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s failed: %v", inputDir, err)
	}
	res := make([]string, 0)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), inputFileExt) {
			res = append(res, inputDir+"/"+f.Name())
		}
	}
	return res, nil
}

// generateFromFile generates Go bindings from one input JSON file
func generateFromFile(inputFile, outputDir string) error {
	ctx, err := getContext(inputFile, outputDir)
	if err != nil {
		return err
	}
	// read the file
	ctx.inputData, err = readFile(inputFile)
	if err != nil {
		return err
	}

	// parse JSON
	jsonRoot, err := parseJSON(ctx.inputData)
	if err != nil {
		return err
	}

	// create output directory
	err = os.MkdirAll(ctx.packageDir, 0777)
	if err != nil {
		return fmt.Errorf("creating output directory %s failed: %v", ctx.packageDir, err)
	}

	// open output file
	f, err := os.Create(ctx.outputFile)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("creating output file %s failed: %v", ctx.outputFile, err)
	}
	w := bufio.NewWriter(f)

	// generate Go package code
	err = generatePackage(ctx, w, jsonRoot)
	if err != nil {
		return err
	}

	// go format the output file (non-fatal if fails)
	exec.Command("gofmt", "-w", ctx.outputFile).Run()

	return nil
}

// getContext returns context details of the code generation task
func getContext(inputFile, outputDir string) (*context, error) {
	if !strings.HasSuffix(inputFile, inputFileExt) {
		return nil, fmt.Errorf("invalid input file name %s", inputFile)
	}

	ctx := &context{inputFile: inputFile}
	inputFileName := filepath.Base(inputFile)

	ctx.packageName = inputFileName[0:strings.Index(inputFileName, ".")]
	if ctx.packageName == "interface" {
		// 'interface' cannot be a package name, it is a go keyword
		ctx.packageName = "interfaces"
	}

	ctx.packageDir = outputDir + "/" + ctx.packageName + "/"
	ctx.outputFile = ctx.packageDir + ctx.packageName + ".go"

	return ctx, nil
}

// readFile reads content of a file into memory
func readFile(inputFile string) ([]byte, error) {

	inputData, err := ioutil.ReadFile(inputFile)

	if err != nil {
		return nil, fmt.Errorf("reading data from file failed: %v", err)
	}

	return inputData, nil
}

// parseJSON parses a JSON data into an in-memory tree
func parseJSON(inputData []byte) (*jsongo.JSONNode, error) {
	root := jsongo.JSONNode{}

	err := json.Unmarshal(inputData, &root)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshall failed: %v", err)
	}

	return &root, nil

}

// generatePackage generates Go code of a package from provided JSON
func generatePackage(ctx *context, w *bufio.Writer, jsonRoot *jsongo.JSONNode) error {
	// generate file header
	generatePackageHeader(ctx, w, jsonRoot)

	// generate data types
	ctx.inputBuff = bytes.NewBuffer(ctx.inputData)
	ctx.inputLine = 0
	ctx.types = make(map[string]string)
	types := jsonRoot.Map("types")
	for i := 0; i < types.Len(); i++ {
		typ := types.At(i)
		err := generateMessage(ctx, w, typ, true)
		if err != nil {
			return err
		}
	}

	// generate messages
	ctx.inputBuff = bytes.NewBuffer(ctx.inputData)
	ctx.inputLine = 0
	messages := jsonRoot.Map("messages")
	for i := 0; i < messages.Len(); i++ {
		msg := messages.At(i)
		err := generateMessage(ctx, w, msg, false)
		if err != nil {
			return err
		}
	}

	// flush the data:
	err := w.Flush()
	if err != nil {
		return fmt.Errorf("flushing data to %s failed: %v", ctx.outputFile, err)
	}

	return nil
}

// generateMessage generates Go code of one VPP message encoded in JSON into provided writer
func generateMessage(ctx *context, w io.Writer, msg *jsongo.JSONNode, isType bool) error {
	if msg.Len() == 0 || msg.At(0).GetType() != jsongo.TypeValue {
		return errors.New("invalid JSON for message specified")
	}

	msgName, ok := msg.At(0).Get().(string)
	if !ok {
		return fmt.Errorf("invalid JSON for message specified, message name is %T, not a string", msg.At(0).Get())
	}
	structName := camelCaseName(strings.Title(msgName))

	// generate struct fields into the slice & determine message type
	fields := make([]string, 0)
	msgType := otherMessage
	wasClientIndex := false
	for j := 0; j < msg.Len(); j++ {
		if jsongo.TypeArray == msg.At(j).GetType() {
			fld := msg.At(j)
			if !isType {
				// determine whether ths is a request / reply / other message
				fieldName, ok := fld.At(1).Get().(string)
				if ok {
					if j == 2 {
						if fieldName == "client_index" {
							// "client_index" as the second member, this might be an event message or a request
							msgType = eventMessage
							wasClientIndex = true
						} else if fieldName == "context" {
							// reply needs "context" as the second member
							msgType = replyMessage
						}
					} else if j == 3 {
						if wasClientIndex && fieldName == "context" {
							// request needs "client_index" as the second member and "context" as the third member
							msgType = requestMessage
						}
					}
				}
			}
			err := processMessageField(ctx, &fields, fld, isType)
			if err != nil {
				return err
			}
		}
	}

	// generate struct comment
	generateMessageComment(ctx, w, structName, msgName, isType)

	// generate struct header
	fmt.Fprintln(w, "type", structName, "struct {")

	// print out the fields
	for _, field := range fields {
		fmt.Fprintln(w, field)
	}

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate name getter
	if isType {
		generateTypeNameGetter(w, structName, msgName)
	} else {
		generateMessageNameGetter(w, structName, msgName)
	}

	// generate message type getter method
	if !isType {
		generateMessageTypeGetter(w, structName, msgType)
	}

	// generate CRC getter
	crcIf := msg.At(msg.Len() - 1).At("crc").Get()
	if crc, ok := crcIf.(string); ok {
		generateCrcGetter(w, structName, crc)
	}

	// generate message factory
	if !isType {
		generateMessageFactory(w, structName)
	}

	// if this is a type, save it in the map for later use
	if isType {
		ctx.types[fmt.Sprintf("vl_api_%s_t", msgName)] = structName
	}

	return nil
}

// processMessageField process JSON describing one message field into Go code emitted into provided slice of message fields
func processMessageField(ctx *context, fields *[]string, fld *jsongo.JSONNode, isType bool) error {
	if fld.Len() < 2 || fld.At(0).GetType() != jsongo.TypeValue || fld.At(1).GetType() != jsongo.TypeValue {
		return errors.New("invalid JSON for message field specified")
	}
	fieldVppType, ok := fld.At(0).Get().(string)
	if !ok {
		return fmt.Errorf("invalid JSON for message specified, field type is %T, not a string", fld.At(0).Get())
	}
	fieldName, ok := fld.At(1).Get().(string)
	if !ok {
		return fmt.Errorf("invalid JSON for message specified, field name is %T, not a string", fld.At(1).Get())
	}

	// skip internal fields
	fieldNameLower := strings.ToLower(fieldName)
	if fieldNameLower == "crc" || fieldNameLower == "_vl_msg_id" {
		return nil
	}
	if !isType && len(*fields) == 0 && (fieldNameLower == "client_index" || fieldNameLower == "context") {
		return nil
	}

	fieldName = strings.TrimPrefix(fieldName, "_")
	fieldName = camelCaseName(strings.Title(fieldName))

	fieldStr := ""
	isArray := false
	arraySize := 0

	fieldStr += "\t" + fieldName + " "
	if fld.Len() > 2 {
		isArray = true
		arraySize = int(fld.At(2).Get().(float64))
		fieldStr += "[]"
	}

	dataType := translateVppType(ctx, fieldVppType, isArray)
	fieldStr += dataType

	if isArray {
		if arraySize == 0 {
			// variable sized array
			if fld.Len() > 3 {
				// array size is specified by another field
				arraySizeField := string(fld.At(3).Get().(string))
				arraySizeField = camelCaseName(strings.Title(arraySizeField))
				// find & update the field that specifies the array size
				for i, f := range *fields {
					if strings.Contains(f, fmt.Sprintf("\t%s ", arraySizeField)) {
						(*fields)[i] += fmt.Sprintf("\t`struc:\"sizeof=%s\"`", fieldName)
					}
				}
			}
		} else {
			// fixed size array
			fieldStr += fmt.Sprintf("\t`struc:\"[%d]%s\"`", arraySize, dataType)
		}
	}

	*fields = append(*fields, fieldStr)
	return nil
}

// generatePackageHeader generates package header into provider writer
func generatePackageHeader(ctx *context, w io.Writer, rootNode *jsongo.JSONNode) {
	fmt.Fprintln(w, "// Code generated by govpp binapi-generator DO NOT EDIT.")
	fmt.Fprintln(w, "// Package "+ctx.packageName+" represents the VPP binary API of the '"+ctx.packageName+"' VPP module.")
	fmt.Fprintln(w, "// Generated from '"+ctx.inputFile+"'")

	fmt.Fprintln(w, "package "+ctx.packageName)

	fmt.Fprintln(w, "import \""+apiImportPath+"\"")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "// VlApiVersion contains version of the API.")
	vlAPIVersion := rootNode.Map("vl_api_version")
	if vlAPIVersion != nil {
		fmt.Fprintln(w, "const VlAPIVersion = ", vlAPIVersion.Get())
	}
	fmt.Fprintln(w)
}

// generateMessageComment generates comment for a message into provider writer
func generateMessageComment(ctx *context, w io.Writer, structName string, msgName string, isType bool) {
	fmt.Fprintln(w)
	if isType {
		fmt.Fprintln(w, "// "+structName+" represents the VPP binary API data type '"+msgName+"'.")
	} else {
		fmt.Fprintln(w, "// "+structName+" represents the VPP binary API message '"+msgName+"'.")
	}

	// print out the source of the generated message - the JSON
	msgFound := false
	msgTitle := "\"" + msgName + "\","
	var msgIndent int
	for {
		lineBuff, err := ctx.inputBuff.ReadBytes('\n')
		if err != nil {
			break
		}
		ctx.inputLine++
		line := string(lineBuff)

		if !msgFound {
			msgIndent = strings.Index(line, msgTitle)
			if msgIndent > -1 {
				prefix := line[:msgIndent]
				suffix := line[msgIndent+len(msgTitle):]
				// If no other non-whitespace character then we are at the message header.
				if strings.IndexFunc(prefix, isNotSpace) == -1 && strings.IndexFunc(suffix, isNotSpace) == -1 {
					fmt.Fprintf(w, "// Generated from '%s', line %d:\n", ctx.inputFile, ctx.inputLine)
					fmt.Fprintln(w, "//")
					fmt.Fprint(w, "//", line)
					msgFound = true
				}
			}
		} else {
			if strings.IndexFunc(line, isNotSpace) < msgIndent {
				break // end of the message in JSON
			}
			fmt.Fprint(w, "//", line)
		}
	}
	fmt.Fprintln(w, "//")
}

// generateMessageNameGetter generates getter for original VPP message name into the provider writer
func generateMessageNameGetter(w io.Writer, structName string, msgName string) {
	fmt.Fprintln(w, "func (*"+structName+") GetMessageName() string {")
	fmt.Fprintln(w, "\treturn \""+msgName+"\"")
	fmt.Fprintln(w, "}")
}

// generateTypeNameGetter generates getter for original VPP type name into the provider writer
func generateTypeNameGetter(w io.Writer, structName string, msgName string) {
	fmt.Fprintln(w, "func (*"+structName+") GetTypeName() string {")
	fmt.Fprintln(w, "\treturn \""+msgName+"\"")
	fmt.Fprintln(w, "}")
}

// generateMessageTypeGetter generates message factory for the generated message into the provider writer
func generateMessageTypeGetter(w io.Writer, structName string, msgType messageType) {
	fmt.Fprintln(w, "func (*"+structName+") GetMessageType() api.MessageType {")
	if msgType == requestMessage {
		fmt.Fprintln(w, "\treturn api.RequestMessage")
	} else if msgType == replyMessage {
		fmt.Fprintln(w, "\treturn api.ReplyMessage")
	} else if msgType == eventMessage {
		fmt.Fprintln(w, "\treturn api.EventMessage")
	} else {
		fmt.Fprintln(w, "\treturn api.OtherMessage")
	}
	fmt.Fprintln(w, "}")
}

// generateCrcGetter generates getter for CRC checksum of the message definition into the provider writer
func generateCrcGetter(w io.Writer, structName string, crc string) {
	crc = strings.TrimPrefix(crc, "0x")
	fmt.Fprintln(w, "func (*"+structName+") GetCrcString() string {")
	fmt.Fprintln(w, "\treturn \""+crc+"\"")
	fmt.Fprintln(w, "}")
}

// generateMessageFactory generates message factory for the generated message into the provider writer
func generateMessageFactory(w io.Writer, structName string) {
	fmt.Fprintln(w, "func New"+structName+"() api.Message {")
	fmt.Fprintln(w, "\treturn &"+structName+"{}")
	fmt.Fprintln(w, "}")
}

// translateVppType translates the VPP data type into Go data type
func translateVppType(ctx *context, vppType string, isArray bool) string {
	// basic types
	switch vppType {
	case "u8":
		if isArray {
			return "byte"
		}
		return "uint8"
	case "i8":
		return "int8"
	case "u16":
		return "uint16"
	case "i16":
		return "int16"
	case "u32":
		return "uint32"
	case "i32":
		return "int32"
	case "u64":
		return "uint64"
	case "i64":
		return "int64"
	case "f64":
		return "float64"
	}

	// typedefs
	typ, ok := ctx.types[vppType]
	if ok {
		return typ
	}

	panic(fmt.Sprintf("Unknown VPP type %s", vppType))
}

// camelCaseName returns correct name identifier (camelCase).
func camelCaseName(name string) (should string) {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

// isNotSpace returns true if the rune is NOT a whitespace character.
func isNotSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

// commonInitialisms is a set of common initialisms that need to stay in upper case.
var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"ICMP":  true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}
