package download

import (
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

var CreateRepositoryDirectory = "create-repositories/"

var ReposToClone = map[string]string{
	"flux-config-for-applications": "https://github.com/cd-create/flux-config-for-applications.git",
	"flux-config-for-tooling":      "https://github.com/cd-create/flux-config-for-tooling.git",
	"test-project":                 "https://github.com/cd-create/test-project.git",
	"concourse-tasks":              "https://github.com/cd-create/concourse-tasks.git",
	"backstage":                    "https://github.com/cd-create/backstage.git",
	"backstage-software-templates": "https://github.com/cd-create/backstage-software-templates.git",
	"base-helm-chart":              "https://github.com/cd-create/base-helm-chart.git",
}

func ReturnRepoListWithCloudProviderTemplate(cloudProvider string) map[string]string {
	if strings.ToLower(cloudProvider) == "azure" {
		ReposToClone["azure-create-platform"] = "https://github.com/cd-create/azure-create-platform.git"
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "aws" {
		ReposToClone["aws-create-platform-template"] = "https://github.com/cd-create/aws-create-platform-template.git"
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "gcp" {
		ReposToClone["gcp-create-platform-template"] = "https://github.com/cd-create/gcp-create-platform-template.git"
		return ReposToClone
	}

	panic("Cloud Provider not valid")
}

func Download() {

	cloudProvider := viper.GetString("cloud-provider")
	personalAccessToken := viper.GetString("pat")
	tag := "1.0.0.0"

	ReturnRepoListWithCloudProviderTemplate(cloudProvider)

	// create the directory to which all repositories will be cloned within
	err := os.Mkdir(CreateRepositoryDirectory, 0755)
	if err != nil {
		panic(err)
	}

	for repoName, repoUrl := range ReposToClone {
		log.Printf("Cloning %s into %s", repoName, CreateRepositoryDirectory)
		r, err := git.PlainClone(CreateRepositoryDirectory+repoName, false, &git.CloneOptions{
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Auth: &http.BasicAuth{
				Username: "git",
				Password: personalAccessToken,
			},
			URL:      repoUrl,
			Progress: os.Stdout,
		})
		if err != nil {
			panic(err)
		}

		// get work tree for cloned repo
		tree, err := r.Worktree()
		if err != nil {
			panic(err)
		}

		checkoutTag(tag, tree)
		deletesMainBranch(r)
		createAndCheckoutMainBranchOffTag(tag, r, tree)
	}
}

// checkoutTag checks out the passed tag
func checkoutTag(tag string, tree *git.Worktree) {
	err := tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/tags/" + tag),
	})
	if err != nil {
		panic(err)
	}
}

// deletesMainBranch deletes the main branch because we want to create a new
// main branch off the specific tag. otherwise, the main branch will have the entire
// history of the repository which in this case we don't want. just the specific version
func deletesMainBranch(r *git.Repository) {
	err := r.Storer.RemoveReference("refs/heads/main")
	if err != nil {
		panic(err)
	}
}

// createAndCheckpoutMainBranchOffTag will create and checkout a main branch off the commit hash reference
// of the tag that is passed in. this is to ensure that the main branch only has the code on the tag.
// otherwise, if you want the code on the 1.0.0.0 tag, without deleting and creating a new main
// branch off the tag, the main branch will have all of the latest code - which in this case is not
// what you want.
func createAndCheckoutMainBranchOffTag(tag string, r *git.Repository, tree *git.Worktree) {
	// we get the commit reference that is at the tip of the tag
	headRef, err := r.Reference(plumbing.ReferenceName("refs/tags/"+tag), false)
	if err != nil {
		panic(err)
	}

	// we create a new branch called `main` off the commit hash from the tag
	ref := plumbing.NewHashReference("refs/heads/main", headRef.Hash())
	err = r.Storer.SetReference(ref)
	if err != nil {
		panic(err)
	}

	// we checkout the `main` branch just created.
	err = tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/main"),
	})
	if err != nil {
		panic(err)
	}
}
