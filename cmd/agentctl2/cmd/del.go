package cmd

import (
	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
	"github.com/spf13/cobra"

	"errors"
)

// RootCmd represents the base command when called without any subcommands.
var delConfig = &cobra.Command{
	Use:     "del",
	Aliases: []string{"d"},
	Short:   "Delete configuration file",
	Long: `
	Delete configuration file
`,
	Args: cobra.MaximumNArgs(1),
	Run:  delFunction,
}

func init() {
	RootCmd.AddCommand(delConfig)
}

func delFunction(cmd *cobra.Command, args []string) {

	key := args[0]

	//fmt.Printf("key: %s\n", key)

	db, err := utils.GetDbForAllAgents(globalFlags.Endpoints)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
	}

	utils.DelDataFromDb(db.NewTxn(), key)
}
