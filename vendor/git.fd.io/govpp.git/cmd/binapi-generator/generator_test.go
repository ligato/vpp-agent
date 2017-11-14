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
	"os"
	"testing"

	"github.com/bennyscetbun/jsongo"
	. "github.com/onsi/gomega"
)

func TestGetInputFiles(t *testing.T) {
	RegisterTestingT(t)
	result, err := getInputFiles("testdata")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).To(HaveLen(5))
	for _, file := range result {
		Expect(file).To(BeAnExistingFile())
	}
}

func TestGetInputFilesError(t *testing.T) {
	RegisterTestingT(t)
	result, err := getInputFiles("nonexisting_directory")
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
}

func TestGenerateFromFile(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	// remove directory created during test
	defer os.RemoveAll(outDir)
	err := generateFromFile("testdata/acl.api.json", outDir)
	Expect(err).ShouldNot(HaveOccurred())
	fileInfo, err := os.Stat(outDir + "/acl/acl.go")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(fileInfo.IsDir()).To(BeFalse())
	Expect(fileInfo.Name()).To(BeEquivalentTo("acl.go"))
}

func TestGenerateFromFileInputError(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	err := generateFromFile("testdata/nonexisting.json", outDir)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("reading data from file failed"))
}

func TestGenerateFromFileReadJsonError(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	err := generateFromFile("testdata/input-read-json-error.json", outDir)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("JSON unmarshall failed"))
}

func TestGenerateFromFileGeneratePackageError(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	// generate package throws panic, recover after it
	defer func() {
		if recovery := recover(); recovery != nil {
			t.Logf("Recovered from panic: %v", recovery)
		}
		os.RemoveAll(outDir)
	}()
	err := generateFromFile("testdata/input-generate-error.json", outDir)
	Expect(err).Should(HaveOccurred())
}

func TestGetContext(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	result, err := getContext("testdata/af_packet.api.json", outDir)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())
	Expect(result.outputFile).To(BeEquivalentTo(outDir + "/af_packet/af_packet.go"))
}

func TestGetContextNoJsonFile(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	result, err := getContext("testdata/input.txt", outDir)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid input file name"))
	Expect(result).To(BeNil())
}

func TestGetContextInterfaceJson(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	result, err := getContext("testdata/interface.json", outDir)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())
	Expect(result.outputFile)
	Expect(result.outputFile).To(BeEquivalentTo(outDir + "/interfaces/interfaces.go"))

}

func TestReadJson(t *testing.T) {
	RegisterTestingT(t)
	inputData, err := readFile("testdata/af_packet.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	result, err := parseJSON(inputData)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())
	Expect(result.Len()).To(BeEquivalentTo(3))
}

func TestReadJsonError(t *testing.T) {
	RegisterTestingT(t)
	inputData, err := readFile("testdata/input-read-json-error.json")
	Expect(err).ShouldNot(HaveOccurred())
	result, err := parseJSON(inputData)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("JSON unmarshall failed"))
	Expect(result).To(BeNil())
}

func TestGeneratePackage(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/ip.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	testCtx.inputBuff = bytes.NewBuffer(inputData)
	inFile, _ := parseJSON(inputData)
	outDir := "test_output_directory"
	outFile, _ := os.Create(outDir)
	defer os.RemoveAll(outDir)

	// prepare writer
	writer := bufio.NewWriter(outFile)
	Expect(writer.Buffered()).To(BeZero())
	err = generatePackage(testCtx, writer, inFile)
	Expect(err).ShouldNot(HaveOccurred())
}

