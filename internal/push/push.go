package push

import (
	"create-cli/internal/download"
	"create-cli/internal/log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Push"}
var gitlabClient *gitlab.Client

func Push() {
	for repoName := range download.ReposToClone {
		pushRepo(repoName)
	}
}

func pushRepo(repoName string) {
	group := viper.GetString("gitlab-group")
	namespaceId := getNamespaceIDOfGroup(group)
	project := createNewRepository(repoName, namespaceId)
	pushCode(project.HTTPURLToRepo, repoName)
}

func pushCode(httpsUrlToRepo string, repoName string) {
	logger.Waitingf("Pushing project %s...", repoName)
	r, err := git.PlainOpen(download.CreateRepositoryDirectory + repoName)
	if err != nil {
		panic(err)
	}

	// delete existing remote that comes from public OSS git repo
	err = r.DeleteRemote("origin")
	if err != nil {
		panic(err)
	}

	// we want to create a new origin remote that points to the place that
	// we want to push the repository
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{httpsUrlToRepo},
	})
	if err != nil {
		panic(err)
	}

	// pushes cloned and configured repo into new git project
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Password: viper.GetString("pat"),
		},
	})
	if err != nil {
		panic(err)
	}
}

// initialises a gitlab client, we could merge this with the one in the gitlab package.
// but make sure not to break the post-install parts as it uses an environment variable
func initialiseGitlabClient() {
	personalAccessTokenPush := viper.GetString("pat")
	logger.Waitingf("Initialising Gitlab Client...")

	// TODO needs to do the custom gitlab-host here
	// gitlab.NewClient("personalAccessTokenPush", gitlab.WithBaseURL("https://"+viper.GetString("gitlab-host")))
	gc, err := gitlab.NewClient(personalAccessTokenPush)
	if err != nil {
		logger.Failuref("Failed to create gitlab client: %v", err)
	}
	logger.Successf("Initialised Gitlab Client")

	gitlabClient = gc
}

// gets the namespace ID of the group given as Gitlab treats users and groups as
// namespaces. therefore in order to create new projects/repositories in groups
// we need the namespace id of it
func getNamespaceIDOfGroup(groupName string) int {
	if gitlabClient == nil {
		initialiseGitlabClient()
	}

	ids, _, err := gitlabClient.Namespaces.SearchNamespace(groupName)
	if err != nil {
		panic(err)
	}

	if len(ids) > 1 {
		panic("more than one namespaces for group was found")
	}

	return ids[0].ID
}

func createNewRepository(projectName string, namespaceId int) *gitlab.Project {
	if gitlabClient == nil {
		initialiseGitlabClient()
	}

	logger.Waitingf("Creating project: %s ...", projectName)
	project, _, err := gitlabClient.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        gitlab.String(projectName),
		NamespaceID: gitlab.Int(namespaceId),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
	})

	if err != nil {
		logger.Failuref("Failed to create project: %s", projectName)
		panic(err)
	}
	logger.Successf("Project Created %s", projectName)

	return project
}
