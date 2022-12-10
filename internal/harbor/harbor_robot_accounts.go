package harbor

import (
	"bytes"
	"create-cli/internal/k8s"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type RobotAccountRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Level       string        `json:"level"`
	Secret      string        `json:"secret"`
	Permissions []Permissions `json:"permissions"`
	Duration    int           `json:"duration"`
}

type Permissions struct {
	Access    []Access `json:"access"`
	Kind      string   `json:"kind"`
	Namespace string   `json:"namespace"`
}

type Access struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Effect   string `json:"effect"`
}

// GenerateRobotAccountUsername generates a Harbor Robot Account Username based on name
// to which the account will belong to. Example, if it's a concourse Robot Account. The
// `name` will be `concourse` and the returning robot account username will be `robot-concourse`.
func GenerateRobotAccountUsername(name string) string {
	logger.Generatef("Generated Robot Account Username for: %s", name)
	return fmt.Sprintf("robot-%s", name)
}

func CreateRobotAccounts() {
	// we need to get these secrets from kubernetes if we can, or generate them via sops
	createRobotAccount(createRobotAccountRequestObject("concourse", "Concourse robot account", getConcourseRobotAccountSecret()))
	createRobotAccount(createRobotAccountRequestObject("flux", "Flux robot account", getFluxRobotAccountSecret()))
}

type Wrapper struct {
	Auths map[string]*Auths `json:"auths"`
}

type Auths struct {
	Password string `json:"password"`
}

// getFluxRobotAccountSecret retrieves the dockerfilejson that exists in the flux harbor pull secret
// and then retrieves the password inside of it. there should only be on entry in the json (harbor),
// and because dockerconfigjson has dynamic JSON keys, we have to do some trickery with loops and maps
// to retrieve the password.
func getFluxRobotAccountSecret() string {
	logger.Waitingf("Retrieving Flux Robot Account Secret...")
	secret := k8s.GetSecret("flux-harbor-pull-secret", "flux-system")
	logger.Successf("Retrieved Flux Robot Account Secret")

	var wrapper Wrapper
	if err := json.Unmarshal([]byte(secret.Data[".dockerconfigjson"]), &wrapper); err != nil {
		log.Fatal(err)
	}

	var fluxHarborPassword string
	for _, registry := range wrapper.Auths {
		fluxHarborPassword = registry.Password
		// we break on the first loop because we only expect one registry
		// to be in the dockerconfigjson (which is Harbor)
		break
	}
	return fluxHarborPassword
}

func getConcourseRobotAccountSecret() string {
	logger.Waitingf("Retrieving Concourse Robot Account Secret...")
	secret := k8s.GetSecret("create-secrets", "default")
	logger.Successf("Retrieved Concourse Robot Account Secret")

	//now we have retrieved the secret, we want to delete it
	deleteConcourseRobotAccountSecret()
	return string(secret.Data["concourse-robot-account-password"])
}

func deleteConcourseRobotAccountSecret() {
	logger.Waitingf("Deleting Concourse Robot Account Secret from default k8s secrets...")
	patchSecretRequest := []k8s.PatchSecretRequest{{
		Op:   "remove",
		Path: "/data/concourse-robot-account-password",
	}}
	k8s.PatchOpaqueSecret(patchSecretRequest)
	logger.Successf("Concourse Robot Account Secret Deleted")
}

func createRobotAccount(request RobotAccountRequest) {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{}
	logger.Waitingf("Creating %s robot account...", request.Name)
	req, err := http.NewRequest("POST", harborUrl+"/api/v2.0/robots", bytes.NewBuffer(requestJSON))
	if err != nil {
		logger.Failuref("Error creating HTTP request for robot account creation", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to create %s robot account", request.Name, err)
	}

	if resp.StatusCode == 201 {
		logger.Successf("Created %s robot account", request.Name)

		// we have to patch the robot account secret because for some reason
		// when we provide it on creation of robot account, it doesn't work
		// and generates a random secret. So patching it will ensure its the
		// value we expect.
		id := getRobotAccountId(resp)
		patchRobotAccountSecret(id, request.Name, request.Secret)
	}

	if resp.StatusCode == 409 {
		logger.Warningf("%s robot account already exists", request.Name)
		return
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}

type RobotAccount struct {
	Id int `json:"id,omitempty"`
}

func getRobotAccountId(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Failuref("Error reading response body", err)
		panic(err)
	}

	var data RobotAccount
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Failuref("Error marshalling respones body json to Robot object", err)
		panic(err)
	}

	return strconv.Itoa(data.Id)
}

func patchRobotAccountSecret(id string, name string, secret string) {
	secretData, _ := json.Marshal(map[string]string{
		"secret": secret,
	})
	responseBody := bytes.NewBuffer(secretData)

	client := &http.Client{}
	logger.Waitingf("Patching %s robot account secret...", name)
	req, err := http.NewRequest("PATCH", harborUrl+"/api/v2.0/robots/"+id, responseBody)
	if err != nil {
		logger.Failuref("Error creating HTTP request for Patching robot account secret", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to patch %s robot account secret", name, err)
	}

	if resp.StatusCode == 200 {
		logger.Successf("Patched %s robot account secret", name)
		return
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}

func createRobotAccountRequestObject(name string, description string, secret string) RobotAccountRequest {
	return RobotAccountRequest{
		Name:        name,
		Description: description,
		Level:       "system",
		Secret:      secret,
		Duration:    -1,
		Permissions: []Permissions{
			{
				Namespace: "create",
				Kind:      "project",
				Access: []Access{
					{
						Action:   "push",
						Resource: "repository",
					},
					{
						Action:   "pull",
						Resource: "repository",
					},
					{
						Action:   "delete",
						Resource: "artifact",
					},
					{
						Action:   "read",
						Resource: "helm-chart",
					},
					{
						Action:   "create",
						Resource: "helm-chart-version",
					},
					{
						Action:   "delete",
						Resource: "helm-chart-version",
					},
					{
						Action:   "create",
						Resource: "tag",
					},
					{
						Action:   "delete",
						Resource: "tag",
					},
					{
						Action:   "create",
						Resource: "artifact-label",
					},
					{
						Action:   "create",
						Resource: "scan",
					},
					{
						Action:   "stop",
						Resource: "scan",
					},
					{
						Action:   "list",
						Resource: "artifact",
					},
					{
						Action:   "list",
						Resource: "repository",
					},
				},
			},
		},
	}
}
