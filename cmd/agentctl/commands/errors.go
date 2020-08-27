//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package commands

import (
	"fmt"
	"strings"
)

// Errors is a list of errors.
// Useful in a loop if you don't want to return the error right away and you want to display after the loop,
// all the errors that happened during the loop.
type Errors []error

func (errList Errors) Error() string {
	if len(errList) == 0 {
		return ""
	}
	out := make([]string, len(errList))
	for i := range errList {
		out[i] = errList[i].Error()
	}
	return strings.Join(out, ", ")
}

// StatusError reports an unsuccessful exit by a command.
type StatusError struct {
	Status     string
	StatusCode int
}

func (e StatusError) String() string {
	return fmt.Sprintf("Status: %s, Code: %d", e.Status, e.StatusCode)
}

func (e StatusError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Status, e.StatusCode)
}

// ExitCode returns proper exit code for err or 0 if err is nil.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if sterr, ok := err.(StatusError); ok {
		// StatusError should only be used for errors, and all errors should
		// have a non-zero exit status, so never exit with 0
		if sterr.StatusCode != 0 {
			return sterr.StatusCode
		}
	}
	return 1
}
