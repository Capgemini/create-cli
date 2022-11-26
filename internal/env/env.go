package env

import (
	"create-cli/internal/log"
	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "EnvVar Retriever"}

func GetEnvVar(envVar string) string {
	var envVarValue, isEnvVarPresent = os.LookupEnv(envVar)
	if !isEnvVarPresent {
		logger.Failuref("%s env variable is not defined, cannot configure OIDC", envVar)
		panic("exiting.")
	}
	return envVarValue
}
