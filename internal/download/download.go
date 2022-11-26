package download

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

var CreateRepositoryDirectory = "create-repositories/"

var ReposToClone = map[string]string{
	"flux-config-for-applications": "https://gitlab.com/cap-osce/create-oss/flux-config-for-applications.git",
	"flux-config-for-tooling":      "https://gitlab.com/cap-osce/create-oss/flux-config-for-tooling.git",
	"test-project":                 "https://gitlab.com/cap-osce/create-oss/test-project.git",
	"concourse-tasks":              "https://gitlab.com/cap-osce/create-oss/concourse-tasks.git",
	"backstage":                    "https://gitlab.com/cap-osce/create-oss/backstage.git",
	"backstage-software-templates": "https://gitlab.com/cap-osce/create-oss/backstage-software-templates.git",
	"base-helm-chart":              "https://gitlab.com/cap-osce/create-oss/base-helm-chart.git",
}

func returnRepoListWithCloudProviderTemplate(cloudProvider string) map[string]string {
	if strings.ToLower(cloudProvider) == "azure" {
		ReposToClone["azure-create-platform"] = "https://gitlab.com/cap-osce/create-oss/azure-create-platform.git"
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "aws" {
		ReposToClone["aws-create-platform-template"] = "https://gitlab.com/cap-osce/create-oss/aws-create-platform-template.git"
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "gcp" {
		ReposToClone["gcp-create-platform-template"] = "https://gitlab.com/cap-osce/create-oss/gcp-create-platform-template.git"
		return ReposToClone
	}

	panic("Cloud Provider not valid")
}

func Download() {
	cloudProvider := viper.GetString("cloud-provider")
	personalAccessToken := viper.GetString("pat")

	returnRepoListWithCloudProviderTemplate(cloudProvider)

	// create the directory to which all repositories will be cloned within
	err := os.Mkdir(CreateRepositoryDirectory, 0755)
	if err != nil {
		panic(err)
	}

	for repoName, repoUrl := range ReposToClone {
		_, err := git.PlainClone(CreateRepositoryDirectory+repoName, false, &git.CloneOptions{
			// The intended use of a GitHub / GitLab personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Auth: &http.BasicAuth{
				Password: personalAccessToken,
			},
			URL:      repoUrl,
			Progress: os.Stdout,
		})

		if err != nil {
			panic(err)
		}
	}

}
