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

	"github.com/spf13/cobra"
)

func NewKvdbCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kvdb",
		Aliases: []string{"kv"},
		Short:   "Manage agent data in KVDB",
		Args:    cobra.NoArgs,
	}
	cmd.AddCommand(
		newKvdbListCommand(cli),
		newKvdbGetCommand(cli),
		newKvdbPutCommand(cli),
		newKvdbDelCommand(cli),
	)
	return cmd
}

func newKvdbListCommand(cli *AgentCli) *cobra.Command {
	var keysOnly bool

	cmd := &cobra.Command{
		Use:     "list [PREFIX]",
		Aliases: []string{"l"},
		Short:   "List key-value entries from KVDB",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var key string
			if len(args) > 0 {
				key = args[0]
			}
			runKvdbList(cli, key, keysOnly)
		},
	}
	cmd.Flags().BoolVar(&keysOnly, "keys-only", false, "List only the keys")
	return cmd
}

func runKvdbList(cli *AgentCli, key string, keysOnly bool) {
	Debugf("kvdb.List- KEY: %s\n", key)

	kvdb := cli.NewKVDBClient()
	if keysOnly {
		iter, err := kvdb.ListKeys(key)
		if err != nil {
			ExitWithError(errors.New("Failed to list values from Etcd: " + err.Error()))
			return
		}
		for {
			key, _, stop := iter.GetNext()
			if stop {
				break
			}
			fmt.Printf("%s\n", key)
		}
	} else {
		iter, err := kvdb.ListValues(key)
		if err != nil {
			ExitWithError(errors.New("Failed to list values from Etcd: " + err.Error()))
			return
		}
		for {
			kv, stop := iter.GetNext()
			if stop {
				break
			}
			fmt.Printf("%s\n%s\n", kv.GetKey(), kv.GetValue())
		}
	}
}

func newKvdbGetCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get KEY",
		Aliases: []string{"g"},
		Short:   "Get key-value entry from KVDB",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			runKvdbGet(cli, key)
		},
	}
	return cmd
}

func runKvdbGet(cli *AgentCli, key string) {
	Debugf("kvdb.Get - KEY: %s\n", key)

	kvdb := cli.NewKVDBClient()

	value, found, _, err := kvdb.GetValue(key)
	if err != nil {
		ExitWithError(errors.New("Failed to get value from Etcd: " + err.Error()))
		return
	} else if !found {
		ExitWithError(errors.New("key not found"))
		return
	}

	fmt.Printf("%s\n", value)
}

func newKvdbPutCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put KEY VALUE",
		Aliases: []string{"p"},
		Short:   "Put key-value entry into KVDB",
		Long: `
Put configuration file to Etcd.

Supported key formats:
	/vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
	config/vpp/v2/interfaces/iface1

For short key, put command use default microservice label.
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
			runKvdbPut(cli, key, val)
		},
	}
	return cmd
}

func runKvdbPut(cli *AgentCli, key, value string) {
	Debugf("kvdb.Put - KEY: %s VAL: %s\n", key, value)

	kvdb := cli.NewKVDBClient()

	data := []byte(value)
	if err := kvdb.Put(key, data); err != nil {
		ExitWithError(err)
	}

	fmt.Println("OK")
}

func newKvdbDelCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del KEY",
		Aliases: []string{"d"},
		Short:   "Delete key-value entry from KVDB",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			runKvdbDel(cli, key)
		},
	}
	return cmd
}

func runKvdbDel(cli *AgentCli, key string) {
	Debugf("kvdb.Del - KEY: %s \n", key)

	kvdb := cli.NewKVDBClient()

	if _, found, _, err := kvdb.GetValue(key); err != nil {
		ExitWithError(err)
	} else if !found {
		ExitWithError(fmt.Errorf("key does not exist"))
	}

	_, err := kvdb.Delete(key) // FIXME: existed can never be true, missing WithPrevKV() option
	if err != nil {
		ExitWithError(err)
	}

	fmt.Println("OK")
}
