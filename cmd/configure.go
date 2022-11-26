package cmd

import (
	"create-cli/internal/configure"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Configure(args []string) {
	configure.Configure()
}

var acmeRegistrationEmail string
var backstageGitlabUserToken string
var backstageGitlabUserUsername string
var gitlabGroup string
var gitlabHost string

func init() {
	configureCmd.Flags().StringVarP(&createUrl, "create-url", "", "", "The URL of CREATE (e.g. create.company.com")
	configureCmd.MarkFlagRequired("create-url")
	viper.BindPFlag("create-url", configureCmd.Flags().Lookup("create-url"))

	configureCmd.Flags().StringVarP(&acmeRegistrationEmail, "acme-reg-email", "", "", "The ACME registration email to use for LetsEncrypt certificates")
	configureCmd.MarkFlagRequired("acme-reg-email")
	viper.BindPFlag("acme-reg-email", configureCmd.Flags().Lookup("acme-reg-email"))

	configureCmd.Flags().StringVarP(&backstageGitlabUserToken, "backstage-gitlab-token", "", "", "The Token belonging to the Backstage Gitlab user")
	configureCmd.MarkFlagRequired("backstage-gitlab-token")
	viper.BindPFlag("backstage-gitlab-token", configureCmd.Flags().Lookup("backstage-gitlab-token"))

	configureCmd.Flags().StringVarP(&backstageGitlabUserUsername, "backstage-gitlab-username", "", "", "The username of the Backstage Gitlab user")
	configureCmd.MarkFlagRequired("backstage-gitlab-username")
	viper.BindPFlag("backstage-gitlab-username", configureCmd.Flags().Lookup("backstage-gitlab-username"))

	configureCmd.Flags().StringVarP(&gitlabHost, "gitlab-host", "", "gitlab.com", "The Gitlab host to which the CREATE Git repositories will live. Defaults to gitlab.com")
	viper.BindPFlag("gitlab-host", configureCmd.Flags().Lookup("gitlab-host"))

	configureCmd.Flags().StringVarP(&gitlabGroup, "gitlab-group", "", "", "The group (or owner) to which the CREATE Git repositories will live. Example: 'group/subgroup' in 'gitlab.com/group/subgroup'")
	configureCmd.MarkFlagRequired("gitlab-group")
	viper.BindPFlag("gitlab-group", configureCmd.Flags().Lookup("gitlab-group"))

	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures clone repositories with generated values", // better description
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		Configure(args)
	},
}
