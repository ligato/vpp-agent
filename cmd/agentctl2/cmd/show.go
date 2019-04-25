package cmd

import (
	"fmt"
	"os"

	"github.com/ligato/cn-infra/health/statuscheck/model/status"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
	"github.com/spf13/cobra"

	"errors"
)

// RootCmd represents the base command when called without any subcommands.
var showCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"s"},
	Short:   "Show detailed config and status data",
	Long: `
'show' prints out Etcd configuration and status data (where applicable)
for agents whose microservice label matches the label filter specified
in the command's '[agent-label-filter] argument. The filter contains a
list of comma-separated strings. A match is performed for each string
in the filter list.

Note that agent's configuration data is stored into Etcd by a 3rd party
orchestrator. Agent's state data is periodically updated in Etcd by
healthy agents themselves. Agents for which only configuration records
exist (i.e. they do not push status records into Etcd) are listed as
'INACTIVE'.

The etcd flag set to true enables the printout of Etcd metadata for each
data record (except JSON-formatted output)
`,
	Run: confFunction,
}

var showConfig = &cobra.Command{
	Use:     "config",
	Aliases: []string{""},
	Short:   "Show config data",
	Long: `
	Show config data
`,
	Run: confFunction,
}

var (
	showConf bool
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().BoolVar(&showConf, "verbose", false,
		"Show Configuration")
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

	agentLabels := make([]string, 0)

	for {
		if key, _, done := keyIter.GetNext(); !done {

			agentLabel := utils.GetAgentLabel(key)
			addUniqueString(agentLabel, &agentLabels)
			continue
		}
		break
	}

	for _, agentLabel := range agentLabels {
		db1 := db.NewBroker(servicelabel.GetAllAgentsPrefix() + agentLabel + "/")
		printAgentStatus(db1, agentLabel)
		printAgentConfig(db1, agentLabel)
	}
}

func printAgentStatus(db keyval.ProtoBroker, agentLabel string) {

	keyIter, err := db.ListKeys(status.StatusPrefix)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to get keys - "+err.Error()))
	}

	ed := utils.NewEtcdDump()
	for {
		if key, _, done := keyIter.GetNext(); !done {
			//fmt.Printf("Key: '%s'\n", key)
			if _, err = ed.ReadStatusDataFromDb(db, key, agentLabel); err != nil {
				utils.ExitWithError(utils.ExitError, err)
			}
			continue
		}
		break
	}

	if len(ed) > 0 {
		buffer, err := ed.PrintStatus(showConf)
		if err == nil {
			fmt.Fprintf(os.Stdout, buffer.String())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "No data found.\n")
	}
}

func printAgentConfig(db keyval.ProtoBroker, agentLabel string) {

	keyIter, err := db.ListKeys("config")
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to get keys - "+err.Error()))
	}

	ed := utils.NewEtcdDump()
	for {
		if key, _, done := keyIter.GetNext(); !done {
			//fmt.Printf("Key: '%s'\n", key)
			if _, err = ed.ReadDataFromDb(db, key, agentLabel); err != nil {
				utils.ExitWithError(utils.ExitError, err)
			}
			continue
		}
		break
	}

	if len(ed) > 0 {
		buffer, err := ed.PrintConfig(showConf)
		if err == nil {
			fmt.Fprintf(os.Stdout, buffer.String())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "No data found.\n")
	}
}

func addUniqueString(str string, unique *[]string) {

	for _, value := range *unique {
		if value == str {
			return
		}
	}

	*unique = append(*unique, str)
}
