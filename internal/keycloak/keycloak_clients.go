package keycloak

import (
	"bytes"
	"create-cli/internal/generators"
	"create-cli/internal/k8s"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

type Client struct {
	Id                        string   `json:"id,omitempty"`
	Name                      string   `json:"name"`
	ClientId                  string   `json:"clientId"`
	Secret                    string   `json:"secret,omitempty"`
	Enabled                   bool     `json:"enabled"`
	StandardFlowEnabled       bool     `json:"standardFlowEnabled"`
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled"`
	BearerOnly                bool     `json:"bearerOnly"`
	PublicClient              bool     `json:"publicClient"`
	RedirectURIs              []string `json:"redirectUris"`
}

type ClientProtocolMapper struct {
	Name           string               `json:"name"`
	Protocol       string               `json:"protocol"`
	ProtocolMapper string               `json:"protocolMapper"`
	Config         ProtocolMapperConfig `json:"config"`
}

type ProtocolMapperConfig struct {
	AccessTokenClaim       bool   `json:"access.token.claim,omitempty"`
	ClaimName              string `json:"claim.name,omitempty"`
	FullPath               bool   `json:"full.path,omitempty"`
	IdTokenPath            bool   `json:"id.token.path,omitempty"`
	IdTokenClaim           bool   `json:"id.token.claim,omitempty"`
	UserInfoTokenClaim     bool   `json:"userinfo.token.claim,omitempty"`
	IncludedClientAudience string `json:"included.client.audience,omitempty"`
}

func CreateClientAndClientProtocolMapper(name string, secret string, redirectUris []string) {
	client := Client{
		Name:                      name,
		ClientId:                  name,
		Secret:                    secret,
		Enabled:                   true,
		StandardFlowEnabled:       true,
		DirectAccessGrantsEnabled: true,
		BearerOnly:                false,
		PublicClient:              false,
		RedirectURIs:              redirectUris,
	}

	clientJSON, err := json.Marshal(client)
	if err != nil {
		logger.Failuref("Error marshalling client json", err)
		panic(err)
	}

	logger.Waitingf("Creating Keycloak Client %s...", name)
	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", keycloakUrl+"/admin/realms/sso/clients", bytes.NewBuffer(clientJSON))

	if err != nil {
		logger.Failuref("Error creating HTTP request to create Keycloak Client")
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to create Keycloak Client")
		panic(err)
	}

	if resp.StatusCode == 409 {
		logger.Warningf("%s client already created", name)
		return
	}

	if resp.StatusCode != 201 {
		logger.Failuref("There was a problem creating client: %s", name)
		os.Exit(1)
	}

	logger.Successf("%s client created", name)

	// we skip over the `concourse`, `grafana` & `oauth2-proxy` secrets
	// because these secrets are accessible in k8s
	if name != "concourse" && name != "grafana" && name != "oauth2-proxy" {
		createClientSecretForLater(secret, name)
	}

	CreateClientProtocolMapper(name, groupsClientProtocolMapper())

	if name == "oauth2-proxy" {
		CreateClientProtocolMapper(name, audienceClientProtocolMapper(name))
	}
}

// createClientSecretForLater will create the relevant client secrets as k8s secrets
// for other parts of the CLI to use later on. For example, when creating
// the Vault Keycloak Client, we want to ensure that when setting up the
// Vault OIDC later on, it uses the same clientSecret that is created when
// creating the Vault client.
func createClientSecretForLater(clientSecret string, clientId string) {
	logger.Waitingf("Adding %s Client Secret to default k8s secrets...", clientId)
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/" + clientId + "-keycloak-client-secret",
		Value: base64.StdEncoding.EncodeToString([]byte(clientSecret)),
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("%s Client Secret added", clientId)
}

func audienceClientProtocolMapper(name string) ClientProtocolMapper {
	return ClientProtocolMapper{
		Name:           "audience-mapper",
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-audience-mapper",
		Config: ProtocolMapperConfig{
			IncludedClientAudience: name,
			AccessTokenClaim:       true,
		},
	}
}

func groupsClientProtocolMapper() ClientProtocolMapper {
	return ClientProtocolMapper{
		Name:           "group-membership-mapper",
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-group-membership-mapper",
		Config: ProtocolMapperConfig{
			AccessTokenClaim:   true,
			ClaimName:          "groups",
			FullPath:           false,
			IdTokenClaim:       true,
			IdTokenPath:        true,
			UserInfoTokenClaim: true,
		},
	}
}

func CreateClientProtocolMapper(name string, clientProtocolMapper ClientProtocolMapper) {
	clientId := GetClientClientId(name)

	clientProtocolMapperJSON, err := json.Marshal(clientProtocolMapper)
	if err != nil {
		logger.Failuref("Error marshalling object into JSON", err)
		panic(err)
	}

	logger.Waitingf("Creating protocol mapper for client: %s...", name)
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/realms/sso/clients/%s/protocol-mappers/models", keycloakUrl, clientId), bytes.NewBuffer(clientProtocolMapperJSON))

	if err != nil {
		logger.Failuref("Error creating HTTP request to create protocol mapper for client: %s", name)
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to create protocol mapper for client: %s", name)
		panic(err)
	}

	if resp.StatusCode == 409 {
		logger.Warningf("%s client protocol mapper already created \n", name)
		return
	}

	if resp.StatusCode != 201 {
		logger.Failuref("There was a runtime problem creating client protocol mapper for client: %s", name)
		os.Exit(1)
	}

	logger.Successf("%s client protocol mapper created.", name)
}

func GetClientClientId(name string) string {
	logger.Waitingf("Getting ClientID for client: %s...", name)
	data := getClientByClientID(name)
	logger.Successf("Retrieved Client ID")
	return data[0].Id
}

func getConcourseClientSecret() string {
	logger.Waitingf("Retrieving Concourse Keycloak Client Secret")
	secret := k8s.GetSecret("oidc", "concourse")
	logger.Successf("Retrieved Concourse Keycloak Client Secret")
	return string(secret.Data["oidcClientSecret"])
}

func getGrafanaClientSecret() string {
	logger.Waitingf("Retrieving Grafana Keycloak Client Secret")
	secret := k8s.GetSecret("oidc", "prometheus")
	logger.Successf("Retrieved Grafana Keycloak Client Secret")
	return string(secret.Data["clientSecret"])
}

func getOAuth2ProxyClientSecret() string {
	logger.Waitingf("Retrieving OAuth2Proxy Keycloak Client Secret")
	secret := k8s.GetSecret("oidc", "oauth2-proxy")
	logger.Successf("Retrieved OAuth2Proxy Keycloak Client Secret")
	return string(secret.Data["oidcClientSecret"])
}

func CreateClientAndClientProtocolMappers() {
	createUrl := viper.GetString("create-url")

	if clientDoesNotExist("vault") {
		CreateClientAndClientProtocolMapper("vault", generators.GenerateSecret(20, 1, 0, false), []string{
			"https://vault.tooling." + createUrl + "/ui/vault/auth/oidc/oidc/callback",
			"https://vault.tooling." + createUrl + "/oidc/oidc/callback"})
	}

	if clientDoesNotExist("concourse") {
		CreateClientAndClientProtocolMapper("concourse", getConcourseClientSecret(), []string{
			"https://concourse.tooling." + createUrl + "/sky/issuer/callback"})
	}

	if clientDoesNotExist("harbor") {
		CreateClientAndClientProtocolMapper("harbor", generators.GenerateSecret(20, 1, 0, false), []string{
			"https://harbor.tooling." + createUrl + "/c/oidc/callback"})
	}

	if clientDoesNotExist("grafana") {
		CreateClientAndClientProtocolMapper("grafana", getGrafanaClientSecret(), []string{
			"https://grafana.tooling." + createUrl + "/login/generic_oauth"})
	}

	if clientDoesNotExist("oauth2-proxy") {
		CreateClientAndClientProtocolMapper("oauth2-proxy", getOAuth2ProxyClientSecret(), []string{
			"*"})
	}

}

func clientDoesNotExist(clientName string) bool {
	logger.Waitingf("Checking if %s client exists...", clientName)
	clients := getClientByClientID(clientName)
	if len(clients) > 0 {
		logger.Warningf("%s client already exists", clientName)
		return false
	}

	logger.Successf("%s client does not exist", clientName)
	return true
}

func getClientByClientID(clientName string) []Client {
	req, err := http.NewRequest("GET", keycloakUrl+"/admin/realms/sso/clients", nil)
	if err != nil {
		logger.Failuref("Error create HTTP request to get Keycloak Client")
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	// adds query param to search by clientId
	q := req.URL.Query()
	q.Add("clientId", clientName)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to get Keycloak Client", err)
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Failuref("Error reading response body", err)
		panic(err)
	}

	var data []Client
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Failuref("Error marshalling respones body json to Client object", err)
		panic(err)
	}

	return data
}
