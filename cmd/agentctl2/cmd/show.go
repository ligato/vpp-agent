package cmd

import (
	"fmt"
	"os"

	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"

	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
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
	Use:     "config",
	Aliases: []string{""},
	Short:   "Print list of set configuration data",
	Long: `
	Print list of set configuration data
`,
	Run: configFunction,
}

var showConfigVpp = &cobra.Command{
	Use:     "vpp",
	Aliases: []string{""},
	Short:   "Print vpp configuration data",
	Long: `
	Print vpp configuration data
`,
	Run: configVppFunction,
}

var showConfigLinux = &cobra.Command{
	Use:     "linux",
	Aliases: []string{""},
	Short:   "Print linux configuration data",
	Long: `
	Print linux configuration data
`,
	Run: configLinuxFunction,
}

var showConfigAcl = &cobra.Command{
	Use:     "acl",
	Aliases: []string{""},
	Short:   "Print vpp acl configuration data",
	Long: `
	Print vpp acl configuration data
`,
	Run: aclFunction,
}

var showConfigBd = &cobra.Command{
	Use:     "bridgedomain",
	Aliases: []string{""},
	Short:   "Print vpp bridge domain configuration data",
	Long: `
	Print vpp bridge domain configuration data
`,
	Run: bdFunction,
}

var showConfigFib = &cobra.Command{
	Use:     "fib",
	Aliases: []string{""},
	Short:   "Print vpp fib configuration data",
	Long: `
	Print vpp fib configuration data
`,
	Run: fibFunction,
}

var showConfigXconnect = &cobra.Command{
	Use:     "xconnect",
	Aliases: []string{""},
	Short:   "Print vpp xconnect configuration data",
	Long: `
	Print vpp xconnect configuration data
`,
	Run: xconnectFunction,
}

var showConfigArp = &cobra.Command{
	Use:     "arp",
	Aliases: []string{""},
	Short:   "Print arp configuration data",
	Long: `
	Print vpp arp configuration data
`,
	Run: arpFunction,
}

var showConfigRoute = &cobra.Command{
	Use:     "route",
	Aliases: []string{""},
	Short:   "Print vpp route configuration data",
	Long: `
	Print vpp route configuration data
`,
	Run: routeFunction,
}

var showConfigProxyArp = &cobra.Command{
	Use:     "proxyarp",
	Aliases: []string{""},
	Short:   "Print vpp proxy arp configuration data",
	Long: `
	Print vpp proxy arp configuration data
`,
	Run: proxyArpFunction,
}

var showConfigInterface = &cobra.Command{
	Use:     "interface",
	Aliases: []string{""},
	Short:   "Print vpp interface configuration data",
	Long: `
	Print interface configuration data
`,
	Run: ipInterfaceFunction,
}

var showConfigIpneighbor = &cobra.Command{
	Use:     "ipneighbor",
	Aliases: []string{""},
	Short:   "Print vpp ip neighbor configuration data",
	Long: `
	Print ip neighbor configuration data
`,
	Run: ipNeighborFunction,
}

var showConfigSpolicy = &cobra.Command{
	Use:     "ipsecpolicy",
	Aliases: []string{""},
	Short:   "Print vpp ip sec policy configuration data",
	Long: `
	Print ip sec policy configuration data
`,
	Run: ipSecPolicyFunction,
}

var showConfigSAss = &cobra.Command{
	Use:     "ipsecassociation",
	Aliases: []string{""},
	Short:   "Print vpp ip sec association configuration data",
	Long: `
	Print ip sec association configuration data
`,
	Run: ipSecAsssociationFunction,
}

var showConfigLInterface = &cobra.Command{
	Use:     "interface",
	Aliases: []string{""},
	Short:   "Print interface configuration data",
	Long: `
	Print Linux interface configuration data
`,
	Run: linterfaceFunction,
}

var showConfigLArp = &cobra.Command{
	Use:     "arp",
	Aliases: []string{""},
	Short:   "Print Linux arp configuration data",
	Long: `
	Print arp configuration data
`,
	Run: arpFunction,
}

