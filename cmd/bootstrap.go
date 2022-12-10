package cmd

import (
	"create-cli/internal/concourse"
	"create-cli/internal/harbor"
	"create-cli/internal/keycloak"
	"create-cli/internal/sonarqube"
	"create-cli/internal/vault"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Bootstrap(args []string) {
	keycloak.Keycloak()
	vault.Vault()
	sonarqube.SonarQube()
	concourse.Concourse()
	harbor.Harbor()
	log.Println("Bootstrap complete.")
}

func init() {
	bootstrapCmd.Flags().StringVarP(&createUrl, "create-url", "", "", "The URL of CREATE (e.g. create.company.com")
	bootstrapCmd.MarkFlagRequired("create-url")
	viper.BindPFlag("create-url", bootstrapCmd.Flags().Lookup("create-url"))

	rootCmd.AddCommand(bootstrapCmd)
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstraps the initial tooling cluster by ensuring all tooling applications have been configured in the correct way ready for use.",
	Run: func(cmd *cobra.Command, args []string) {
		Bootstrap(args)
	},
}
