package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of create-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create-cli v0.1")
	},
}
