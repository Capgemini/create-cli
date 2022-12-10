package vault

import (
	"create-cli/internal/k8s"
	"create-cli/internal/log"
	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Vault"}
var vaultUrl, vaultUrlPresent = os.LookupEnv("VAULT_URL")

func Vault() {
	k8s.InitKubeClient()

	if !vaultUrlPresent {
		logger.Warningf("VAULT_URL env variable is not defined. Therefore using internal Vault SVC URL")
		vaultUrl = "http://vault.vault:8200"
	}
	CreateVaultClient()

	initialised := Healthcheck()
	if !initialised {
		InitAndUnseal()
		MountSecretsEngine()
		CreateAppRole()
		SetupOIDC()
	}
}
