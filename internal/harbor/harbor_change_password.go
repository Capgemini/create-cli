package harbor

import (
	"bytes"
	"create-cli/internal/generators"
	"create-cli/internal/k8s"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type DefaultAdminPassChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

var newPassword string
var oldPassword = "Harbor12345" // this is the default when harbor is installed

func ChangeDefaultAdminPassword() bool {
	logger.Waitingf("Using secret generator to generate new admin password...")

	newPassword = generators.GenerateSecret(20, 1, 0, false)

	request := DefaultAdminPassChangeRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		logger.Failuref("Error marshalling Harbor change admin password request", err)
	}

	client := &http.Client{}
	logger.Waitingf("Changing default password of Harbor admin user...")
	req, err := http.NewRequest("PUT", harborUrl+"/api/v2.0/users/1/password", bytes.NewBuffer(requestJSON))
	if err != nil {
		logger.Failuref("Error creating HTTP request for changing of admin user default password", err)
	}

	req.SetBasicAuth("admin", oldPassword)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making request to change admin user default password", err)
	}

	if resp.StatusCode == 400 {
		logger.Failuref("Bad password supplied")
		return true
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Warningf("The `old_password` provided in request is not correct, meaning that the default password has already been changed")
		return true
	}

	if resp.StatusCode == 200 {
		logger.Successf("Admin user password updated")

		logger.Waitingf("Adding new admin password to default k8s secrets...")
		patchSecretRequest := []k8s.PatchSecretRequest{{
			Op:    "add",
			Path:  "/data/harbor-admin-password",
			Value: base64.StdEncoding.EncodeToString([]byte(newPassword)),
		}}
		k8s.PatchOpaqueSecret(patchSecretRequest)
		logger.Successf("Admin password added")
	}

	return false
}
