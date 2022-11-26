package cmd

import (
	"create-cli/internal/keycloak"

	"github.com/spf13/cobra"
)

func Keycloak(args []string) {
	keycloak.Keycloak()
}

func init() {
	rootCmd.AddCommand(keycloakCmd)
}

var keycloakCmd = &cobra.Command{
	Use:   "keycloak",
	Short: "Configures keycloak",
	Run: func(cmd *cobra.Command, args []string) {
		Keycloak(args)
	},
}
