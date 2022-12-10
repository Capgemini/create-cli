package vault

import (
	"context"
	"create-cli/internal/k8s"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"
)

var roleId string
var secretId string

// getConcourseRoleIDAndSecretID retrieves the secret already created in the concourse
// namespace that holds the values of the `role_id` and `secret_id` that we want to use
// when creating the Concourse AppRole in Vault. Because it is in a specific format we
// want to get the parts that we care about the most which are the `role_id` and `secret_id`
// Example:
// If the following is retrieved as the secret, we want to take the `ROLE_ID_VALUE` and
// `SECRET_ID_VALUE`.
// `role_id:ROLE_ID_VALUE,secret_id:SECRET_ID_VALUE`
func getConcourseRoleIDAndSecretID() {
	logger.Waitingf("Retrieving Concourse AppRole Secrets...")
	secret := k8s.GetSecret("vault-app-role", "concourse")
	stringToSplit := string(secret.Data["value"])
	strArray := strings.Split(stringToSplit, "\\,")
	roleId = strings.Split(strArray[0], "role_id:")[1]
	secretId = strings.Split(strArray[1], "secret_id:")[1]
	logger.Successf("Retrieved Concourse AppRole Secrets")
}

func createPolicy() {
	policy := `path "secrets/*" {
		capabilities = [ "read" ]
	}
	`

	logger.Waitingf("Creating `concourse` Vault policy...")
	err := vaultClient.Sys().PutPolicy("concourse", policy)
	if err != nil {
		logger.Failuref("Error creating `concourse` Vault policy")
		panic(err)
	}
	logger.Successf("Created `concourse` Vault policy")
}

func enableAppRole() {
	appRoleOptions := &vault.EnableAuthOptions{
		Type: "approle",
	}

	logger.Waitingf("Enabling AppRole identity method...")
	vaultClient.Sys().EnableAuthWithOptions("approle", appRoleOptions)
	logger.Successf("Enabled AppRole identity method")
}

func createConcourseRoleID() {
	logger.Waitingf("Creating concourse app-role role-id...")
	_, err := vaultClient.Logical().Write("auth/approle/role/concourse/role-id", map[string]interface{}{
		"role_id": roleId,
	})
	if err != nil {
		logger.Failuref("Error creating concourse role-id", err)
		panic(err)
	}
	logger.Successf("Concourse role-id created")
}

func createConcourseSecretID() {
	logger.Waitingf("Creating concourse app-role secret-id...")
	_, err := vaultClient.Logical().Write("auth/approle/role/concourse/custom-secret-id", map[string]interface{}{
		"secret_id": secretId,
	})
	if err != nil {
		logger.Failuref("Error creating concourse secret-id", err)
		panic(err)
	}
	logger.Successf("Concourse secret-id created")
}

func checkConcourseAppRoleCreationWithLogin() {
	logger.Waitingf("Checking concourse app-role login...")
	appRoleAuth, err := auth.NewAppRoleAuth(roleId, &auth.SecretID{FromString: secretId})
	if err != nil {
		logger.Failuref("Couldn't get new approle auth for login check", err)
		panic(err)
	}

	vaultClientForLogin := createTemporaryVaultClient()
	secret, err := vaultClientForLogin.Auth().Login(context.TODO(), appRoleAuth)
	if err != nil {
		logger.Failuref("Couldn't login with new concourse app-role", err)
		panic(err)
	}
	if secret.Auth.ClientToken == "" {
		logger.Failuref("expected a successful login", err)
		os.Exit(1)
	}
	logger.Successf("Login successful")
}

func createConcourseAppRole() {
	logger.Waitingf("Creating concourse app-role...")
	_, err := vaultClient.Logical().Write("auth/approle/role/concourse", map[string]interface{}{
		"backend":                 "approle",
		"role_name":               "concourse",
		"token_policies":          []string{"concourse"},
		"token_no_default_policy": "true",
		"bind_secret_id":          "true",
		"token_period":            "0",
	})
	if err != nil {
		logger.Failuref("Error creating Concourse app-role", err)
		panic(err)
	}

	createConcourseRoleID()
	createConcourseSecretID()
}

func CreateAppRole() {
	getConcourseRoleIDAndSecretID()
	createPolicy()
	enableAppRole()
	createConcourseAppRole()
	checkConcourseAppRoleCreationWithLogin()
}

// createVaultClient creates a temporary vault client just for the purpose of testing the
// concourse app role login. We create a new temporary one because reusing the existin
// vault client will override its tokens and cause errors when using it in future.
func createTemporaryVaultClient() *vault.Client {
	config := vault.DefaultConfig()
	config.Address = vaultUrl
	newVaultClient, err := vault.NewClient(config)
	if err != nil {
		logger.Failuref("unable to initialize vault client: %w", err)
		panic(err)
	}

	return newVaultClient
}
