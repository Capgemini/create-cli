package cmd

import (
	"create-cli/internal/sealed_secrets"

	"github.com/spf13/cobra"
)

func SealedSecrets(args []string) {
	sealed_secrets.SealedSecrets()
}
func init() {
	rootCmd.AddCommand(sealedSecretsCmd)
}

var sealedSecretsCmd = &cobra.Command{
	Use:   "sealed-secrets",
	Short: "Generates sealed secret keys",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		SealedSecrets(args)
	},
}
