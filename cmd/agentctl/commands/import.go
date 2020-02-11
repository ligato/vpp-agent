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
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

func NewImportCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts    ImportOptions
		timeout uint
	)
	cmd := &cobra.Command{
		Use:   "import file",
		Args:  cobra.ExactArgs(1),
		Short: "Import config data from file",
		Example: `
 To import file contents into Etcd, run:
  $ cat input.txt
  config/vpp/v2/interfaces/loop1 {"name":"loop1","type":"SOFTWARE_LOOPBACK"}
  config/vpp/l2/v2/bridge-domain/bd1 {"name":"bd1"}
  
  $ {{.CommandPath}} input.txt

 To import it via gRPC, include --grpc flag:
  $ {{.CommandPath}} --grpc=localhost:9111 input.txt

 FILE FORMAT
    Contents of the import file must contain single key-value pair per line:

    <key1> <value1>
    <key2> <value2>
    ...
    <keyN> <valueN>

    Empty lines and lines starting with '#' are ignored.

 KEY FORMAT
    Keys can be defined in two ways:

    - full: 	/vnf-agent/vpp1/config/vpp/v2/interfaces/iface1
    - short:	config/vpp/v2/interfaces/iface1
 
    For short keys, the import command uses microservice label defined with --service-label.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InputFile = args[0]
			opts.Timeout = time.Second * time.Duration(timeout)
			return RunImport(cli, opts)
		},
	}

	flags := cmd.Flags()
	flags.UintVar(&opts.TxOps, "txops", 128, "Number of ops per transaction")
	flags.UintVarP(&timeout, "time", "t", 30, "Timeout (in seconds) to wait for server response")
	flags.BoolVar(&opts.ViaGrpc, "grpc", false, "Enable to import config via gRPC")

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

	if opts.ViaGrpc {
		// Set up a connection to the server.
		c, err := cli.Client().ConfigClient()
		if err != nil {
			return err
		}

		fmt.Printf("importing %d key vals\n", len(keyVals))

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

		fmt.Printf("importing %d key vals\n", len(keyVals))

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

		//key = completeFullKey(key)

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
