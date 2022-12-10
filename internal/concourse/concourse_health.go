package concourse

import (
	"net/http"
	"time"
)

func Healthcheck() {
	for {
		logger.Waitingf("Checking if concourse is healthy...")
		client := &http.Client{}
		req, err := http.NewRequest("GET", concourseUrl+"/api/v1/info", nil)
		if err != nil {
			logger.Failuref("Error creating HTTP request for concourse health check endpoint", err)
			logger.Waitingf("Concourse still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Failuref("Error making HTTP request for concourse health check endpoint", err)
			logger.Waitingf("Concourse still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if resp.StatusCode == 200 {
			logger.Successf("Concourse is healthy")
			break
		}
		logger.Waitingf("Concourse still unhealthy, waiting for 5 seconds...")
		time.Sleep(time.Duration(5) * time.Second)
	}
}
