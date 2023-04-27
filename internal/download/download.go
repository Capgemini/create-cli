package download

import (
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
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

	ReturnRepoListWithCloudProviderTemplate(cloudProvider)

	log.Printf("Cloning create-v1...")
	_, err := git.PlainClone(CreateRepositoryDirectory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth: &http.BasicAuth{
			Username: "git",
			Password: personalAccessToken,
		},
		URL:      "https://github.com/cd-create/create-v1.git",
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
