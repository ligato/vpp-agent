package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/spf13/cobra"

	linux_interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linux_l3 "github.com/ligato/vpp-agent/api/models/linux/l3"
	vpp_acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	vpp_l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	vpp_nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
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
	Aliases: []string{"c"},
	Short:   "Print list of set configuration type",
	Long: `
'show config' prints out Etcd configuration type.
'show config <module name>' print out Etcd configuration type for set module name
'show config <module name> <module type>' print out Etcd configuration for set module type
`,
	Example: `Specify the Etcd to connect to and run show command:
	$ export ETCD_ENDPOINTS=172.17.0.1:2379
	$ ./agentctl show config
OR
	$ ./agenctl show config vpp
OR
	$ ./agenctl show config vpp route

Do as above, but with a command line flag:
	$ ./agentctl --endpoints 172.17.0.1:2379 show config
OR
	$ ./agentctl --endpoints 172.17.0.1:2379 show config vpp
OR
	$ ./agentctl --endpoints 172.17.0.1:2379 show config vpp route
`,

	Run: configFunction,
}

var keyConfig = &cobra.Command{
	Use:     "key",
	Aliases: []string{""},
	Short:   "Print list of stored key",
	Long: `
	Print list of stored key.
`,

	Run: keyFunction,
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
	showCmd.AddCommand(keyConfig)
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

func keyFunction(cmd *cobra.Command, args []string) {
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

	for {
		if key, _, done := keyIter.GetNext(); !done {
			fmt.Printf("Key: '%s'\n", key)
			continue
		}
		break
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
