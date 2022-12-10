package vault

import (
	vault "github.com/hashicorp/vault/api"
)

func MountSecretsEngine() {

	mountInput := &vault.MountInput{
		Type:        "kv-v2",
		Description: "Secrets engine",
	}

	logger.Waitingf("Mounting `secrets/dev-team` KV2 engine...")
	err := vaultClient.Sys().Mount("secrets/dev-team", mountInput)
	if err != nil {
		logger.Failuref("unable to initialize vault client: %w", err)
		panic(err)
	}
	logger.Successf("Mounted `secrets/dev-team` KV2 engine")
}
