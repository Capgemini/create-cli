package concourse

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

func CreateBootstrapPipeline() {
	yamlFile, err := ioutil.ReadFile("internal/concourse/bootstrap.yaml")
	if err != nil {
		logger.Failuref("Error reading bootstrap pipeline yaml", err)
		panic(err)
	}

	client := &http.Client{}
	logger.Waitingf("Creating bootstrap pipeline..")
	req, err := http.NewRequest("PUT", concourseUrl+"/api/v1/teams/dev-team/pipelines/bootstrap/config", bytes.NewBuffer(yamlFile))
	if err != nil {
		logger.Failuref("Error creating HTTP request for bootstrapping pipeline", err)
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/x-yaml")
	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request for bootstrapping pipeline", err)
		panic(err)
	}

	if resp.StatusCode != 201 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Failuref("Error reading error body: %s", err)
		}

		bodyString := string(bodyBytes)
		logger.Failuref("Error creating bootstrap pipeline due to it probably already existing. %s", bodyString)
		return
	}
	logger.Successf("Created bootstrap pipeline")

	UnpauseBootStrapPipeline()
}

func UnpauseBootStrapPipeline() {
	client := &http.Client{}
	logger.Waitingf("Unpausing bootstrap pipeline..")
	req, err := http.NewRequest("PUT", concourseUrl+"/api/v1/teams/dev-team/pipelines/bootstrap/unpause", nil)
	if err != nil {
		logger.Failuref("Error creating HTTP request for unpausing bootstrap pipeline", err)
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/x-yaml")
	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request for unpausing bootstrap pipeline", err)
		panic(err)
	}

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Failuref("Error reading error body: %s", err)
		}
		bodyString := string(bodyBytes)
		logger.Failuref("Error unpausing bootstrap pipeline. %s", bodyString)
	}
	logger.Successf("Unpaused bootstrap pipeline")
}
