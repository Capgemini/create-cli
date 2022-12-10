package sonarqube

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type SonarQubeHealthCheckResponse struct {
	Health string `json:"health"`
}

func HealthCheck() {
	for {
		logger.Waitingf("Checking if SonarQube is healthy...")
		client := &http.Client{}
		req, err := http.NewRequest("GET", sonarqubeUrl+"/api/system/health", nil)
		if err != nil {
			logger.Failuref("Error creating HTTP request for SonarQube healthcheck", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		basicAuthString := b64.StdEncoding.EncodeToString([]byte("admin:" + sonarAdminPassword))
		req.Header.Add("Authorization", "Basic "+basicAuthString)
		resp, err := client.Do(req)
		if err != nil {
			logger.Failuref("Error making HTTP request to SonarQube health endpoint", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			logger.Waitingf("SonarQube still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Failuref("error reading healthcheck response body", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		var data SonarQubeHealthCheckResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Failuref("error unmarshalling healthcheck response into object", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if data.Health == "GREEN" {
			logger.Successf("SonarQube is healthy")
			break
		}
		logger.Waitingf("SonarQube still unhealthy, waiting for 5 seconds...")
		time.Sleep(time.Duration(5) * time.Second)
	}
}
