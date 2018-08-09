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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

var logger logging.Logger

func init() {
	logger = logrus.DefaultLogger()
	logger.SetLevel(logging.DebugLevel)
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			logger.WithFields(logging.Fields{
				"omg":    true,
				"err":    err,
				"number": 100,
			}).Fatal("The ice breaks!")
		}
	}()

	logger.WithFields(logging.Fields{
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	logger.WithFields(logging.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	logger.WithFields(logging.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	logger.WithFields(logging.Fields{
		"temperature": -4,
	}).Debug("Temperature changes")

	logger.WithFields(logging.Fields{
		"animal": "orca",
		"size":   9009,
	}).Panic("It's over 9000!")
}
