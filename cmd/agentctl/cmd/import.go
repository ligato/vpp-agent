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

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/pkg/models"
)

var (
	txops uint
)

func init() {
	RootCmd.AddCommand(importConfig)
	importConfig.PersistentFlags().UintVar(&txops, "txops", 128, "Number of OPs per transaction")
}

var importConfig = &cobra.Command{
	Use:     "import",
	Aliases: []string{"i"},
	Short:   "Import configuration from file",
	Long: `
Import configuration from file.

File format:
  <key1> <value1>
  <key2> <value2>
  # ...
  <keyN> <valueN>
 
  Empty lines and lines starting with # are ignored.

Supported key formats:
  - /vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
  - config/vpp/v2/interfaces/iface1

	For short keys, the import command uses microservice label defined with --label.
`,
	Args: cobra.RangeArgs(1, 1),
	Example: `  Import configuration from file:
	$ cat input.txt
	config/vpp/v2/interfaces/loop1 {"name":"loop1","type":"SOFTWARE_LOOPBACK"}
	config/vpp/l2/v2/bridge-domain/bd1 {"name":"bd1"}
	$ agentctl import input.txt
`,
	Run: importFunction,
}

type keyVal struct {
	Key string
	Val proto.Message
}

func importFunction(cmd *cobra.Command, args []string) {
	var db keyval.ProtoBroker
	var err error

	// read file
	b, err := ioutil.ReadFile(args[0])
	if err != nil {
		utils.ExitWithError(utils.ExitError, fmt.Errorf("reading input file failed: %v", err))
		return
	}

	var keyVals []keyVal

	// parse lines
	lines := bytes.Split(b, []byte("\n"))
	for _, l := range lines {
		line := bytes.TrimSpace(l)
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}

		parts := bytes.SplitN(line, []byte(" "), 2)
		if len(parts) < 2 {
			continue
		}
		key := string(parts[0])
		data := string(parts[1])
		if key == "" || data == "" {
			continue
		}

		key, err = parseKey(key)
		if err != nil {
			utils.ExitWithError(utils.ExitError, fmt.Errorf("parsing key failed: %v", err))
			return
		}

		val, err := unmarshalKeyVal(key, data)
		if err != nil {
			utils.ExitWithError(utils.ExitError, fmt.Errorf("decoding value failed: %v", err))
			return
		}

		fmt.Printf("KEY: %s - %v\n", key, val)
		keyVals = append(keyVals, keyVal{key, val})
	}

	db, err = utils.GetDbForAllAgents(globalFlags.Endpoints)
	if err != nil {
		utils.ExitWithError(utils.ExitError, fmt.Errorf("connecting to etcd failed: %v", err))
		return
	}

	fmt.Printf("importing %d key vals\n", len(keyVals))

	var txn = db.NewTxn()
	ops := 0
	for i := 0; i < len(keyVals); i++ {
		keyVal := keyVals[i]
		txn.Put(keyVal.Key, keyVal.Val)
		fmt.Printf("PUT %s\n", keyVal.Key)
		ops++
		if ops == int(txops) || i+1 == len(keyVals) {
			fmt.Printf("commiting tx with %d ops\n", ops)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			err = txn.Commit(ctx)
			cancel()
			if err != nil {
				utils.ExitWithError(utils.ExitError, fmt.Errorf("commit failed: %v", err))
				return
			}
			ops = 0
			txn = db.NewTxn()
		}
	}

	fmt.Println("OK")
}

func parseKey(key string) (string, error) {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key, nil
	}
	if !strings.HasPrefix(key, "config/") {
		return "", fmt.Errorf("invalid format for key: %q", key)
	}
	return path.Join(servicelabel.GetAllAgentsPrefix(), globalFlags.Label, key), nil
}

func unmarshalKeyVal(fullKey string, data string) (proto.Message, error) {
	keyParts := strings.Split(fullKey, "/")
	if len(keyParts) < 4 || keyParts[0] != "" {
		return nil, fmt.Errorf("invalid key: %q", fullKey)
	}
	key := path.Join(keyParts[3:]...)

	model, err := models.GetModelForKey(key)
	if err != nil {
		return nil, err
	}
	valueType := proto.MessageType(model.ProtoName())
	if valueType == nil {
		return nil, fmt.Errorf("unknown proto message defined for key %s", key)
	}
	value := reflect.New(valueType.Elem()).Interface().(proto.Message)

	if err := jsonpb.UnmarshalString(data, value); err != nil {
		return nil, err
	}
	return value, nil
}

type lazyProto struct {
	val proto.Message
}

// GetValue returns the value of the pair.
func (lazy *lazyProto) GetValue(out proto.Message) error {
	if lazy.val != nil {
		proto.Merge(out, lazy.val)
	}
	return nil
}
