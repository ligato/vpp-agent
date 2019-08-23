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

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/client/remoteclient"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/pkg/models"
)

var (
	txops    uint
	grpcAddr string
	timeout  uint
)

func NewImportCommand(cli *AgentCli) *cobra.Command {
	opts := ImportOptions{}

	cmd := &cobra.Command{
		Use:     "import [file]",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Import config data from file",
		Example: `
 Import file contents example:
  $ cat input.txt
  config/vpp/v2/interfaces/loop1 {"name":"loop1","type":"SOFTWARE_LOOPBACK"}
  config/vpp/l2/v2/bridge-domain/bd1 {"name":"bd1"}

 To import it into Etcd, run:
  $ agentctl import input.txt

 To import it via gRPC, use --grpc flag:
  $ agentctl import --grpc=localhost:9111 input.txt

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
 
 For short keys, the import command uses microservice label defined with --label.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Filename = args[0]
			return RunImport(cli, opts)
		},
	}

	cmd.PersistentFlags().UintVar(&txops, "txops", 128,
		"Number of OPs per transaction")
	cmd.PersistentFlags().StringVar(&grpcAddr, "grpc", "",
		"Address of gRPC server.")
	cmd.PersistentFlags().UintVarP(&timeout, "time", "t", 60,
		"Client timeout in seconds (how long to wait for response from server)")

	return cmd
}

type ImportOptions struct {
	Filename string
}

func RunImport(cli *AgentCli, opts ImportOptions) error {
	b, err := ioutil.ReadFile(opts.Filename)
	if err != nil {
		return fmt.Errorf("reading input file failed: %v", err)
	}

	keyVals, err := parseKeyVals(b)
	if err != nil {
		return fmt.Errorf("parsing import input failed: %v", err)
	}

	if grpcAddr != "" {
		if err := grpcImport(keyVals); err != nil {
			return fmt.Errorf("import via gRPC failed: %v", err)
		}
	} else {
		if err := etcdImport(keyVals); err != nil {
			return fmt.Errorf("import to Etcd failed: %v", err)
		}
	}

	logging.Debug("OK")
	return nil
}

type keyVal struct {
	Key string
	Val proto.Message
}

func parseKeyVals(b []byte) (keyVals []keyVal, err error) {
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
			return nil, fmt.Errorf("parsing key failed: %v", err)
		}

		val, err := unmarshalKeyVal(key, data)
		if err != nil {
			return nil, fmt.Errorf("decoding value failed: %v", err)
		}

		Debugf("KEY: %s - %v\n", key, val)
		keyVals = append(keyVals, keyVal{key, val})
	}
	return
}

func grpcImport(keyVals []keyVal) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("connecting to gRPC failed: %v", err)
	}
	defer conn.Close()

	c := remoteclient.NewClientGRPC(genericmanager.NewGenericManagerClient(conn))

	fmt.Printf("importing %d key vals\n", len(keyVals))

	req := c.ChangeRequest()
	for _, keyVal := range keyVals {
		fmt.Printf(" - %s\n", keyVal.Key)
		req.Update(keyVal.Val)
	}

	t := getTimeout()
	fmt.Printf("sending via gRPC (timeout: %v)\n", t)

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	if err := req.Send(ctx); err != nil {
		return fmt.Errorf("send failed: %v", err)
	}

	return nil
}

func etcdImport(keyVals []keyVal) error {
	// Connect to etcd
	db, err := utils.GetDbForAllAgents(global.Endpoints)
	if err != nil {
		return fmt.Errorf("connecting to Etcd failed: %v", err)
	}

	fmt.Printf("importing %d key vals\n", len(keyVals))

	var txn = db.NewTxn()
	ops := 0
	for i := 0; i < len(keyVals); i++ {
		keyVal := keyVals[i]

		fmt.Printf(" - %s\n", keyVal.Key)
		txn.Put(keyVal.Key, keyVal.Val)
		ops++

		if ops == int(txops) || i+1 == len(keyVals) {
			fmt.Printf("commiting tx with %d ops\n", ops)

			ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
			err = txn.Commit(ctx)
			cancel()
			if err != nil {
				return fmt.Errorf("commit failed: %v", err)
			}

			ops = 0
			txn = db.NewTxn()
		}
	}

	return nil
}

func parseKey(key string) (string, error) {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key, nil
	}
	if !strings.HasPrefix(key, "config/") {
		return "", fmt.Errorf("invalid format for key: %q", key)
	}
	return path.Join(servicelabel.GetAllAgentsPrefix(), global.ServiceLabel, key), nil
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

func getTimeout() time.Duration {
	return time.Second * time.Duration(timeout)
}
