package cmd

import (
	"errors"
	"fmt"

	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var putConfig = &cobra.Command{
	Use:     "put",
	Aliases: []string{"p"},
	Short:   "Put configuration file",
	Long: `
	Put configuration file
`,
	Args: cobra.MinimumNArgs(2),
	Run:  putFunction,
}

func init() {
	RootCmd.AddCommand(putConfig)
}

func putFunction(cmd *cobra.Command, args []string) {
	key := args[0]
	json := args[1]

	fmt.Printf("key: %s, json: %s\n", key, json)

	db, err := utils.GetDbForAllAgents(globalFlags.Endpoints)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
	}

	utils.WriteData(db.NewTxn(), key, json)
}
