package cmd

import (
	"create-cli/internal/configure"
	"create-cli/internal/download"
	"create-cli/internal/push"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloudProvider string
var acmeRegistrationEmail string
var backstageGitlabUserToken string
var backstageGitlabUserUsername string
var gitlabGroup string
var gitlabHost string

func init() {
	downloadFlags()
	configureFlags()
	pushFlags()

	preInstallCmd.AddCommand(downloadCmd)
	preInstallCmd.AddCommand(configureCmd)
	preInstallCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(preInstallCmd)
}

func downloadFlags() {
	downloadCmd.Flags().StringVarP(&cloudProvider, "cloud-provider", "", "", "The Cloud Provider that CREATE will exist in")
	downloadCmd.MarkFlagRequired("cloud-provider")

	// temporary whilst the repos are private. they will be open sourced and this won't be needed after that.
	downloadCmd.Flags().StringVarP(&personalAccessToken, "pat", "", "", "Personal Access Token for the Git Repository (Gitlab/GitHub)")
	downloadCmd.MarkFlagRequired("pat")
}

func configureFlags() {
	configureCmd.Flags().StringVarP(&createUrl, "create-url", "", "", "The URL of CREATE (e.g. create.company.com")
	configureCmd.MarkFlagRequired("create-url")

	configureCmd.Flags().StringVarP(&acmeRegistrationEmail, "acme-reg-email", "", "", "The ACME registration email to use for LetsEncrypt certificates")
	configureCmd.MarkFlagRequired("acme-reg-email")

	configureCmd.Flags().StringVarP(&backstageGitlabUserToken, "backstage-gitlab-token", "", "", "The Token belonging to the Backstage Gitlab user")
	configureCmd.MarkFlagRequired("backstage-gitlab-token")

	configureCmd.Flags().StringVarP(&backstageGitlabUserUsername, "backstage-gitlab-username", "", "", "The username of the Backstage Gitlab user")
	configureCmd.MarkFlagRequired("backstage-gitlab-username")

	configureCmd.Flags().StringVarP(&gitlabHost, "gitlab-host", "", "gitlab.com", "The Gitlab host to which the CREATE Git repositories will live. Defaults to gitlab.com")

	configureCmd.Flags().StringVarP(&gitlabGroup, "gitlab-group", "", "", "The group (or owner) to which the CREATE Git repositories will live. Example: 'group/subgroup' in 'gitlab.com/group/subgroup'")
	configureCmd.MarkFlagRequired("gitlab-group")
}

func pushFlags() {
	pushCmd.Flags().StringVarP(&personalAccessToken, "pat", "", "", "Personal Access Token for the Git Repository (Gitlab/GitHub)")
	pushCmd.MarkFlagRequired("pat")

	pushCmd.Flags().StringVarP(&gitlabHost, "host", "", "gitlab.com", "The Gitlab host to push CREATE repositories into. Defaults to gitlab.com")

	pushCmd.Flags().StringVarP(&gitlabGroup, "gitlab-group", "", "", "The Gitlab group to push CREATE repositories into. Example: 'subgroup' in 'group/subgroup'")
	pushCmd.MarkFlagRequired("gitlab-group")
}

var preInstallCmd = &cobra.Command{
	Use:   "pre-install",
	Short: "Runs actions that are focused on the pre-installation configuration phase of CREATE.",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: must also specify a sub-command.")
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads all CREATE repositories via Git",
	Args:  cobra.MinimumNArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("cloud-provider", cmd.Flags().Lookup("cloud-provider"))
		viper.BindPFlag("pat", cmd.Flags().Lookup("pat"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		download.Download()
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures clone repositories with generated values", // better description
	Args:  cobra.MinimumNArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("create-url", cmd.Flags().Lookup("create-url"))
		viper.BindPFlag("acme-reg-email", cmd.Flags().Lookup("acme-reg-email"))
		viper.BindPFlag("backstage-gitlab-token", cmd.Flags().Lookup("backstage-gitlab-token"))
		viper.BindPFlag("backstage-gitlab-username", cmd.Flags().Lookup("backstage-gitlab-username"))
		viper.BindPFlag("gitlab-host", cmd.Flags().Lookup("gitlab-host"))
		viper.BindPFlag("gitlab-group", cmd.Flags().Lookup("gitlab-group"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		configure.Configure()
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes all repositories into upstream Git repository",
	Args:  cobra.MinimumNArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("pat", cmd.Flags().Lookup("pat"))
		viper.BindPFlag("gitlab-host", cmd.Flags().Lookup("gitlab-host"))
		viper.BindPFlag("gitlab-group", cmd.Flags().Lookup("gitlab-group"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		push.Push()
	},
}
