package vault

import (
	"create-cli/internal/k8s"
	"encoding/json"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// need to be ENV VAR driven
var oidcConfigAccessor string
var oidcKeyName = "keycloak"

func createOIDCKey() {
	logger.Waitingf("Creating OIDC Key...")
	_, err := vaultClient.Logical().Write("/identity/oidc/key/"+oidcKeyName, map[string]interface{}{
		"algorithm": "RS256",
	})
	if err != nil {
		logger.Failuref("Error creating OIDC Key")
		panic(err)
	}

	logger.Successf("OIDC Key Created")
}

func enableOIDC() {
	logger.Waitingf("Enabling OIDC...")
	oidcOptions := &vault.EnableAuthOptions{
		Type: "oidc",
	}

	vaultClient.Sys().EnableAuthWithOptions("oidc", oidcOptions)
	logger.Successf("OIDC Enabled")
}

func getVaultOIDCSecret() string {
	logger.Waitingf("Retrieving Vault Keycloak Client Secret")
	secret := k8s.GetSecret("create-secrets", "default")
	logger.Successf("Retrieved Vault Keycloak Client Secret")

	return string(secret.Data["vault-keycloak-client-secret"])
}

func deleteVaultOIDCSecret() {
	logger.Waitingf("Deleting Vault Client Secret from default k8s secrets...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:   "remove",
		Path: "/data/vault-keycloak-client-secret",
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Vault Client Secret Deleted")
}

func createOIDCConfig() {
	createUrl := viper.GetString("create-url")

	// we wrap it in a for loop because sometimes the oidc config call fails around the keycloak
	// discovery url. we aren't sure why. but we know if we send the request again it works.
	// we suspect there's just some weird race condition sometimes that keycloak oidc isn't ready yet.
	for {
		logger.Waitingf("Creating OIDC JWT Config...")
		_, err := vaultClient.Logical().Write("/auth/oidc/config", map[string]interface{}{
			"default_role":       "default",
			"oidc_discovery_url": "https://keycloak.tooling." + createUrl + "/realms/sso",
			"oidc_client_id":     "vault",
			"oidc_client_secret": getVaultOIDCSecret(),
		})

		if err != nil {
			logger.Failuref("Error creating OIDC JWT config, retrying.")
			time.Sleep(time.Duration(10) * time.Second)
			continue
		}
		//now we have retrieved the secret, we want to delete it
		deleteVaultOIDCSecret()
		logger.Successf("OIDC JWT Config Created")
		break
	}

	logger.Waitingf("Creating OIDC Tuning config...")
	// tuning here https://www.vaultproject.io/api-docs/system/auth#tune-auth-method
	_, err := vaultClient.Logical().Write("/sys/auth/oidc/tune", map[string]interface{}{
		"default_lease_ttl":            "1h",
		"listing_visibility":           "unauth",
		"max_lease_ttl":                "1h",
		"token_type":                   "default-service",
		"audit_non_hmac_request_keys":  []string{},
		"audit_non_hmac_response_keys": []string{},
		"passthrough_request_headers":  []string{},
	})
	if err != nil {
		logger.Failuref("Error creating OIDC JWT Tuning", err)
		panic(err)
	}
	logger.Successf("OIDC Tuning Config Created")

	resp, err := vaultClient.Sys().ListAuth()
	if err != nil {
		logger.Failuref("Error retrieving list of enabled auth methods", err)
	}

	// here we get the accessor of the OIDC method so that we can use it later on
	oidcConfigAccessor = resp["oidc/"].Accessor
}

type OIDCRoleRequest struct {
	RoleType            string        `json:"role_type"`
	TokenTTL            string        `json:"token_ttl"`
	TokenMaxTTL         string        `json:"token_max_ttl"`
	BoundAudiences      []string      `json:"bound_audiences"`
	UserClaim           string        `json:"user_claim"`
	ClaimMappings       ClaimMappings `json:"claim_mappings"`
	AllowedRedirectUris []string      `json:"allowed_redirect_uris"`
	GroupsClaim         string        `json:"groups_claim"`
}

type ClaimMappings struct {
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
}

func createOIDCRole() {
	createUrl := viper.GetString("create-url")

	logger.Waitingf("Creating OIDC Role...")
	oidcRoleRequest := OIDCRoleRequest{
		RoleType:       "oidc",
		TokenTTL:       "3600",
		TokenMaxTTL:    "3600",
		BoundAudiences: []string{"vault"},
		UserClaim:      "sub",
		ClaimMappings: ClaimMappings{
			PreferredUsername: "preferred_username",
			Email:             "email",
		},
		AllowedRedirectUris: []string{
			"https://vault.tooling." + createUrl + "/ui/vault/auth/oidc/oidc/callback",
			"https://vault.tooling." + createUrl + "/oidc/oidc/callback",
		},
		GroupsClaim: "/groups",
	}

	var request map[string]interface{}
	inrec, _ := json.Marshal(oidcRoleRequest)
	json.Unmarshal(inrec, &request)

	_, err := vaultClient.Logical().Write("/auth/oidc/role/default", request)
	if err != nil {
		logger.Failuref("Error creating OIDC role", err)
		panic(err)
	}

	logger.Successf("OIDC Role Created")
}

func createGroupPolicy(groupName string) {
	var policy string

	if groupName == "dev" {
		policy = `path "secrets/dev-team/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}
`
	}

	if groupName == "platform" {
		policy = `path "secrets/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}
`
	}

	logger.Waitingf("Creating %s group policy...", groupName)
	err := vaultClient.Sys().PutPolicy(groupName, policy)
	if err != nil {
		logger.Failuref("Error creating %s group policy", groupName, err)
		panic(err)
	}
	logger.Successf("%s group policy created", groupName)
}

func createGroup(groupName string) string {
	logger.Waitingf("Creating %s group...", groupName)
	resp, err := vaultClient.Logical().Write("/identity/group", map[string]interface{}{
		"name":     groupName,
		"type":     "external",
		"policies": []string{groupName},
	})

	if err != nil {
		logger.Failuref("Error creating group", err)
		panic(err)
	}

	logger.Successf("%s group created", groupName)
	return resp.Data["id"].(string)
}

func createGroupAlias(groupName string, groupId string) {
	logger.Waitingf("Creating %s group alias...", groupName)
	_, err := vaultClient.Logical().Write("/identity/group-alias", map[string]interface{}{
		"name":           groupName,
		"mount_accessor": oidcConfigAccessor,
		"canonical_id":   groupId,
	})
	if err != nil {
		logger.Failuref("Error creating group alias", err)
		panic(err)
	}

	logger.Successf("%s group alias created", groupName)
}

func createRole(groupName string) {
	logger.Waitingf("Creating %s role...", groupName)
	_, err := vaultClient.Logical().Write("/identity/oidc/role/"+groupName, map[string]interface{}{
		"name": groupName,
		"key":  oidcKeyName,
	})
	if err != nil {
		logger.Failuref("Error creating role", err)
		panic(err)
	}

	logger.Successf("%s role created", groupName)
}

func createGroupAndRole(groupName string) {
	createGroupPolicy(groupName)
	groupId := createGroup(groupName)
	createGroupAlias(groupName, groupId)
	createRole(groupName)
}

func SetupOIDC() {
	createOIDCKey()
	enableOIDC()
	createOIDCConfig()
	createOIDCRole()
	createGroupAndRole("dev")
	createGroupAndRole("platform")
}
