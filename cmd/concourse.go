package cmd

import (
	"create-cli/internal/concourse"

	"github.com/spf13/cobra"
)

func Concourse(args []string) {
	concourse.Concourse()
}

func init() {
	rootCmd.AddCommand(concourseCmd)
}

var concourseCmd = &cobra.Command{
	Use:   "concourse",
	Short: "Configures Concourse",
	Run: func(cmd *cobra.Command, args []string) {
		Concourse(args)
	},
}
