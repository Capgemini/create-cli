package harbor

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HarborSystemRequest struct {
	ProjectCreationRestriction string `json:"project_creation_restriction"`
	RobotTokenDuration         int    `json:"robot_token_duration"`
	RobotNamePrefix            string `json:"robot_name_prefix"`
}

var HarborRobotNamePrefix = "robot-"

func ConfigureHarborSystem() {
	configureHarborSystemRequest := HarborSystemRequest{
		ProjectCreationRestriction: "adminonly",
		RobotTokenDuration:         30,
		RobotNamePrefix:            HarborRobotNamePrefix,
	}

	configureHarborSystemRequestJSON, err := json.Marshal(configureHarborSystemRequest)
	if err != nil {
		logger.Failuref("Error marshalling object to json for request to configure harbor system", err)
	}

	client := &http.Client{}
	logger.Waitingf("Configuring Harbor system...")
	req, err := http.NewRequest("PUT", harborUrl+"/api/v2.0/configurations", bytes.NewBuffer(configureHarborSystemRequestJSON))
	if err != nil {
		logger.Failuref("Error creating HTTP request for configuring of harbor system", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to configure harbor system", err)
	}

	if resp.StatusCode == 200 {
		logger.Successf("Configured Harbor System")
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}
