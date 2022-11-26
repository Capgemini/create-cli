package cmd

import (
	"create-cli/internal/harbor"

	"github.com/spf13/cobra"
)

func Harbor(args []string) {
	harbor.Harbor()
}

func init() {
	rootCmd.AddCommand(harborCmd)
}

var harborCmd = &cobra.Command{
	Use:   "harbor",
	Short: "Configures Harbor",
	Run: func(cmd *cobra.Command, args []string) {
		Harbor(args)
	},
}
