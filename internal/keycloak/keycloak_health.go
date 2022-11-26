package keycloak

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type KeycloakHealthCheckResponse struct {
	Status string `json:"status"`
}

func HealthCheck() {
	for {
		logger.Waitingf("Checking if Keycloak is healthy...")
		client := &http.Client{}
		req, err := http.NewRequest("GET", keycloakUrl+"/health/ready", nil)
		if err != nil {
			logger.Failuref("Error creating HTTP request for Keycloak healthcheck", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Failuref("Error making HTTP request to Keycloak health endpoint", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			logger.Waitingf("Keycloak still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Failuref("error reading healthcheck response body", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		var data KeycloakHealthCheckResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Failuref("error unmarshalling healthcheck response into object", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if data.Status == "UP" {
			logger.Successf("Keycloak is healthy")
			break
		}

		logger.Waitingf("Keycloak still unhealthy, waiting for 5 seconds...")
		time.Sleep(time.Duration(5) * time.Second)
	}
}
