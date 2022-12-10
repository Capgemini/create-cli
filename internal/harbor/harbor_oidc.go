package harbor

import (
	"bytes"
	"create-cli/internal/k8s"
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

type HarborOIDCRequest struct {
	AuthMode         string `json:"auth_mode"`
	OIDCName         string `json:"oidc_name"`
	OIDCEndpoint     string `json:"oidc_endpoint"`
	OIDCClientID     string `json:"oidc_client_id"`
	OIDCClientSecret string `json:"oidc_client_secret"`
	OIDCScope        string `json:"oidc_scope"`
	OIDCVerifyCert   bool   `json:"oidc_verify_cert"`
	OIDCAutoOnboard  bool   `json:"oidc_auto_onboard"`
	OIDCUserClaim    string `json:"oidc_user_claim"`
	OIDCGroupClaim   string `json:"oidc_groups_claim"`
	OIDCAdminGroup   string `json:"oidc_admin_group"`
}

func getHarborOIDCSecret() string {
	logger.Waitingf("Retrieving Harbor Keycloak Client Secret...")
	secret := k8s.GetSecret("create-secrets", "default")
	logger.Successf("Retrieved Harbor Keycloak Client Secret")

	//now we have retrieved the secret, we want to delete it
	deleteHarborOIDCSecret()
	return string(secret.Data["harbor-keycloak-client-secret"])
}

func deleteHarborOIDCSecret() {
	logger.Waitingf("Deleting Harbor Client Secret from default k8s secrets...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:   "remove",
		Path: "/data/harbor-keycloak-client-secret",
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Harbor Client Secret Deleted")
}

func ConfigureHarborOIDC() {
	createUrl := viper.GetString("create-url")

	harborOIDCRequest := HarborOIDCRequest{
		AuthMode:         "oidc_auth",
		OIDCName:         "keycloak",
		OIDCEndpoint:     "https://keycloak.tooling." + createUrl + "/realms/sso",
		OIDCClientID:     "harbor",
		OIDCClientSecret: getHarborOIDCSecret(),
		OIDCScope:        "openid",
		OIDCVerifyCert:   false,
		OIDCAutoOnboard:  true,
		OIDCUserClaim:    "preferred_username",
		OIDCGroupClaim:   "groups",
		OIDCAdminGroup:   "platform",
	}

	harborOIDCRequestJSON, err := json.Marshal(harborOIDCRequest)
	if err != nil {
		logger.Failuref("Error marshalling object into json for harbor OIDC request", err)
	}

	client := &http.Client{}
	logger.Waitingf("Configuring Harbor OIDC...")
	req, err := http.NewRequest("PUT", harborUrl+"/api/v2.0/configurations", bytes.NewBuffer(harborOIDCRequestJSON))
	if err != nil {
		logger.Failuref("Error creating HTTP request to configure harbor OIDC", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to configure harbor OIDC", err)
	}

	if resp.StatusCode == 200 {
		logger.Successf("Configured Harbor OIDC")
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}
