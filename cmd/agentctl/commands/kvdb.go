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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewKvdbCommand(cli agentcli.Cli) *cobra.Command {
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

func newKvdbListCommand(cli agentcli.Cli) *cobra.Command {
	var keysOnly bool

	cmd := &cobra.Command{
		Use:     "list [flags] [PREFIX]",
		Aliases: []string{"ls", "l"},
		Short:   "List key-value entries",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string
			if len(args) > 0 {
				key = args[0]
			}
			return runKvdbList(cli, key, keysOnly)
		},
	}
	cmd.Flags().BoolVar(&keysOnly, "keys-only", false, "List only the keys")
	return cmd
}

func runKvdbList(cli agentcli.Cli, key string, keysOnly bool) error {
	logrus.Debugf("kvdb.List - KEY: %q", key)

	kvdb, err := cli.Client().KVDBClient()
	if err != nil {
		return fmt.Errorf("connecting to KVDB failed: %v", err)
	}
	if keysOnly {
		iter, err := kvdb.ListKeys(key)
		if err != nil {
			return errors.New("failed to list values from KVDB: " + err.Error())
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
			return errors.New("failed to list values from KVDB: " + err.Error())
		}
		for {
			kv, stop := iter.GetNext()
			if stop {
				break
			}
			fmt.Printf("%s\n%s\n", kv.GetKey(), kv.GetValue())
		}
	}
	return nil
}

func newKvdbGetCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get KEY",
		Aliases: []string{"g"},
		Short:   "Get key-value entry",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			return runKvdbGet(cli, key)
		},
	}
	return cmd
}

func runKvdbGet(cli agentcli.Cli, key string) error {
	logrus.Debugf("kvdb.Get - KEY: %q", key)

	kvdb, err := cli.Client().KVDBClient()
	if err != nil {
		return fmt.Errorf("connecting to KVDB failed: %v", err)
	}

	value, found, _, err := kvdb.GetValue(key)
	if err != nil {
		return errors.New("Failed to get value from Etcd: " + err.Error())
	} else if !found {
		return errors.New("key not found")
	}

	fmt.Fprintf(cli.Out(), "%s\n", value)
	return nil
}

func newKvdbPutCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put KEY VALUE",
		Aliases: []string{"p"},
		Short:   "Put key-value entry",
		Long: `
Put configuration file to Etcd.

Supported key formats:
	/vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
	config/vpp/v2/interfaces/iface1

For short key, put command use default microservice label.
`,
		Example: `  Set route configuration for "vpp1":
	$ {{.CommandPath}} /vnf-agent/vpp1/config/vpp/v2/route/vrf/1/dst/10.1.1.3/32/gw/192.168.1.13 '{
	"type": 1,
	"vrf_id": 1,
	"dst_network": "10.1.1.3/32",
	"next_hop_addr": "192.168.1.13"
}'
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			val := args[1]
			return runKvdbPut(cli, key, val)
		},
	}
	return cmd
}

func runKvdbPut(cli agentcli.Cli, key, value string) error {
	logrus.Debugf("kvdb.Put - KEY: %q VAL: %q", key, value)

	kvdb, err := cli.Client().KVDBClient()
	if err != nil {
		return fmt.Errorf("connecting to KVDB failed: %v", err)
	}

	data := []byte(value)
	if err := kvdb.Put(key, data); err != nil {
		return err
	}

	fmt.Fprintln(cli.Out(), "OK")
	return nil
}

func newKvdbDelCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del KEY",
		Aliases: []string{"d"},
		Short:   "Delete key-value entry",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			return runKvdbDel(cli, key)
		},
	}
	return cmd
}

func runKvdbDel(cli agentcli.Cli, key string) error {
	logrus.Debugf("kvdb.Del - KEY: %q", key)

	kvdb, err := cli.Client().KVDBClient()
	if err != nil {
		return fmt.Errorf("connecting to KVDB failed: %v", err)
	}

	if _, found, _, err := kvdb.GetValue(key); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("key does not exist")
	}

	// FIXME: existed can never be true, missing WithPrevKV() option
	if _, err := kvdb.Delete(key); err != nil {
		return err
	}

	fmt.Fprintln(cli.Out(), "OK")
	return nil
}
