package download

import (
	"create-cli/internal/configure"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

func Download() {

	personalAccessToken := viper.GetString("pat")

	log.Printf("Cloning create-v1...")
	_, err := git.PlainClone(configure.CreateRepositoryDirectory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth: &http.BasicAuth{
			Username: "git",
			Password: personalAccessToken,
		},
		URL:      "https://github.com/Capgemini/create-v1.git",
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
