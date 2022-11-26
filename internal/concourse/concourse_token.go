package concourse

import (
	"context"
	"create-cli/internal/k8s"
	"strings"

	"golang.org/x/oauth2"
)

var accessToken string
var adminPassword string

// getConcourseAdminUserPassword retrieves the secret already created in the concourse
// namespace that holds the credentials for the admin user that we want to use.
// Because it is in a specific format we want to get the parts that we care about the most
// which is the password value
// Example:
// If the following is retrieved as the secret, we want to take the `PASSWORD` value
// `USERNAME:PASSWORD`
func getConcourseAdminUserPassword() {
	logger.Waitingf("Retrieving Concourse admin user credentials...")
	secret := k8s.GetSecret("admin-user", "concourse")
	stringToSplit := string(secret.Data["value"])
	adminPassword = strings.Split(stringToSplit, ":")[1]
	logger.Successf("Retrieved Concourse admin user credentials")
}

func GetAccessTokenForConcourse() {
	// Setup OAuth2 authentication
	conf := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",
		Scopes:       []string{"openid", "profile", "federated:id"},
		Endpoint: oauth2.Endpoint{
			TokenURL: concourseUrl + "/sky/issuer/token",
		},
	}

	getConcourseAdminUserPassword()
	logger.Waitingf("Retrieving access token from Concourse...")
	token, err := conf.PasswordCredentialsToken(context.Background(), "admin", adminPassword)
	if err != nil {
		logger.Failuref("Error getting access token from concourse", err)
		panic(err)
	}

	logger.Successf("Access token retrieved")
	accessToken = token.AccessToken
}
