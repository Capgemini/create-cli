package vault

import (
	"create-cli/internal/k8s"
	"encoding/base64"
	"strconv"

	vault "github.com/hashicorp/vault/api"
)

var initRequest = vault.InitRequest{
	SecretShares:    5,
	SecretThreshold: 3,
}

func InitAndUnseal() {
	vaultInitalisedResponse := initializeVault()
	unsealVault(vaultInitalisedResponse)
	addVaultSecrets(vaultInitalisedResponse)
}

func initializeVault() *vault.InitResponse {
	logger.Waitingf("Vault is uninitialized, initializing now...")

	vaultInitalisedResponse, err := vaultClient.Sys().Init(&initRequest)
	if err != nil {
		logger.Failuref("error initialising vault: %w", err)
		panic(err)
	}

	return vaultInitalisedResponse
}

func unsealVault(vaultInitalisedResponse *vault.InitResponse) {
	logger.Waitingf("Vault is initialized, unsealing now...")
	for i := 0; i < initRequest.SecretThreshold; i++ {
		unsealResponse, err := vaultClient.Sys().Unseal(vaultInitalisedResponse.Keys[i])
		if err != nil {
			logger.Failuref("error unsealing vault: %w", err)
			panic(err)
		}

		if !unsealResponse.Sealed {
			logger.Successf("Vault is unsealed.")
			break
		}
	}
}

func addVaultSecrets(vaultInitalisedResponse *vault.InitResponse) {
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/vault-root-token",
		Value: base64.StdEncoding.EncodeToString([]byte(vaultInitalisedResponse.RootToken)),
	}}

	logger.Waitingf("Adding unseal keys and root key to default secret...")
	for i := 0; i < initRequest.SecretShares; i++ {
		patchSecretRequest = append(patchSecretRequest, k8s.PatchSecretRequest{
			Op:    "add",
			Path:  "/data/vault-unseal-key-" + strconv.Itoa(i+1),
			Value: base64.StdEncoding.EncodeToString([]byte(vaultInitalisedResponse.KeysB64[i])),
		})
	}

	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Keys added to default secret")
}