func TestGenerateMessageType(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/ip.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	testCtx.inputBuff = bytes.NewBuffer(inputData)
	inFile, _ := parseJSON(inputData)
	outDir := "test_output_directory"
	outFile, _ := os.Create(outDir)
	defer os.RemoveAll(outDir)

	// prepare writer
	writer := bufio.NewWriter(outFile)

	types := inFile.Map("types")
	testCtx.types = map[string]string{
		"u32": "sw_if_index",
		"u8":  "weight",
	}
	Expect(types.Len()).To(BeEquivalentTo(1))
	for i := 0; i < types.Len(); i++ {
		typ := types.At(i)
		Expect(writer.Buffered()).To(BeZero())
		err := generateMessage(testCtx, writer, typ, true)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(writer.Buffered()).ToNot(BeZero())

	}
}

func TestGenerateMessageName(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/ip.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	testCtx.inputBuff = bytes.NewBuffer(inputData)
	inFile, _ := parseJSON(inputData)
	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	defer os.RemoveAll(outDir)

	// prepare writer
	writer := bufio.NewWriter(outFile)

	types := inFile.Map("types")
	Expect(types.Len()).To(BeEquivalentTo(1))
	for i := 0; i < types.Len(); i++ {
		typ := types.At(i)
		Expect(writer.Buffered()).To(BeZero())
		err := generateMessage(testCtx, writer, typ, false)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(writer.Buffered()).ToNot(BeZero())

	}
}

func TestGenerateMessageFieldTypes(t *testing.T) {
	// expected results according to acl.api.json in testdata
	expectedTypes := []string{"\tIsPermit uint8", "\tIsIpv6 uint8", "\tSrcIPAddr []byte	`struc:\"[16]byte\"`",
		"\tSrcIPPrefixLen uint8", "\tDstIPAddr []byte	`struc:\"[16]byte\"`", "\tDstIPPrefixLen uint8", "\tProto uint8",
		"\tSrcportOrIcmptypeFirst uint16", "\tSrcportOrIcmptypeLast uint16", "\tDstportOrIcmpcodeFirst uint16",
		"\tDstportOrIcmpcodeLast uint16", "\tTCPFlagsMask uint8", "\tTCPFlagsValue uint8"}
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/acl.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	inFile, _ := parseJSON(inputData)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(inFile).ToNot(BeNil())

	// test types
	types := inFile.Map("types")
	fields := make([]string, 0)
	for i := 0; i < types.Len(); i++ {
		for j := 0; j < types.At(i).Len(); j++ {
			field := types.At(i).At(j)
			if jsongo.TypeArray == field.GetType() {
				err := processMessageField(testCtx, &fields, field, otherMessage)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(fields[j-1]).To(BeEquivalentTo(expectedTypes[j-1]))
			}
		}
	}
}

func TestGenerateMessageFieldMessages(t *testing.T) {
	// expected results according to acl.api.json in testdata
	expectedTypes := []string{"\tMajor uint32", "\tMinor uint32", "\tACLIndex uint32",
		"\tTag []byte	`struc:\"[64]byte\"`", "\tACLIndex uint32", "\tRetval int32", "\tACLIndex uint32"}
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/acl.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	inFile, err := parseJSON(inputData)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(inFile).ToNot(BeNil())

	// test types
	messages := inFile.Map("messages")
	customIndex := 0
	fields := make([]string, 0)
	for i := 0; i < messages.Len(); i++ {
		for j := 0; j < messages.At(i).Len(); j++ {
			field := messages.At(i).At(j)
			if jsongo.TypeArray == field.GetType() {
				specificFieldName := field.At(1).Get().(string)
				if specificFieldName == "crc" || specificFieldName == "_vl_msg_id" ||
					specificFieldName == "client_index" || specificFieldName == "context" {
					continue
				}
				err := processMessageField(testCtx, &fields, field, requestMessage)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(fields[customIndex]).To(BeEquivalentTo(expectedTypes[customIndex]))
				customIndex++
			}
		}
	}
}

