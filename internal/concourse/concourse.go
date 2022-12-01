package concourse

import (
	"create-cli/internal/k8s"
	"create-cli/internal/log"

	"os"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Concourse"}
var concourseUrl, concourseUrlPresent = os.LookupEnv("CONCOURSE_URL")
var botGitlabToken, botGitlabTokenPresent = os.LookupEnv("CONCOURSE_GITLAB_TOKEN")

func Concourse() {

	k8s.InitKubeClient()

	if !concourseUrlPresent {
		logger.Warningf("CONCOURSE_URL env variable is not defined. Therefore using internal Concourse Web SVC URL")
		concourseUrl = "http://concourse-web.concourse:80"
	}

	if !botGitlabTokenPresent {
		logger.Failuref("CONCOURSE_GITLAB_TOKEN env variable is not defined. This is needed for Concourse to be able to access Gitlab API's.")
		os.Exit(1)
	}

	Healthcheck()
	GetAccessTokenForConcourse()
	ConfigureMainTeam()
	ConfigureDevTeam()
	CreateConcourseSecrets()
	CreateBootstrapPipeline()
}
