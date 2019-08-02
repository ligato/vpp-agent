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

	//fmt.Printf("key: %s\n", key)

	if !strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		tmp := strings.Split(key, "/")
		if tmp[0] != "config" {
			globalFlags.Label = tmp[0]
			key = strings.Join(tmp[1:], "/")
		}

		db, err = utils.GetDbForOneAgent(globalFlags.Endpoints, globalFlags.Label)
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