var showConfigLRoute = &cobra.Command{
	Use:     "route",
	Aliases: []string{""},
	Short:   "Print Linux route configuration data",
	Long: `
	Print Linux route interface configuration data
`,
	Run: lrouteFunction,
}

var (
	showAll     bool
	showConf    bool
	showConfAll bool
	showStatus  bool
	keyPrefix   string
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.AddCommand(showConfig)
	showConfig.AddCommand(showConfigVpp)
	showConfig.AddCommand(showConfigLinux)
	showConfigVpp.AddCommand(showConfigAcl)
	showConfigVpp.AddCommand(showConfigBd)
	showConfigVpp.AddCommand(showConfigFib)
	showConfigVpp.AddCommand(showConfigXconnect)
	showConfigVpp.AddCommand(showConfigArp)
	showConfigVpp.AddCommand(showConfigRoute)
	showConfigVpp.AddCommand(showConfigProxyArp)
	showConfigVpp.AddCommand(showConfigInterface)
	showConfigVpp.AddCommand(showConfigIpneighbor)
	showConfigVpp.AddCommand(showConfigSpolicy)
	showConfigVpp.AddCommand(showConfigSAss)
	showConfigLinux.AddCommand(showConfigLInterface)
	showConfigLinux.AddCommand(showConfigLArp)
	showConfigLinux.AddCommand(showConfigLRoute)
	showCmd.PersistentFlags().BoolVar(&showAll, "all", false,
		"Show all configuration")
	showConfig.PersistentFlags().BoolVar(&showConfAll, "all", false,
		"Show all configuration")
	showConfigVpp.PersistentFlags().BoolVar(&showAll, "all", false,
		"Show all configuration")

	showConf = false
	showStatus = true
	keyPrefix = "config"
}

func configFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false

	showFunction(cmd, args)
}

func configVppFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false
	keyPrefix = "config/vpp"

	showFunction(cmd, args)
}

func configLinuxFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false
	keyPrefix = "config/linux"

	showFunction(cmd, args)
}

func aclFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false
	showConfAll = true
	keyPrefix = vpp_acl.ModelACL.KeyPrefix()

	showFunction(cmd, args)
}

func arpFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false
	showConfAll = true
	keyPrefix = vpp_l3.ModelARPEntry.KeyPrefix()

	showFunction(cmd, args)
}

func bdFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showStatus = false
	showConfAll = true
	keyPrefix = vpp_l2.ModelBridgeDomain.KeyPrefix()

	showFunction(cmd, args)
}

func fibFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_l2.ModelFIBEntry.KeyPrefix()

	showFunction(cmd, args)
}

func ipInterfaceFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_interfaces.ModelInterface.KeyPrefix()

	showFunction(cmd, args)
}

func ipNeighborFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_l3.ModelIPScanNeighbor.KeyPrefix()

	showFunction(cmd, args)
}

func ipSecAsssociationFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_ipsec.ModelSecurityAssociation.KeyPrefix()

	showFunction(cmd, args)
}

func ipSecPolicyFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_ipsec.ModelSecurityPolicyDatabase.KeyPrefix()

	showFunction(cmd, args)
}

func proxyArpFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_l3.ModelProxyARP.KeyPrefix()

	showFunction(cmd, args)
}

func routeFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_l3.ModelRoute.KeyPrefix()

	showFunction(cmd, args)
}

func xconnectFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = vpp_l2.ModelXConnectPair.KeyPrefix()

	showFunction(cmd, args)
}

func linterfaceFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = linux_interfaces.ModelInterface.KeyPrefix()

	showFunction(cmd, args)
}

func lrouteFunction(cmd *cobra.Command, args []string) {
	showConf = true
	showConfAll = true
	showStatus = false
	keyPrefix = linux_l3.ModelRoute.KeyPrefix()

	showFunction(cmd, args)
}

func showFunction(cmd *cobra.Command, args []string) {
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
		if showStatus || showAll {
			printAgentStatus(db1, agentLabel)
		}

		if showConf || showAll {
			printAgentConfig(db1, agentLabel)
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
		buffer, err := ed.PrintStatus(showAll)
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

	keyIter, err := db.ListKeys(keyPrefix)
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
