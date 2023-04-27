package push

import (
	"create-cli/internal/download"
	"create-cli/internal/log"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Push"}
var gitlabClient *gitlab.Client
var gitlabGroup string
var namespaceId int

func Push() {
	gitlabGroup = viper.GetString("gitlab-group")
	namespaceId = getNamespaceIDOfGroup()

	for repoName := range download.ReturnRepoListWithCloudProviderTemplate(viper.GetString("cloud-provider")) {
		pushRepo(repoName)
	}
}

func pushRepo(repoName string) {
	project := createNewRepository(repoName)

	logger.Waitingf("Initialising Git project %s...", repoName)

	// we have to do a `plainOpen` so the git library can perform actions on it that we need below.
	r, err := git.PlainInit(download.CreateRepositoryDirectory+repoName, false)
	if err != nil {
		panic(err)
	}

	// we want to create a new origin remote that points to the
	// new place that we want to push the repository
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{project.HTTPURLToRepo},
	})
	if err != nil {
		panic(err)
	}

	// need to commit the code in the repositories as there has been changes in the `configure` step
	w, err := r.Worktree()
	if err != nil {
		panic(err)
	}
	_, err = w.Add(".")
	if err != nil {
		panic(err)
	}

	commit, err := w.Commit("Configures repository", &git.CommitOptions{
		Author: &object.Signature{
			Name: "create-cli",
			When: time.Now(),
		},
	})
	if err != nil {
		panic(err)
	}

	_, err = r.CommitObject(commit)
	if err != nil {
		panic(err)
	}

	// pushes cloned and configured repo into new git project
	logger.Waitingf("Pushing project %s...", repoName)
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Password: viper.GetString("pat"),
		},
	})
	if err != nil {
		panic(err)
	}

	if repoName == "backstage" || repoName == "base-helm-chart" {
		tag := "1.0.0"
		created, err := createTag(r, tag)
		if err != nil {
			logger.Failuref("create tag error: %s", err)
			return
		}

		if created {
			err = pushTag(r, tag)
			if err != nil {
				logger.Failuref("push tag error: %s", err)
				return
			}
		}
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
func getNamespaceIDOfGroup() int {
	if gitlabClient == nil {
		initialiseGitlabClient()
	}

	ids, _, err := gitlabClient.Namespaces.SearchNamespace(gitlabGroup)
	if err != nil {
		panic(err)
	}

	if len(ids) > 1 {
		panic("more than one namespaces for group was found")
	}

	return ids[0].ID
}

// createNewRepository creates the new Gitlab repository with the name passed in as `projectName`.
// This function will return the Gitlab.Project data object with all relevant details returned from Gitlab API.
func createNewRepository(projectName string) *gitlab.Project {
	if gitlabClient == nil {
		initialiseGitlabClient()
	}

	logger.Waitingf("Creating project: %s ...", projectName)
	project, r, err := gitlabClient.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        gitlab.String(projectName),
		NamespaceID: gitlab.Int(namespaceId),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
	})

	if r.StatusCode == 400 && err != nil {
		logger.Successf("Project already created.")
		panic(err)
	}

	if err != nil {
		logger.Failuref("Failed to create project: %s", projectName)
		panic(err)
	}
	logger.Successf("Project Created %s", projectName)

	return project
}

func createTag(r *git.Repository, tag string) (bool, error) {
	h, err := r.Head()
	if err != nil {
		logger.Failuref("get HEAD error: %s", err)
		return false, err
	}
	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
	})

	if err != nil {
		logger.Failuref("create tag error: %s", err)
		return false, err
	}

	return true, nil
}

func pushTag(r *git.Repository, tag string) error {
	tagRefSpec := fmt.Sprintf("refs/tags/%[1]s:refs/tags/%[1]s", tag)
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec(tagRefSpec)},
		Auth: &http.BasicAuth{
			Password: viper.GetString("pat"),
		},
	}
	err := r.Push(po)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			logger.Failuref("origin remote was up to date, no push done")
			return nil
		}
		logger.Failuref("push to remote origin error: %s", err)
		return err
	}

	return nil
}