func TestGeneratePackageHeader(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"

	// prepare input/output output files
	inputData, err := readFile("testdata/acl.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	inFile, err := parseJSON(inputData)
	Expect(err).ShouldNot(HaveOccurred())
	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	defer os.RemoveAll(outDir)
	// prepare writer
	writer := bufio.NewWriter(outFile)
	Expect(writer.Buffered()).To(BeZero())
	generatePackageHeader(testCtx, writer, inFile)
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestGenerateMessageCommentType(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"
	testCtx.inputBuff = bytes.NewBuffer([]byte("test content"))

	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	writer := bufio.NewWriter(outFile)
	defer os.RemoveAll(outDir)
	Expect(writer.Buffered()).To(BeZero())
	generateMessageComment(testCtx, writer, "test-struct", "msg-name", true)
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestGenerateMessageCommentMessage(t *testing.T) {
	RegisterTestingT(t)
	// prepare context
	testCtx := new(context)
	testCtx.packageName = "test-package-name"
	testCtx.inputBuff = bytes.NewBuffer([]byte("test content"))

	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	writer := bufio.NewWriter(outFile)
	defer os.RemoveAll(outDir)
	Expect(writer.Buffered()).To(BeZero())
	generateMessageComment(testCtx, writer, "test-struct", "msg-name", false)
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestGenerateMessageNameGetter(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	writer := bufio.NewWriter(outFile)
	defer os.RemoveAll(outDir)
	Expect(writer.Buffered()).To(BeZero())
	generateMessageNameGetter(writer, "test-struct", "msg-name")
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestGenerateTypeNameGetter(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	writer := bufio.NewWriter(outFile)
	defer os.RemoveAll(outDir)
	Expect(writer.Buffered()).To(BeZero())
	generateTypeNameGetter(writer, "test-struct", "msg-name")
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestGenerateCrcGetter(t *testing.T) {
	RegisterTestingT(t)
	outDir := "test_output_directory"
	outFile, err := os.Create(outDir)
	Expect(err).ShouldNot(HaveOccurred())
	writer := bufio.NewWriter(outFile)
	defer os.RemoveAll(outDir)
	Expect(writer.Buffered()).To(BeZero())
	generateCrcGetter(writer, "test-struct", "msg-name")
	Expect(writer.Buffered()).ToNot(BeZero())
}

func TestTranslateVppType(t *testing.T) {
	RegisterTestingT(t)
	context := new(context)
	typesToTranslate := []string{"u8", "i8", "u16", "i16", "u32", "i32", "u64", "i64", "f64"}
	expected := []string{"uint8", "int8", "uint16", "int16", "uint32", "int32", "uint64", "int64", "float64"}
	translated := []string{}
	for _, value := range typesToTranslate {
		translated = append(translated, translateVppType(context, value, false))
	}
	for index, value := range expected {
		Expect(value).To(BeEquivalentTo(translated[index]))
	}

}

func TestTranslateVppTypeArray(t *testing.T) {
	RegisterTestingT(t)
	context := new(context)
	translated := translateVppType(context, "u8", true)
	Expect(translated).To(BeEquivalentTo("byte"))
}

func TestTranslateVppUnknownType(t *testing.T) {
	defer func() {
		if recovery := recover(); recovery != nil {
			t.Logf("Recovered from panic: %v", recovery)
		}
	}()
	context := new(context)
	translateVppType(context, "?", false)
}

func TestCamelCase(t *testing.T) {
	RegisterTestingT(t)
	// test camel case functionality
	expected := "allYourBaseAreBelongToUs"
	result := camelCaseName("all_your_base_are_belong_to_us")
	Expect(expected).To(BeEquivalentTo(result))
	// test underscore
	expected = "_"
	result = camelCaseName(expected)
	Expect(expected).To(BeEquivalentTo(result))
	// test all lower
	expected = "lower"
	result = camelCaseName(expected)
	Expect(expected).To(BeEquivalentTo(result))
}

func TestCommonInitialisms(t *testing.T) {
	RegisterTestingT(t)

	for key, value := range commonInitialisms {
		Expect(value).ShouldNot(BeFalse())
		Expect(key).ShouldNot(BeEmpty())
	}
}
