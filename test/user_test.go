package test

import (
	"encoding/json"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func (s *CLITestSuite) TestUserList() {
	tests := []CLITest {
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
	tests := []CLITest {
		{
			args:    "admin user describe u-0",
			fixture: "admin/user-describe.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestUserDelete() {
	tests := []CLITest {
		{
			args:    "admin user delete u-0",
			fixture: "admin/user-delete.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestUserInvite() {
	tests := []CLITest {
		{
			args:    "admin user invite miles@confluent.io",
			fixture: "admin/user-invite.golden",
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
			users := []*orgv1.User {
				&orgv1.User{
					Id:                   1,
					Email:                "bstrauch@confluent.io",
					FirstName:            "Brian",
					LastName:             "Strauch",
					OrganizationId:       0,
					Deactivated:          false,
					Verified:             nil,
					ResourceId:           "u11",
				},
			}
			// if no query param is present then it's a list call
			if len(r.URL.Query()["user"]) == 0 {
				users = append(users, &orgv1.User{
					Id:                   2,
					Email:                "mtodzo@confluent.io",
					FirstName:            "Miles",
					LastName:             "Todzo",
					OrganizationId:       0,
					Deactivated:          false,
					Verified:             nil,
					ResourceId:           "u17",
				})
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
		res := orgv1.DeleteUserReply {
			Error:                nil,
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

func handleInvite(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := flowv1.SendInviteReply{
			Error:                nil,
			User:                 &orgv1.User{
				Id:                   1,
				Email:                "miles@confluent.io",
				FirstName:            "Miles",
				LastName:             "Todzo",
				OrganizationId:       0,
				Deactivated:          false,
				Verified:             nil,
			},
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}
