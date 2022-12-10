package vault

import (
	vault "github.com/hashicorp/vault/api"
)

var vaultClient vault.Client

func CreateVaultClient() {
	newVaultClient, err := vault.NewClient(getVaultConfig())

	if err != nil {
		logger.Failuref("unable to initialize vault client: %w", err)
		panic(err)
	}

	vaultClient = *newVaultClient
}

// CreateVaultClientForUnsealedInstance will return a new vaultClient for a Vault
// instance that is already initialised and unsealed
func CreateVaultClientForUnsealedInstance(rootToken string) *vault.Client {
	newVaultClient, err := vault.NewClient(getVaultConfig())
	if err != nil {
		logger.Failuref("unable to initialize vault client: %w", err)
		panic(err)
	}

	newVaultClient.SetToken(rootToken)
	return newVaultClient
}

func getVaultConfig() *vault.Config {
	config := vault.DefaultConfig()
	config.Address = vaultUrl
	return config
}
