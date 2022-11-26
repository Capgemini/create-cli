package sonarqube

import (
	"create-cli/internal/k8s"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type SonarQubeUserTokenResponse struct {
	Token string `json:"token"`
}

var concourseUserToken string

func CreateConcourseUserToken() {
	createUserToken(concourseUsername)
	saveConcourseSonarQubeUserToken(concourseUserToken)
}

func createUserToken(username string) {

	logger.Waitingf("Creating User Token for user: %s...", username)
	client := &http.Client{}
	req, err := http.NewRequest("POST", sonarqubeUrl+"/api/user_tokens/generate", nil)
	if err != nil {
		logger.Failuref("Error creating HTTP request for User Token generation", err)
	}

	basicAuthString := b64.StdEncoding.EncodeToString([]byte("admin:" + sonarAdminPassword))
	req.Header.Add("Authorization", "Basic "+basicAuthString)

	q := req.URL.Query()

	q.Add("login", username)
	q.Add("name", username)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to User Token generation", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Failuref("error reading response body", err)
		panic(err.Error())
	}

	var data SonarQubeUserTokenResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Failuref("Error marshalling response body json to SonarQubeUserTokenResponse object", err)
		panic(err)
	}

	if resp.StatusCode == 503 {
		logger.Failuref("There was an internal error when creating User Token for user: %s...", username)
		return
	}

	if resp.StatusCode == 400 {
		logger.Failuref("Failed to create User Token for user: %s", username)
		return
	}

	if resp.StatusCode == 200 {
		logger.Successf("User token created for user %s", username)
		concourseUserToken = data.Token
	}
}

// saveConcourseSonarQubeUserToken saves the token of the Concourse SonarQube user
// to the kubernetes secrets for use when creating the Concourse Vault secrets later on.
func saveConcourseSonarQubeUserToken(token string) {
	logger.Waitingf("Adding Concourse Sonarqube User Token...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/concourse-sonarqube-user-token",
		Value: b64.StdEncoding.EncodeToString([]byte(token)),
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Concourse Sonarqube user Token added")
}
