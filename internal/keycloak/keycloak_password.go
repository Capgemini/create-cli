package keycloak

import (
	"create-cli/internal/k8s"
	"encoding/base64"
)

// GetKeycloakPasswords retrieves the passwords of interest from secrets in the keycloak namespace
// and puts them into a new secret for easy retrieval when create is fully setup and configured.
func GetKeycloakPasswords() {
	logger.Waitingf("Retrieving keycloak admin user password")
	keycloakSecret := k8s.GetSecret("keycloak", "keycloak")
	logger.Successf("Retrieved keycloak admin user password")

	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/keycloak-admin",
		Value: base64.StdEncoding.EncodeToString(keycloakSecret.Data["admin-password"]),
	}}
	logger.Waitingf("Adding password to default `create-secrets`...")
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Added password to default `create-secrets`")
}
