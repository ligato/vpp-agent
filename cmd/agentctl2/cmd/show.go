package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ligato/vpp-agent/api/models/vpp/acl"

	linux_interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linux_l3 "github.com/ligato/vpp-agent/api/models/linux/l3"

	"github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/api/models/vpp/nat"

	"github.com/ligato/vpp-agent/api/models/vpp/l2"

	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/l3"

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
	Run: showFunction,
}

var showConfig = &cobra.Command{
	Use:     "config [<module name> [<module type>]]",
	Aliases: []string{""},
	Short:   "Print list of set configuration type",
	Long: `
	Print list of set configuration type.
`,

	Run: configFunction,
}

var (
	showAll     bool
	showConf    bool
	showConfAll bool
	showStatus  bool
	keyPrefix   []string
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.AddCommand(showConfig)
	showCmd.PersistentFlags().BoolVar(&showAll, "all", false,
		"Show all configuration")
	showConfig.PersistentFlags().BoolVar(&showConfAll, "all", false,
		"Show all configuration")

	showConf = false
	showStatus = true
	keyPrefix = append(keyPrefix, "config")
}

func configFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false

	showFunction(cmd, args)
}

func setKeyPrefix(args []string) {
	var modulName, modulType string

	modulName = args[0]
	keyPrefix = nil

	if len(args) == 1 {
		keyPrefix = append(keyPrefix, "config/"+modulName)
		return
	}

	modulType = args[1]

	if strings.HasPrefix(vpp_interfaces.ModuleName, modulName) {
		if strings.HasPrefix(vpp_interfaces.ModelInterface.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_interfaces.ModelInterface.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_acl.ModelACL.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_acl.ModelACL.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l2.ModelBridgeDomain.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l2.ModelBridgeDomain.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l2.ModelFIBEntry.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l2.ModelFIBEntry.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l2.ModelXConnectPair.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l2.ModelXConnectPair.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l3.ModelRoute.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l3.ModelRoute.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l3.ModelARPEntry.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l3.ModelARPEntry.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l3.ModelProxyARP.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l3.ModelProxyARP.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_l3.ModelIPScanNeighbor.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_l3.ModelIPScanNeighbor.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_nat.ModelNat44Global.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_nat.ModelNat44Global.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_nat.ModelDNat44.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_nat.ModelDNat44.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_ipsec.ModelSecurityPolicyDatabase.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_ipsec.ModelSecurityPolicyDatabase.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(vpp_ipsec.ModelSecurityAssociation.Type, modulType) {
			keyPrefix = append(keyPrefix, vpp_ipsec.ModelSecurityAssociation.KeyPrefix())
			showConfAll = true
		}
	}

	if strings.HasPrefix(linux_interfaces.ModuleName, modulName) {
		if strings.HasPrefix(linux_interfaces.ModelInterface.Type, modulType) {
			keyPrefix = append(keyPrefix, linux_interfaces.ModelInterface.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(linux_l3.ModelARPEntry.Type, modulType) {
			keyPrefix = append(keyPrefix, linux_l3.ModelARPEntry.KeyPrefix())
			showConfAll = true
		}

		if strings.HasPrefix(linux_l3.ModelRoute.Type, modulType) {
			keyPrefix = append(keyPrefix, linux_l3.ModelRoute.KeyPrefix())
			showConfAll = true
		}
	}
}

func showFunction(cmd *cobra.Command, args []string) {
	db, err := utils.GetDbForAllAgents(globalFlags.Endpoints)
	if err != nil {
		utils.ExitWithError(utils.ExitError, errors.New("Failed to connect to Etcd - "+err.Error()))
	}

	if len(args) > 0 {
		setKeyPrefix(args)
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
		if showStatus || showAll {
			printAgentStatus(db1, agentLabel)
		}

		if showConf || showAll {
			for _, val := range keyPrefix {
				printAgentConfig(db1, agentLabel, val)
			}
		}
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
		buffer, err := ed.PrintStatus()
		if err == nil {
			fmt.Fprintf(os.Stdout, buffer.String())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "No data found.\n")
	}
}

func printAgentConfig(db keyval.ProtoBroker, agentLabel string, kprefix string) {

	keyIter, err := db.ListKeys(kprefix)
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
		buffer, err := ed.PrintConfig(showAll || showConfAll)
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
