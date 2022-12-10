package sonarqube

import (
	"create-cli/internal/generators"
	"create-cli/internal/k8s"
	b64 "encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var concourseUsername = "concourse"

func CreateConcourseUser() {
	concourseSonarqubeUserPassword := generators.GenerateSecret(20, 1, 0, false)
	_, err := createUser(concourseUsername, concourseSonarqubeUserPassword)
	if err != nil {
		os.Exit(1)
	}

	saveConcourseSonarQubeUserPassword(concourseUsername, concourseSonarqubeUserPassword)
}

func createUser(username string, password string) (string, error) {
	logger.Waitingf("Creating local %s user...", username)
	client := &http.Client{}
	req, err := http.NewRequest("POST", sonarqubeUrl+"/api/users/create", nil)
	if err != nil {
		logger.Failuref("Error creating HTTP request for SonarQube user creation", err)
	}

	basicAuthString := b64.StdEncoding.EncodeToString([]byte("admin:" + sonarAdminPassword))
	req.Header.Add("Authorization", "Basic "+basicAuthString)

	q := req.URL.Query()
	q.Add("login", username)
	q.Add("name", username)
	q.Add("password", password)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to SonarQube user creation", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Failuref("error reading response body", err)
		panic(err.Error())
	}

	if resp.StatusCode == 503 {
		logger.Failuref("There was an internal error when creating %s user...", username)
		return "", errors.New("")
	}

	if resp.StatusCode == 400 {
		// we check if the body has an `already exists` string because if so
		// then user already has been created. we do this because the status code
		// returned is 400, not 409 - which isn't ideal.
		if strings.Contains(string(body), "already exists") {
			logger.Failuref("%s user already exists", username)
			return "", nil
		}
		logger.Failuref("Failed to create %s user: %s", username, string(body))
		return "", errors.New("")
	}

	if resp.StatusCode == 200 {
		logger.Successf("%s user created", username)
		return "", nil
	}

	return "", nil
}

// saveConcourseSonarQubeUserPassword saves the password of the Concourse SonarQube user
// to the kubernetes secrets for CREATE admins to use in future if needed.
func saveConcourseSonarQubeUserPassword(username string, password string) {
	logger.Waitingf("Adding %s Sonarqube User Password...", username)
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/concourse-sonarqube-user-password",
		Value: b64.StdEncoding.EncodeToString([]byte(password)),
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("%s Sonarqube user Password added", username)
}
