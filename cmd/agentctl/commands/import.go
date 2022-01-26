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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

const (
	defaultTimeout = time.Second * 30
	defaultTxOps   = 128
)

func NewImportCommand(cli agentcli.Cli) *cobra.Command {
	var opts ImportOptions
	cmd := &cobra.Command{
		Use:   "import FILE",
		Args:  cobra.ExactArgs(1),
		Short: "Import config data from file",
		Long: `Import config data from file into Etcd or via gRPC. 
FILE FORMAT
  Contents of the import file must contain single key-value pair per line:

	<key1> <value1>
    <key2> <value2>
    ...
    <keyN> <valueN>

    NOTE: Empty lines and lines starting with '#' are ignored.

  Sample file:
  	config/vpp/v2/interfaces/loop1 {"name":"loop1","type":"SOFTWARE_LOOPBACK"}
  	config/vpp/l2/v2/bridge-domain/bd1 {"name":"bd1"}

KEY FORMAT
  Keys can be defined in two ways:
  
    - Full  - /vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
    - Short - config/vpp/v2/interfaces/iface1
 
  When using short keys, import will use configured microservice label (e.g. --service-label flag).`,
		Example: `
# Import data into Etcd
{{.CommandPath}} input.txt

# Import data directly into agent via gRPC
{{.CommandPath}} --grpc input.txt
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InputFile = args[0]
			return RunImport(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.UintVar(&opts.TxOps, "txops", defaultTxOps, "Number of ops per transaction")
	flags.DurationVarP(&opts.Timeout, "time", "t", defaultTimeout, "Timeout to wait for server response")
	flags.BoolVar(&opts.ViaGrpc, "grpc", false, "Import config directly to agent via gRPC")
	return cmd
}

type ImportOptions struct {
	InputFile string
	TxOps     uint
	Timeout   time.Duration
	ViaGrpc   bool
}

func RunImport(cli agentcli.Cli, opts ImportOptions) error {
	keyVals, err := parseImportFile(opts.InputFile)
	if err != nil {
		return fmt.Errorf("parsing import data failed: %v", err)
	}
	fmt.Printf("importing %d key-value pairs\n", len(keyVals))

	if opts.ViaGrpc {
		// Set up a connection to the server.
		c, err := cli.Client().GenericClient()
		if err != nil {
			return err
		}
		req := c.ChangeRequest()
		for _, keyVal := range keyVals {
			fmt.Printf(" - %s\n", keyVal.Key)
			req.Update(keyVal.Val)
		}
		fmt.Printf("sending via gRPC\n")

		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()

		if err := req.Send(ctx); err != nil {
			return fmt.Errorf("send failed: %v", err)
		}
	} else {
		c, err := cli.Client().KVDBClient()
		if err != nil {
			return fmt.Errorf("KVDB error: %v", err)
		}
		db := c.ProtoBroker()
		var txn = db.NewTxn()
		ops := 0
		for i := 0; i < len(keyVals); i++ {
			keyVal := keyVals[i]
			key, err := c.CompleteFullKey(keyVal.Key)
			if err != nil {
				return fmt.Errorf("key processing failed: %v", err)
			}
			fmt.Printf(" - %s\n", key)
			txn.Put(key, keyVal.Val)
			ops++

			if ops == int(opts.TxOps) || i+1 == len(keyVals) {
				fmt.Printf("commiting tx with %d ops\n", ops)
				ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
				err = txn.Commit(ctx)
				cancel()
				if err != nil {
					return fmt.Errorf("commit failed: %v", err)
				}
				ops = 0
				txn = db.NewTxn()
			}
		}
	}

	logging.Debug("OK")
	return nil
}

type keyVal struct {
	Key string
	Val proto.Message
}

func parseImportFile(importFile string) (keyVals []keyVal, err error) {
	b, err := ioutil.ReadFile(importFile)
	if err != nil {
		return nil, fmt.Errorf("reading input file failed: %v", err)
	}
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
		logrus.Debugf("parse line: %s %s\n", key, data)

		// key = completeFullKey(key)
		val, err := unmarshalKeyVal(key, data)
		if err != nil {
			return nil, fmt.Errorf("decoding value failed: %v", err)
		}
		logrus.Debugf("KEY: %s - %v\n", key, val)
		keyVals = append(keyVals, keyVal{key, val})
	}
	return
}

func unmarshalKeyVal(fullKey string, data string) (proto.Message, error) {
	key := stripAgentPrefix(fullKey)

	model, err := models.GetModelForKey(key)
	if err != nil {
		return nil, err
	}
	valueType := protoMessageType(model.ProtoName())
	if valueType == nil {
		return nil, fmt.Errorf("unknown proto message defined for: %s", model.ProtoName())
	}
	value := valueType.New().Interface()
	if err = protojson.Unmarshal([]byte(data), value); err != nil {
		return nil, err
	}
	return value, nil
}

func stripAgentPrefix(key string) string {
	if !strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key
	}
	keyParts := strings.Split(key, "/")
	if len(keyParts) < 4 || keyParts[0] != "" {
		return path.Join(keyParts[2:]...)
	}
	return path.Join(keyParts[3:]...)
}
