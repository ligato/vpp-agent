// Copyright (c) 2019 Cisco and/or its affiliates.
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
	"context"
	"fmt"
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging/logrus"
	"log"
	"os"
	"strconv"
	"time"
)

// processArgs processes input arguments.
func processArgs() (cfg *etcd.ClientConfig, leadershipTime int64, err error) {
	// default args
	fileConfig := &etcd.Config{}

	if len(os.Args) > 2 && os.Args[1] == "--cfg" {
		err = config.ParseConfigFromYamlFile(os.Args[2], fileConfig)
		if err != nil {
			return
		}
		cfg, err = etcd.ConfigToClient(fileConfig)
		if err != nil {
			return
		}

		if len(os.Args) > 3 {
			seconds, err := strconv.ParseUint(os.Args[3], 10, 32)
			if err != nil {
				return cfg, 0, err
			}
			return cfg, int64(seconds), nil
		}
	} else {
		return cfg, 0, fmt.Errorf("incorrect arguments")
	}

	return cfg, leadershipTime, nil
}

func printUsage() {
	fmt.Printf("\n%s: --cfg CONFIG_FILE [SECONDS_TO_HOLD_LEADERSHIP]\n\n", os.Args[0])
}

func main() {
	cfg, leadershipTime, err := processArgs()
	if err != nil {
		printUsage()
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := etcd.NewEtcdConnectionWithBytes(*cfg, logrus.DefaultLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// start campaign on the given prefix
	resign, err := db.CampaignInElection(context.Background(), "/election")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Elected as leader")

	db.Put("/somewhere", []byte(time.Now().String()), datasync.WithClientLifetimeTTL())

	// sleep until the gained leadership is released
	time.Sleep(time.Second * time.Duration(leadershipTime))

	// calling resign is optional, without this call the new election will be trigger after sessionTTL automatically
	resign(context.Background())

}
