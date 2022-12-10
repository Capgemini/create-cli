package keycloak

import (
	"create-cli/internal/k8s"
	"create-cli/internal/log"
	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Keycloak"}
var keycloakUrl, keycloakUrlPresent = os.LookupEnv("KEYCLOAK_URL")

func Keycloak() {

	k8s.InitKubeClient()

	if !keycloakUrlPresent {
		keycloakUrl = "http://keycloak.keycloak:80"
		logger.Actionf("KEYCLOAK_URL env variable is not defined, using internal K8s service URL instead")
	}

	HealthCheck()
	GetKeycloakToken()
	CreateGroups()
	CreateClientAndClientProtocolMappers()
	GetKeycloakPasswords()
}
