package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ligato/vpp-agent/api/models/linux"

	"github.com/ligato/vpp-agent/api/models/vpp"

	yaml2 "github.com/ghodss/yaml"

	"github.com/ligato/vpp-agent/pkg/models"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/vpp-agent/cmd/agentctl2/cmd_generator"
	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var generateConfig = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"g"},
	Short:   "Generate example command",
	Long: `
	Generate example command
`,
}

var generateACL = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.ACL{}),
	Short: "Generate VPP example ACL config",
	Long: `
	Generate VPP example ACL config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  aclGenerateFunction,
}

var generateInterface = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.Interface{}),
	Short: "Generate VPP example Interface config",
	Long: `
	Generate VPP example Interface config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  interfaceGenerateFunction,
}

var generateBd = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.BridgeDomain{}),
	Short: "Generate VPP example bridge domain config",
	Long: `
	Generate VPP example bridge domain config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  bdGenerateFunction,
}

var generateIPScanNeighbor = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.IPScanNeigh{}),
	Short: "Generate VPP example ip scan neighbor config",
	Long: `
	Generate VPP example ip scan neighbor config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  ipScanNeighborGenerateFunction,
}

var generateNatGlobal = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.NAT44Global{}),
	Short: "Generate VPP example NatGlobal config",
	Long: `
	Generate VPP example NatGlobal config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  natGlobalGenerateFunction,
}

var generateNatDNat = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.DNAT44{}),
	Short: "Generate VPP example dnat config",
	Long: `
	Generate VPP example dnat config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  natDNatGenerateFunction,
}

var generateIPSecPolicy = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.IPSecSPD{}),
	Short: "Generate VPP example ip sec policy config",
	Long: `
	Generate VPP example ip sec policy config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  ipSecPolicyGenerateFunction,
}

var generateIPSecAssociation = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.IPSecSA{}),
	Short: "Generate VPP example ip sec association config",
	Long: `
	Generate VPP example ip sec association config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  ipSecAssociateGenerateFunction,
}

var generateArps = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.ARPEntry{}),
	Short: "Generate VPP example arps config",
	Long: `
	Generate VPP example arps config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  arpsGenerateFunction,
}

var generateRoutes = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.Route{}),
	Short: "Generate VPP example routes config",
	Long: `
	Generate VPP example ip sec policy
`,
	Args: cobra.MaximumNArgs(0),
	Run:  routesGenerateFunction,
}

var generatePArp = &cobra.Command{
	Use:   utils.GetModuleName(&vpp.ProxyARP{}),
	Short: "Generate VPP example proxy arp config",
	Long: `
	Generate VPP example proxy arp config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  parpGenerateConfig,
}

var generateLinuxInterface = &cobra.Command{
	Use:   utils.GetModuleName(&linux.Interface{}),
	Short: "Generate Linux example interface config",
	Long: `
	Generate Linux example interface config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxInterfaceGenerateFunction,
}

var generateLinuxARP = &cobra.Command{
	Use:   utils.GetModuleName(&linux.ARPEntry{}),
	Short: "Generate Linux example arp config",
	Long: `
	Generate Linux example arp config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxARPGenerateFunction,
}

var generateLinuxRoutes = &cobra.Command{
	Use:   utils.GetModuleName(&linux.Route{}),
	Short: "Generate Linux example routes config",
	Long: `
	Generate Linux example routes config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxRoutesGenerateFunction,
}

var formatType *string

func init() {
	RootCmd.AddCommand(generateConfig)
	generateConfig.AddCommand(generateACL)
	generateConfig.AddCommand(generateInterface)
	generateConfig.AddCommand(generateBd)
	generateConfig.AddCommand(generateIPScanNeighbor)
	generateConfig.AddCommand(generateNatGlobal)
	generateConfig.AddCommand(generateNatDNat)
	generateConfig.AddCommand(generateIPSecPolicy)
	generateConfig.AddCommand(generateIPSecAssociation)
	generateConfig.AddCommand(generateArps)
	generateConfig.AddCommand(generateRoutes)
	generateConfig.AddCommand(generatePArp)
	generateConfig.AddCommand(generateLinuxInterface)
	generateConfig.AddCommand(generateLinuxARP)
	generateConfig.AddCommand(generateLinuxRoutes)
	formatType = generateConfig.PersistentFlags().String("format", "json",
		"Format:\n\tjson\n\tyaml\n\tproto\n")
}

func aclGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.ACL)
}

func interfaceGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.Interface)
}

func bdGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.Bd)
}

func ipScanNeighborGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.IPScanNeighbor)
}

func natGlobalGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.NatGlobal)
}

func natDNatGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.NatDNat)
}

func ipSecPolicyGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.IPSecPolicy)
}

func ipSecAssociateGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.IPSecAssociation)
}

func arpsGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.Arps)
}

func routesGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.Routes)
}

func parpGenerateConfig(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.PArp)
}

func linuxInterfaceGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.LinuxInterface)
}

func linuxARPGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.LinuxARPs)
}

func linuxRoutesGenerateFunction(cmd *cobra.Command, args []string) {
	generateFunction(cmd_generator.LinuxRoutes)
}

func generateFunction(gtype cmd_generator.CommandType) {
	msg := cmd_generator.GenerateConfig(gtype)

	switch *formatType {
	case "json":
		printJSON(msg)
	case "proto":
		printProto(msg)
	case "yaml":
		printYaml(msg)
	default:
		utils.ExitWithError(utils.ExitError, errors.New("Unknown format"))
	}

}

const prefix string = "/vnf-agent/vpp1/"

func printJSON(msg proto.Message) {

	js, err := json.MarshalIndent(msg, "", "  ")
	key := models.Key(msg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed generate json, error: "+err.Error()))
	}

	fmt.Fprintf(os.Stdout, "%s\n%s\n", prefix+key, js)
}

func printYaml(msg proto.Message) {
	js, err := json.MarshalIndent(msg, "", "  ")

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed generate json, error: "+err.Error()))
	}

	ym, err := yaml2.JSONToYAML(js)
	key := models.Key(msg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed generate yaml, error: "+err.Error()))
	}

	fmt.Fprintf(os.Stdout, "%s '%s'\n", prefix+key, ym)
}

func printProto(msg proto.Message) {
	text := proto.MarshalTextString(msg)
	key := models.Key(msg)

	fmt.Fprintf(os.Stdout, "%s '%s'\n", prefix+key, text)
}
