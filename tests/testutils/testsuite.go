//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package testutils

import (
	"log"
	"os"
	"strings"
)

func TestSuite() string {
	return os.Getenv("TEST_SUITE")
}

func IsRunningTestSuite(suite string) bool {
	return strings.EqualFold(TestSuite(), suite)
}

func SetTestSuite(suite string) {
	if err := os.Setenv("TEST_SUITE", suite); err != nil {
		log.Panicf("error setting TEST_SUITE=%q: %v", suite, err)
	}
}

func RunTestSuite(suite string) bool {
	if !IsRunningTestSuite(suite) {
		log.Printf("skip running test suite: %[1]s (set TEST_SUITE=%[1]s to run)", suite)
		return false
	}
	return true
}
