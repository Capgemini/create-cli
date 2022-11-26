package sonarqube

import (
	"create-cli/internal/k8s"
	"create-cli/internal/log"
	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Sonarqube"}
var sonarqubeUrl, sonarqubeUrlPresent = os.LookupEnv("SONARQUBE_URL")
var sonarAdminPassword string

func SonarQube() {
	k8s.InitKubeClient()

	if !sonarqubeUrlPresent {
		logger.Warningf("SONARQUBE_URL env variable is not defined. Therefore using internal SonarQube SVC URL")
		sonarqubeUrl = "http://sonarqube-sonarqube.sonarqube:9000"
	}

	getSonarAdminPassword()
	HealthCheck()
	CreateConcourseUser()
	CreateConcourseUserToken()
}

func getSonarAdminPassword() {
	logger.Waitingf("Retrieving SonarQube admin password")
	sonarQubeAdminPasswordSecret := k8s.GetSecret("sonarqube-sonarqube-admin-password", "sonarqube")
	logger.Successf("Retrieved SonarQube admin password")
	sonarAdminPassword = string(sonarQubeAdminPasswordSecret.Data["password"])
}
