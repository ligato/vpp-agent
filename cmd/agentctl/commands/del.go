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
	"errors"
	"strings"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

func delCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del <key>",
		Aliases: []string{"d"},
		Short:   "Delete configuration file",
		Long: `
	Delete configuration file
`,
		Args: cobra.RangeArgs(1, 1),
		Run:  delFunction,
	}
	return cmd
}

func delFunction(cmd *cobra.Command, args []string) {
	var db keyval.ProtoBroker
	var err error

	key := args[0]

	if !strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		tmp := strings.Split(key, "/")
		if tmp[0] != "config" {
			globalFlags.ServiceLabel = tmp[0]
			key = strings.Join(tmp[1:], "/")
		}

		db, err = utils.GetDbForOneAgent(globalFlags.Endpoints, globalFlags.ServiceLabel)
		if err != nil {
			utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
		}
	} else {
		db, err = utils.GetDbForAllAgents(globalFlags.Endpoints)
		if err != nil {
			utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
		}
	}

	utils.DelDataFromDb(db.NewTxn(), key)
}
