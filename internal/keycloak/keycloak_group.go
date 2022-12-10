package keycloak

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

func CreateGroups() {
	CreateGroup("dev")
	CreateGroup("platform")
	CreateGroup("sonar-administrators")
}

func CreateGroup(group string) {
	postBody, _ := json.Marshal(map[string]string{
		"name": group,
	})
	requestBody := bytes.NewBuffer(postBody)

	logger.Waitingf("Creating keycloak group %s...", group)
	client := &http.Client{}
	req, err := http.NewRequest("POST", keycloakUrl+"/admin/realms/sso/groups", requestBody)

	if err != nil {
		logger.Failuref("Error creating HTTP request for group creation")
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request for group creation")
		panic(err)
	}

	if resp.StatusCode == 409 {
		logger.Warningf("%s group already created", group)
		return
	}

	if resp.StatusCode == 201 {
		logger.Successf("%s group created", group)
		return
	}

	logger.Failuref("There was a runtime problem creating group: %s", group)
	os.Exit(1)
}
