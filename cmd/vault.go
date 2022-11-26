package cmd

import (
	"create-cli/internal/vault"

	"github.com/spf13/cobra"
)

func Vault(args []string) {
	vault.Vault()
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Configures vault",
	Run: func(cmd *cobra.Command, args []string) {
		Vault(args)
	},
}
