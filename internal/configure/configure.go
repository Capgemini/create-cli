package configure

import (
	"create-cli/internal/generators"
	"create-cli/internal/harbor"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"k8s.io/utils/strings/slices"
)

var CreateRepositoryDirectory = "create-repositories/"

var ReposToClone = []string{
	"flux-config-for-applications",
	"flux-config-for-tooling",
	"test-project",
	"concourse-tasks",
	"backstage",
	"backstage-software-templates",
	"base-helm-chart",
}

func ReturnRepoListWithCloudProviderTemplate(cloudProvider string) []string {
	if strings.ToLower(cloudProvider) == "azure" {
		ReposToClone = append(ReposToClone, "azure-create-platform")
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "aws" {
		ReposToClone = append(ReposToClone, "aws-create-platform-template")
		return ReposToClone
	}

	if strings.ToLower(cloudProvider) == "gcp" {
		ReposToClone = append(ReposToClone, "gcp-create-platform")
		return ReposToClone
	}

	panic("Cloud Provider not valid")
}

var createUrl string
var acmeRegistrationEmail string
var backstageGitlabUserToken string
var concourseGitlabUserToken string
var fluxHarborRobotAccountUsername string
var fluxHarborRobotAccountUsernameB64 string
var fluxHarborRobotAccountPassword string
var fluxHarborRobotAccountPasswordB64 string
var fluxPullSecretAuthB64 string
var fluxPullSecretJsonB64 string
var concourseAdminUsername = "admin"
var concourseAdminPassword string
var concourseOIDClientSecret string
var concourseAppRoleCreds string
var oAuth2ProxyOIDCClientSecret string
var grafanaOIDCClientSecret string
var sonarqubeNewAdminPassword string
var gitSSHURLCreateProject string
var gitHTTPSURLCreateProject string
var gitlabGroup string
var gitlabPATToken string

// replaceString takes in file contents as a string and finds and replaces all
// instances of an old string with the new one
func replaceString(fileContents string, newString string, oldString string) string {
	return strings.Replace(fileContents, newString, oldString, -1)
}

// simply base64 encodes a string.
func base64EncodeString(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// generates a random 122 bit UUID, otherwise will throw an error
func generateUUID() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}

func replaceTextInFile(fileContents string, path string) {
	fileContents = replaceString(fileContents, "[CREATE_URL]", createUrl)
	fileContents = replaceString(fileContents, "[FLUX_HARBOR_ROBOT_ACCOUNT_USERNAME]", fluxHarborRobotAccountUsernameB64)
	fileContents = replaceString(fileContents, "[FLUX_HARBOR_ROBOT_ACCOUNT_PASSWORD]", fluxHarborRobotAccountPasswordB64)
	fileContents = replaceString(fileContents, "[FLUX_HARBOR_PULL_SECRET]", fluxPullSecretJsonB64)
	fileContents = replaceString(fileContents, "[ACME_REGISTRATION_EMAIL]", acmeRegistrationEmail)
	fileContents = replaceString(fileContents, "[BACKSTAGE_GITLAB_TOKEN]", base64EncodeString(backstageGitlabUserToken))
	fileContents = replaceString(fileContents, "[CONCOURSE_ADMIN_PASSWORD]", base64EncodeString(concourseAdminPassword))
	fileContents = replaceString(fileContents, "[CONCOURSE_ADMIN_CREDS]", base64EncodeString(concourseAdminUsername+":"+concourseAdminPassword))
	fileContents = replaceString(fileContents, "[CONCOURSE_OIDC_CLIENT_SECRET]", base64EncodeString(concourseOIDClientSecret))
	fileContents = replaceString(fileContents, "[CONCOURSE_VAULT_APP_ROLE]", base64EncodeString(concourseAppRoleCreds))
	fileContents = replaceString(fileContents, "[CONCOURSE_GITLAB_TOKEN]", base64EncodeString(concourseGitlabUserToken))
	fileContents = replaceString(fileContents, "[OAUTH2_PROXY_OIDC_CLIENT_SECRET]", base64EncodeString(oAuth2ProxyOIDCClientSecret))
	fileContents = replaceString(fileContents, "[GITLAB_TOKEN]", base64EncodeString(gitlabPATToken))
	fileContents = replaceString(fileContents, "[GRAFANA_OIDC_CLIENT_SECRET]", base64EncodeString(grafanaOIDCClientSecret))
	fileContents = replaceString(fileContents, "[SONARQUBE_NEW_ADMIN_PASSWORD]", base64EncodeString(sonarqubeNewAdminPassword))
	fileContents = replaceString(fileContents, "[GIT_SSH_URL_CREATE_PROJECT]", gitSSHURLCreateProject)
	fileContents = replaceString(fileContents, "[GIT_HTTPS_URL_CREATE_PROJECT]", gitHTTPSURLCreateProject)
	fileContents = replaceString(fileContents, "[GITLAB_GROUP]", gitlabGroup)

	err := ioutil.WriteFile(path, []byte(fileContents), 0)
	if err != nil {
		panic(err)
	}
}

// https://gist.github.com/tdegrunt/045f6b3377f3f7ffa408
// Replace some text in a bunch of files with golang
func visit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	// easier to allow all file types instead of having to declare
	// the ones we care about - as there are so many types.
	matched, err := filepath.Match("*.*", fi.Name())

	if err != nil {
		panic(err)
	}

	if matched {
		read, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		// fmt.Printf("Walked over file: %s \n", path)
		replaceTextInFile(string(read), path)
	}

	return nil
}

