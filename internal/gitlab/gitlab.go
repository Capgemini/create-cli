package gitlab

import (
	"create-cli/internal/log"
	"os"

	"github.com/xanzy/go-gitlab"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Gitlab"}
var gitlabClient *gitlab.Client

func InitialiseGitlabClient() {
	gitlabToken, gitlabTokenPresent := os.LookupEnv("GITLAB_TOKEN")
	if !gitlabTokenPresent {
		logger.Failuref("GITLAB_TOKEN env variable is not defined")
		os.Exit(1)
	}

	logger.Waitingf("Initialising Gitlab Client...")
	gc, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		logger.Failuref("Failed to create gitlab client: %v", err)
	}
	logger.Successf("Initialised Gitlab Client")
	gitlabClient = gc
}

func ListDeployKeys(project string) []*gitlab.ProjectDeployKey {
	if gitlabClient == nil {
		InitialiseGitlabClient()
	}

	logger.Waitingf("Getting deploy keys...")
	deployKeys, _, err := gitlabClient.DeployKeys.ListProjectDeployKeys(project, nil)

	if err != nil {
		logger.Failuref("Failed to list deploy keys")
		panic(err)
	}

	return deployKeys
}

func AddDeployKey(concourseSSHPubKey string, deployKeyTitle string, gitlabGroup string) int {
	if gitlabClient == nil {
		InitialiseGitlabClient()
	}

	logger.Waitingf("Adding deploy key: %s...", deployKeyTitle)
	deployKey, _, err := gitlabClient.DeployKeys.AddDeployKey(gitlabGroup+"/backstage", &gitlab.AddDeployKeyOptions{
		Title:   gitlab.String(deployKeyTitle),
		Key:     gitlab.String(concourseSSHPubKey),
		CanPush: gitlab.Bool(false),
	}, nil)

	if err != nil {
		logger.Failuref("Failed to create deploy key: %v", deployKeyTitle)
		panic(err)
	}

	logger.Successf("Added deploy key: %s", deployKeyTitle)
	return deployKey.ID
}

func DeleteDeployKey(deployKeyId int, projects []string) {
	if gitlabClient == nil {
		InitialiseGitlabClient()
	}

	logger.Waitingf("Deleting deploy key: %d on listed projects...", deployKeyId)
	for _, project := range projects {
		resp, err := gitlabClient.DeployKeys.DeleteDeployKey(project, deployKeyId, nil)
		if err != nil {
			if resp.StatusCode != 404 {
				logger.Failuref("Failed to delete deploy key: %d on project: %s", deployKeyId, project)
				panic(err)
			}
			// we skip if response was 404 because the deploy key didn't exist to delete.
			// so there is nothing further to do. so we continue to next project.
			continue
		}
	}
	logger.Successf("Deleted deploy key: %d on listed projects", deployKeyId)
}

func EnableDeployKey(deployKeyId int, projects []string) {
	if gitlabClient == nil {
		InitialiseGitlabClient()
	}

	logger.Waitingf("Enabling deploy key: %d on listed projects...", deployKeyId)
	for _, project := range projects {
		_, _, err := gitlabClient.DeployKeys.EnableDeployKey(project, deployKeyId, nil)
		if err != nil {
			logger.Failuref("Failed to enable deploy key: %d on project: %s", deployKeyId, project)
			panic(err)
		}
		logger.Successf("Enabled deploy key on %s", project)
	}

	logger.Successf("Enabled deploy key: %d on listed projects", deployKeyId)
}
