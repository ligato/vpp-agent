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
	"fmt"
	"strings"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

func NewConfigCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage agent config data",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newConfigGetCommand(cli),
		newConfigPutCommand(cli),
		newConfigDelCommand(cli),
	)
	return cmd
}

func newConfigGetCommand(cli *AgentCli) *cobra.Command {
	var (
		prefix bool
	)
	cmd := &cobra.Command{
		Use:     "get <key>",
		Aliases: []string{"g"},
		Short:   "Get config entry from Etcd",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			runConfigGet(cli, prefix, key)
		},
	}
	cmd.Flags().BoolVarP(&prefix, "prefix", "p", false, "Get keys with matching prefix")
	return cmd
}

func runConfigGet(cli *AgentCli, prefix bool, key string) {
	if prefix {
		iter, err := cli.KVDBClient().ListValues(key)
		if err != nil {
			utils.ExitWithError(utils.ExitError, errors.New("Failed to list values from Etcd: "+err.Error()))
			return
		}
		for {
			kv, stop := iter.GetNext()
			if stop {
				break
			}
			fmt.Printf("%s\n%s\n", kv.GetKey(), kv.GetValue())
		}
	} else {
		value, found, _, err := cli.KVDBClient().GetValue(key)
		if err != nil {
			utils.ExitWithError(utils.ExitError, errors.New("Failed to get value from Etcd: "+err.Error()))
			return
		} else if !found {
			utils.ExitWithError(utils.ExitNotFound, errors.New("key not found"))
			return
		}
		fmt.Printf("%s\n", value)
	}
}

func newConfigPutCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put <key> <value>",
		Aliases: []string{"p"},
		Short:   "Put config entry into Etcd",
		Long: `
Put configuration file to Etcd.

Supported key formats:
	/vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
	vpp1/config/vpp/v2/interfaces/iface1
	config/vpp/v2/interfaces/iface1

For short key, put command use default microservice label and 'vpp1' as default agent label.
`,
		Example: `  Set route configuration for "vpp1":
	$ agentctl -e 172.17.0.3:2379 config put /vnf-agent/vpp1/config/vpp/v2/route/vrf/1/dst/10.1.1.3/32/gw/192.168.1.13 '{
	"type": 1,
	"vrf_id": 1,
	"dst_network": "10.1.1.3/32",
	"next_hop_addr": "192.168.1.13"
}'
`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			val := args[1]
			runConfigPut(cli, key, val)
		},
	}
	return cmd
}

func runConfigPut(cli *AgentCli, key, value string) {
	var db keyval.ProtoBroker
	var err error

	Debugf("PUT: %s\n%s\n", key, value)

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

	utils.WriteData(db.NewTxn(), key, value)

	fmt.Println("Ok")
}

func newConfigDelCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del <key>",
		Aliases: []string{"d"},
		Short:   "Delete config entry from Etcd",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			runConfigDel(cli, key)
		},
	}
	return cmd
}

func runConfigDel(cli *AgentCli, key string) {
	var db keyval.ProtoBroker
	var err error

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

	err = utils.DelDataFromDb(db.NewTxn(), key)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to delete from Etcd: "+err.Error()))
	}
}
