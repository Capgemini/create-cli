package concourse

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TeamConfig struct {
	Auth Auth `json:"auth,omitempty"`
}

type Auth struct {
	Owner  *Owner  `json:"owner,omitempty"`
	Member *Member `json:"member,omitempty"`
}

type Member struct {
	Groups []string `json:"groups,omitempty"`
}

type Owner struct {
	Groups []string `json:"groups,omitempty"`
	Users  []string `json:"users,omitempty"`
}

func ConfigureMainTeam() {
	ConfigureTeam("main", TeamConfig{
		Auth: Auth{
			Owner: &Owner{
				Groups: []string{
					"oidc:platform",
				},
				Users: []string{
					"local:admin",
				},
			},
		},
	})
}

func ConfigureDevTeam() {
	ConfigureTeam("dev-team", TeamConfig{
		Auth: Auth{
			Member: &Member{
				Groups: []string{
					"oidc:dev",
				},
			},
		},
	})
}

func ConfigureTeam(teamName string, teamConfig TeamConfig) {
	teamConfigJSON, err := json.Marshal(teamConfig)
	if err != nil {
		logger.Failuref("Error marshalling team config object to json", err)
		panic(err)
	}

	client := &http.Client{}
	logger.Waitingf("Creating %s and configuring OIDC and RBAC...", teamName)
	req, err := http.NewRequest("PUT", concourseUrl+"/api/v1/teams/"+teamName, bytes.NewBuffer(teamConfigJSON))
	if err != nil {
		logger.Failuref("Error creating HTTP request for configuring OIDC and RBAC", err)
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request for creating and configuring OIDC and RBAC", err)
		panic(err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		logger.Failuref("Error creating and configuring %s OIDC and RBAC", teamName)
	}

	logger.Successf("Created %s and configured OIDC and RBAC", teamName)
}
