package cmd

import (
	"create-cli/internal/push"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Push(args []string) {
	push.Push()
}

var group string
var host string

func init() {
	pushCmd.Flags().StringVarP(&personalAccessToken, "pat", "", "", "Personal Access Token for the Git Repository (Gitlab/GitHub)")
	rootCmd.MarkFlagRequired("pat")
	viper.BindPFlag("pat", pushCmd.Flags().Lookup("pat"))

	pushCmd.Flags().StringVarP(&host, "host", "", "gitlab.com", "The Gitlab host to push CREATE repositories into. Defaults to gitlab.com")
	viper.BindPFlag("host", pushCmd.Flags().Lookup("host"))

	pushCmd.Flags().StringVarP(&group, "gitlab-group", "", "", "The Gitlab group to push CREATE repositories into. Example: 'subgroup' in 'group/subgroup'")
	pushCmd.MarkFlagRequired("gitlab-group")
	viper.BindPFlag("gitlab-group", pushCmd.Flags().Lookup("gitlab-group"))

	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes all repositories into upstream Git repository",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		Push(args)
	},
}
