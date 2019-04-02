package cmd

import (
	"fmt"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
	"github.com/spf13/cobra"

	"errors"
)

// RootCmd represents the base command when called without any subcommands.
var showConfig = &cobra.Command{
	Use:     "config",
	Aliases: []string{"C"},
	Short:   "Show configuration file",
	Long: `
	Show configuration file
`,
	Run: confFunction,
}

var (
	showConf	bool
)

func init() {
	RootCmd.AddCommand(showConfig)
	showConfig.Flags().BoolVar(&showConf,"verbose", false, "Show Configuration")
}

func confFunction(cmd *cobra.Command, args []string) {
	db, err := utils.GetDbForAllAgents(globalFlags.Endpoints)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
	}

	keyIter, err := db.ListKeys(servicelabel.GetAllAgentsPrefix())
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to get keys - "+err.Error()))
	}

	ed := utils.NewEtcdDump()
	for {
		if key, _, done := keyIter.GetNext(); !done {
			//fmt.Printf("Key: '%s'\n", key)
			if _, err = ed.ReadDataFromDb(db, key); err != nil {
				utils.ExitWithError(utils.ExitError, err)
			}
			continue
		}
		break
	}

	if len(ed) > 0 {
		buffer, err := ed.PrintTest(showConf)
		if err == nil {
			fmt.Print(buffer.String())
		} else {
			fmt.Printf("Error: %v", err)
		}
	} else {
		fmt.Print("No data found.\n")
	}

}
