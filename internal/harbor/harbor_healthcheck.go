package harbor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type HarborHealthCheckResponse struct {
	Status string `json:"status"`
}

func HealthCheck() {
	for {
		logger.Waitingf("Checking if harbor is healthy...")
		client := &http.Client{}
		req, err := http.NewRequest("GET", harborUrl+"/api/v2.0/health", nil)
		if err != nil {
			logger.Failuref("Error creating HTTP request for harbor healthcheck", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Failuref("Error making HTTP request to harbor health endpoint", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			logger.Waitingf("Harbor still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Failuref("error reading healthcheck response body", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		var data HarborHealthCheckResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Failuref("error unmarshalling healthcheck response into object", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if data.Status == "healthy" {
			logger.Successf("Harbor is healthy")
			break
		}
		logger.Waitingf("Harbor still unhealthy, waiting for 5 seconds...")
		time.Sleep(time.Duration(5) * time.Second)
	}
}
