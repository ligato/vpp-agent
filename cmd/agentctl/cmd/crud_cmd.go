package cmd

import "github.com/spf13/cobra"

// Root command 'put' can be used to create vpp configuration elements. The command uses no labels
// (except the global ones), so following command is required
var putCommand = &cobra.Command{
	Use:     "put",
	Aliases: []string{"p"},
	Short:   "Create or update vpp configuration ('C' and 'U' in CrUd).",
	Long: "Put vpp configuration attributes (Interfaces, Bridge Domains, L2," +
		" X-Connects, Routes).",
}

// Root command 'delete' can be used to remove vpp configuration elements. The command uses no labels
// (except the global ones), so following command is required
var deleteCommand = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete vpp configuration ('D' in cruD).",
	Long: "Remove vpp configuration attributes (Interfaces, Bridge Domains, L2," +
		"X-Connects, Routes).",
}

func init() {
	RootCmd.AddCommand(putCommand)
	RootCmd.AddCommand(deleteCommand)
}
