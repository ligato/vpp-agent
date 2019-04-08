package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

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
	Run: generateFunction,
}

var formatType *string

func init() {
	RootCmd.AddCommand(generateConfig)
	formatType = generateConfig.Flags().String("format", "json",
		"Format:\n\tjson\n\tyaml\n\tproto\n")
}

func generateFunction(cmd *cobra.Command, args []string) {
	var gtype cmd_generator.CommandType

	//TODO: Need rewrite, need list, use combra, but how???
	if 1 != len(args) {
		utils.ExitWithError(utils.ExitError, errors.New("Wrong Input arguments"))
	}

	tp := args[0]

	switch tp {
	case "interfaces":
		gtype = cmd_generator.VPPInterface
	case "acl":
		gtype = cmd_generator.VPPACL
	case "arp":
		gtype = cmd_generator.VPPARP
	case "bridge_domain":
		gtype = cmd_generator.VPPBridgeDomain
	case "route":
		gtype = cmd_generator.VPPRoute
	case "proxy_arp":
		gtype = cmd_generator.VPPProxyARP
	case "ip_scan_neighbor":
		gtype = cmd_generator.VPPIPScanNeighbor
	case "dnat44":
		gtype = cmd_generator.VPPDNat
	case "nat44-global":
		gtype = cmd_generator.VPPNat
	case "ipsec_policy":
		gtype = cmd_generator.VPPIPSecPolicy
	case "ipsec_association":
		gtype = cmd_generator.VPPIPSecAssociation
	case "linux_interface":
		gtype = cmd_generator.LinuxInterface
	case "linux_arp":
		gtype = cmd_generator.LinuxARP
	case "linux_route":
		gtype = cmd_generator.LinuxRoute
	default:
		utils.ExitWithError(utils.ExitError, errors.New("Unknown config type"))
	}

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

func printJSON(msg proto.Message) {

	js, err := json.MarshalIndent(msg, "", "  ")
	key := models.Key(msg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed generate json, error: "+err.Error()))
	}

	fmt.Printf("%s\n%s\n", key, js)
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

	fmt.Printf("%s\n%s\n", key, ym)
}

func printProto(msg proto.Message) {
	text := proto.MarshalTextString(msg)
	key := models.Key(msg)

	fmt.Printf("%s\n%s\n", key, text)
}
