package harbor

import (
	"create-cli/internal/k8s"
	"create-cli/internal/log"
	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Harbor"}
var harborUrl, harborUrlPresent = os.LookupEnv("HARBOR_URL")

func Harbor() {
	k8s.InitKubeClient()

	if !harborUrlPresent {
		logger.Warningf("HARBOR_URL env variable is not defined. Therefore using internal Harbor-Core SVC URL")

		// we set the harborUrl with the `harbor-core` service because that is the SVC
		// that serves requests going to the Harbor API
		harborUrl = "http://harbor-core.harbor:80"
	}

	HealthCheck()
	if ChangeDefaultAdminPassword() {
		return
	}
	CreateCreateProject()
	CreateUserGroupsForCreateProject()
	ConfigureHarborOIDC()
	ConfigureHarborSystem()
	CreateRobotAccounts()
}