func Configure() {
	cloudProvider := viper.GetString("cloud-provider")
	gitlabPATToken = viper.GetString("gitlab-pat-token")
	createUrl = viper.GetString("create-url")
	acmeRegistrationEmail = viper.GetString("acme-reg-email")
	backstageGitlabUserToken = viper.GetString("backstage-gitlab-token")
	concourseGitlabUserToken = viper.GetString("concourse-gitlab-token")
	gitlabHost := viper.GetString("gitlab-host")
	gitlabGroup = viper.GetString("gitlab-group")

	fluxHarborRobotAccountUsername = harbor.HarborRobotNamePrefix + "flux"
	fluxHarborRobotAccountUsernameB64 = base64EncodeString(fluxHarborRobotAccountUsername)
	fluxHarborRobotAccountPassword = generators.GenerateSecret(30, 1, 0, false)
	fluxHarborRobotAccountPasswordB64 = base64EncodeString(fluxHarborRobotAccountPassword)
	fluxPullSecretAuthB64 = base64EncodeString(fluxHarborRobotAccountUsername + ":" + fluxHarborRobotAccountPassword)
	fluxPullSecretJsonB64 = base64EncodeString(fmt.Sprintf(`{"auths":{"harbor.tooling.%s":{"auth":"%s","username":"%s","password":"%s"}}}`, createUrl, fluxPullSecretAuthB64, fluxHarborRobotAccountUsername, fluxHarborRobotAccountPassword))
	concourseAdminPassword = generators.GenerateSecret(10, 1, 0, false)
	concourseOIDClientSecret = generators.GenerateSecret(32, 1, 0, false)
	concourseAppRoleCreds = fmt.Sprintf(`role_id:%s\,secret_id:%s`, generateUUID(), generateUUID())
	oAuth2ProxyOIDCClientSecret = generators.GenerateSecret(32, 1, 0, false)
	grafanaOIDCClientSecret = generators.GenerateSecret(32, 1, 0, false)
	sonarqubeNewAdminPassword = generators.GenerateSecret(20, 1, 0, false)

	gitSSHURLCreateProject = fmt.Sprintf("ssh://git@%s/%s", gitlabHost, gitlabGroup)
	gitHTTPSURLCreateProject = fmt.Sprintf("https://%s/%s", gitlabHost, gitlabGroup)

	removeUnrequiredCloudRepos(cloudProvider)

	log.Println("Configuring CREATE repositories...")
	err := filepath.Walk(CreateRepositoryDirectory, visit)
	if err != nil {
		panic(err)
	}
	log.Println("Configured.")
}

func removeUnrequiredCloudRepos(cloudProvider string) {
	entries, err := os.ReadDir(CreateRepositoryDirectory)
	if err != nil {
		log.Fatal(err)
	}

	// we iterate over the list of directories inside `create-repositories`
	// and we remove the cloud directories that are not needed.
	// example, if a user has chosen to run the CLI with the azure cloud provider flag set
	// then we want to remove the GCP and AWS repos that we do not need to configure
	log.Println("Removing unused Cloud Repositories...")
	for _, e := range entries {
		if e.IsDir() && e.Name() != ".git" {
			if !slices.Contains(ReturnRepoListWithCloudProviderTemplate(cloudProvider), e.Name()) {
				err := os.RemoveAll(CreateRepositoryDirectory + e.Name())
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	log.Println("Unused Cloud Repositories removed.")
}
