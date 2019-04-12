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
	Args: cobra.RangeArgs(2, 2),
	Example: ` Set route configuration for "vpp1":
   $./agentctl2 -e 172.17.0.3:2379 put /vnf-agent/vpp1/config/vpp/v2/route/vrf/1/dst/10.1.1.3/32/gw/192.168.1.13 '{
   "type": 1,
   "vrf_id": 1,
   "dst_network": "10.1.1.3/32",
   "next_hop_addr": "192.168.1.13"
    }'
`,

	Run: putFunction,
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
