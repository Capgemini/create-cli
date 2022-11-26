package keycloak

import (
	"create-cli/internal/k8s"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type KeycloakTokenResponse struct {
	AccessToken string `json:"access_token"`
}

var accessToken string

func GetKeycloakToken() {

	logger.Waitingf("Retrieving Keycloak admin password from keycloak secret...")
	secret := k8s.GetSecret("keycloak", "keycloak")
	keyCloakAdminPassword := string(secret.Data["admin-password"])
	logger.Successf("Retrieved Keycloak admin password from keycloak secret")

	payload := strings.NewReader(fmt.Sprintf("username=user&password=%s&client_id=admin-cli&grant_type=password", keyCloakAdminPassword))

	logger.Waitingf("Retrieving Keycloak access token...")
	client := &http.Client{}
	req, err := http.NewRequest("POST", keycloakUrl+"/realms/master/protocol/openid-connect/token", payload)
	if err != nil {
		logger.Failuref("Error creating HTTP request to get keycloak access token")
		panic(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to get keycloak access token")
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Failuref("Error reading HTTP response body when getting keycloak access token")
		panic(err.Error())
	}

	var data KeycloakTokenResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Failuref("Error marhsalling HTTP response body to object when getting keycloak access token")
		panic(err)
	}

	logger.Successf("Retrieved Keycloak access token")
	accessToken = data.AccessToken
}
