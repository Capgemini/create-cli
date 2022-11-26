package harbor

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HarborProjectRequest struct {
	ProjectName string   `json:"project_name"`
	Public      bool     `json:"public"`
	Metadata    Metadata `json:"metadata"`
}

type Metadata struct {
	AutoScan string `json:"auto_scan"`
}

func CreateCreateProject() {
	request := HarborProjectRequest{
		ProjectName: "create",
		Public:      false,
		Metadata: Metadata{
			AutoScan: "true",
		},
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		logger.Failuref("Error marshing Harbor Create project request", err)
	}

	client := &http.Client{}
	logger.Waitingf("Creating `create` Harbor project...")
	req, err := http.NewRequest("POST", harborUrl+"/api/v2.0/projects", bytes.NewBuffer(requestJSON))
	if err != nil {
		logger.Failuref("Error creating Harbor project creation HTTP request", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making request to create Harbor project", err)
	}

	if resp.StatusCode == 201 {
		logger.Successf("`create` Harbor project created")
	}

	if resp.StatusCode == 409 {
		logger.Warningf("`create` project already exists")
		return
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}

type ProjectMemberRequest struct {
	MemberGroup MemberGroup `json:"member_group"`
	RoleID      int         `json:"role_id"`
}

type MemberGroup struct {
	GroupName string `json:"group_name"`
	GroupType int    `json:"group_type"`
}

func CreateUserGroupsForCreateProject() {
	createUserGroup(ProjectMemberRequest{
		RoleID: 1, // admin role
		MemberGroup: MemberGroup{
			GroupName: "platform",
			GroupType: 3, // oidc group type
		},
	})
	createUserGroup(ProjectMemberRequest{
		RoleID: 2, // developer role
		MemberGroup: MemberGroup{
			GroupName: "dev",
			GroupType: 3, // oidc group type
		},
	})
}

func createUserGroup(projectMemberRequest ProjectMemberRequest) {
	projectMemberRequestJSON, err := json.Marshal(projectMemberRequest)
	if err != nil {
		logger.Failuref("error marshalling Harbor project member request", err)
	}

	client := &http.Client{}
	logger.Waitingf("Creating %s Harbor group...", projectMemberRequest.MemberGroup.GroupName)
	req, err := http.NewRequest("POST", harborUrl+"/api/v2.0/projects/create/members", bytes.NewBuffer(projectMemberRequestJSON))
	if err != nil {
		logger.Failuref("Error creating project member HTTP request", err)
	}

	req.SetBasicAuth("admin", newPassword) // make this based on the password setting step
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Failuref("Error making HTTP request to create project member", err)
	}

	if resp.StatusCode == 201 {
		logger.Successf("Created %s Harbor group", projectMemberRequest.MemberGroup.GroupName)
	}

	if resp.StatusCode == 409 {
		logger.Warningf("%s Group already exists", projectMemberRequest.MemberGroup.GroupName)
		return
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logger.Failuref("Error making request due to being unauthenticated or unauthorised")
		return
	}
}
