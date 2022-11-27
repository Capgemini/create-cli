package concourse

import (
	"context"
	"create-cli/internal/generators"
	"create-cli/internal/gitlab"
	"create-cli/internal/harbor"
	"create-cli/internal/k8s"
	"create-cli/internal/vault"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"

	vaultApi "github.com/hashicorp/vault/api"
)

var concourseSSHPubKey string
var concourseSSHPrivKey string
var concourseSSHDeployKeyID int
var gitlabGroup, gitlabGroupPresent = os.LookupEnv("GITLAB_GROUP")

func CreateConcourseSecrets() {
	if !gitlabGroupPresent {
		logger.Failuref("GITLAB_GROUP env variable is not defined. Please set it")
		os.Exit(1)
	}
	// removes existing Concourse deploy keys
	removeExistingConcourseDeployKeys()

	// creates new Concourse deploy keys
	generateECDSAKeys()
	createConcourseGitlabDeployKey()
	enableConcourseGitlabDeployKey()
	createConcourseVaultSecrets()
}

// removeExistingConcourseDeployKeys removes all existing
// Concourse deploy keys that may have been created before.
// We remove them from the `backstage` project because that
// is where they are created. When you remove them from the project
// they were created on, it removes them from all projects.
func removeExistingConcourseDeployKeys() {

	deployKeys := gitlab.ListDeployKeys(gitlabGroup + "/backstage")
	for _, dk := range deployKeys {
		if dk.Title == "CREATE Concourse Deploy Key" {
			deleteDeployKey(dk.ID, gitlabGroup)
		}
	}
}

func deleteDeployKey(id int, gitlabGroup string) {
	gitlab.DeleteDeployKey(id, []string{
		gitlabGroup + "/backstage",
		gitlabGroup + "/test-project",
		gitlabGroup + "/concourse-tasks",
		gitlabGroup + "/base-helm-chart",
	})

}

// GenerateECDSAKeys generates EC public and private key pair with given size for SSH.
func generateECDSAKeys() {
	// generate private key
	logger.Waitingf("Generating SSH Keys for Concourse...")
	var privateKey *ecdsa.PrivateKey
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		logger.Failuref("Error generating SSH Keys for Concourse", err)
		panic(err)
	}

	var publicKey ssh.PublicKey
	publicKey, err = ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		logger.Failuref("Error creating public SSH Key for Concourse", err)
		panic(err)
	}
	pubBytes := ssh.MarshalAuthorizedKey(publicKey)

	// encode private key
	var bytes []byte
	bytes, err = x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		logger.Failuref("Error marshalling private SSH Key for Concourse", err)
		panic(err)
	}
	privBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: bytes,
	})

	logger.Successf("Generated SSH Keys for Concourse")
	concourseSSHPrivKey = string(privBytes)
	concourseSSHPubKey = string(pubBytes)
}

func createConcourseVaultSecrets() {
	vaultClient := vault.CreateVaultClientForUnsealedInstance(getVaultRootToken())

	concourseHarborUserPassword := getConcourseRobotPasswordIfAlreadyExistsOtherwiseGenerateIt(vaultClient)
	logger.Waitingf("Creating common shared secrets in `secrets/dev-team...`")

	data := map[string]interface{}{
		"data": map[string]interface{}{
			"concourse_harbor_user_username": harbor.GenerateRobotAccountUsername("concourse"),
			"concourse_harbor_user_password": concourseHarborUserPassword,
			"bot_private_token":              botGitlabToken,
			"concourse_private_key":          concourseSSHPrivKey,
			"sonar-ci-token":                 getConcourseSonarQubeUserToken(),
		},
	}

	_, err := vaultClient.Logical().WriteWithContext(context.Background(), "secrets/dev-team/data/shared", data)
	if err != nil {
		logger.Failuref("Error putting secrets to `secrets/dev-team`:", err)
		panic(err)
	}

	logger.Successf("Created common shared secrets in `secrets/dev-team`")
}

func getConcourseRobotPasswordIfAlreadyExistsOtherwiseGenerateIt(vaultClient *vaultApi.Client) string {
	readData, err := vaultClient.Logical().ReadWithContext(context.Background(), "secrets/dev-team/data/shared")
	if err != nil {
		logger.Failuref("Error reading secrets found at `secrets/dev-team/shared`:", err)
		panic(err)
	}

	if readData == nil {
		return generateAndSaveConcourseRobotAccountPassword()
	}

	response := readData.Data["data"].(map[string]interface{})
	password := fmt.Sprintf("%v", response["concourse_harbor_user_password"])

	if password == "" {
		return generateAndSaveConcourseRobotAccountPassword()
	}
	return password
}

func generateAndSaveConcourseRobotAccountPassword() string {
	robotAccountPassword := generators.GenerateSecret(20, 1, 0, false)
	createRobotAccountSecretForLater(robotAccountPassword)
	return robotAccountPassword
}

// createRobotAccountSecretForLater will create the relevant robot account
// password secrets as a k8s secrets for the Harbor module to use when actually
// creating the Robot Account for Concourse. If we did it the other way around
// then we would have to put the Robot Account Password into Vault AFTER we
// create the Robot Account - which is possible, but the Vault Client for
// adding secrets is a bit flaky.
func createRobotAccountSecretForLater(password string) {
	logger.Waitingf("Adding Concourse Robot Account Password...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:    "add",
		Path:  "/data/concourse-robot-account-password",
		Value: base64.StdEncoding.EncodeToString([]byte(password)),
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Concourse Robot Account Password added")
}

func createConcourseGitlabDeployKey() {
	concourseSSHDeployKeyID = gitlab.AddDeployKey(concourseSSHPubKey, "CREATE Concourse Deploy Key", gitlabGroup)
}

func enableConcourseGitlabDeployKey() {
	gitlab.EnableDeployKey(concourseSSHDeployKeyID, []string{
		gitlabGroup + "/backstage",
		gitlabGroup + "/test-project",
		gitlabGroup + "/concourse-tasks",
		gitlabGroup + "/base-helm-chart",
	})
}

func getVaultRootToken() string {
	logger.Waitingf("Retrieving vault root token")
	secret := k8s.GetSecret("create-secrets", "default")
	logger.Successf("Retrieved keycloak admin user password")
	return string(secret.Data["vault-root-token"])
}

func getConcourseSonarQubeUserToken() string {
	logger.Waitingf("Retrieving Concourse SonarQube user token")
	sonarQubeAdminPasswordSecret := k8s.GetSecret("create-secrets", "default")
	logger.Successf("Retrieved Concourse SonarQube user token")
	token := string(sonarQubeAdminPasswordSecret.Data["concourse-sonarqube-user-token"])
	deleteConcourseSonarQubeUserToken()
	return token
}

// deleteConcourseSonarQubeUserToken deletes the sonar token of the Concourse
// user from the default create-secrets because we have now put it into the Vault secret.
// We held it in the create-secrets, simply to be able to hold it temporarily whist we
// wait to put it into Vault.
func deleteConcourseSonarQubeUserToken() {
	logger.Waitingf("Deleting Concourse SonarQube User Token Secret from default k8s secrets...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:   "remove",
		Path: "/data/concourse-sonarqube-user-token",
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Deleted Concourse SonarQube User Token Secret from default k8s secrets")
}
