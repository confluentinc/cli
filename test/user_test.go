package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestUserList() {
	tests := []CLITest{
		{
			args:    "admin user list",
			fixture: "admin/user-list.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestUserDescribe() {
	tests := []CLITest{
		{
			args:    		"admin user describe u-0",
			wantErrCode: 	1,
			fixture: 		"admin/user-resource-not-found.golden",
		},
		{
			args:			"admin user describe u-17",
			fixture: 		"admin/user-describe.golden",
		},
		{
			args:       	"admin user describe 0",
			wantErrCode: 	1,
			fixture:     	"admin/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestUserDelete() {
	tests := []CLITest{
		{
			args:    "admin user delete u-0",
			fixture: "admin/user-delete.golden",
		},
		{
			args:        "admin user delete 0",
			wantErrCode: 1,
			fixture:     "admin/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestUserInvite() {
	tests := []CLITest{
		{
			args:    "admin user invite miles@confluent.io",
			fixture: "admin/user-invite.golden",
		},
		{
			args:        "admin user invite bad-email.com",
			wantErrCode: 1,
			fixture:     "admin/user-bad-email.golden",
		},
		{
			args:        "admin user invite test@error.io",
			wantErrCode: 1,
			fixture:     "admin/user-invite-generic-error.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func handleUsers(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			users := []*orgv1.User{
				{
					Id:             1,
					Email:          "bstrauch@confluent.io",
					FirstName:      "Brian",
					LastName:       "Strauch",
					OrganizationId: 0,
					Deactivated:    false,
					Verified:       nil,
					ResourceId:     "u11",
				},
				{
					Id:             2,
					Email:          "mtodzo@confluent.io",
					FirstName:      "Miles",
					LastName:       "Todzo",
					OrganizationId: 0,
					Deactivated:    false,
					Verified:       nil,
					ResourceId:     "u-17",
				},
			}
			userId := r.URL.Query().Get("id")
			if userId != "" {
				intId, err := strconv.Atoi(userId)
				require.NoError(t, err)
				if int32(intId) == deactivatedUserID {
					users = []*orgv1.User{}
				}
			}

			res := orgv1.GetUsersReply{
				Users:                users,
				Error:                nil,
				XXX_NoUnkeyedLiteral: struct{}{},
				XXX_unrecognized:     nil,
				XXX_sizecache:        0,
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}

	}
}

// used for DELETE
func handleUser(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := orgv1.DeleteUserReply{
			Error: nil,
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

func handleInvite(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bs := string(body)
		if strings.Contains(bs, "test@error.io") {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			res := flowv1.SendInviteReply{
				Error: nil,
				User: &orgv1.User{
					Id:             1,
					Email:          "miles@confluent.io",
					FirstName:      "Miles",
					LastName:       "Todzo",
					OrganizationId: 0,
					Deactivated:    false,
					Verified:       nil,
				},
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}
