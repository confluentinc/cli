package iam

import (
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var statusMap = map[flowv1.UserStatus]string{
	flowv1.UserStatus_USER_STATUS_UNKNOWN:     "Unknown",
	flowv1.UserStatus_USER_STATUS_UNVERIFIED:  "Unverified",
	flowv1.UserStatus_USER_STATUS_ACTIVE:      "Active",
	flowv1.UserStatus_USER_STATUS_DEACTIVATED: "Deactivated",
}

var authMethodFormats = map[orgv1.AuthMethod]string{
	orgv1.AuthMethod_AUTH_METHOD_UNKNOWN:      "Unknown",
	orgv1.AuthMethod_AUTH_METHOD_USERNAME_PWD: "Username/Password",
	orgv1.AuthMethod_AUTH_METHOD_SSO:          "SSO",
}

type userCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type userOut struct {
	Id                   string `human:"ID" serialized:"id"`
	Email                string `human:"Email" serialized:"email"`
	FirstName            string `human:"First Name" serialized:"first_name"`
	LastName             string `human:"Last Name" serialized:"last_name"`
	Status               string `human:"Status" serialized:"status"`
	AuthenticationMethod string `human:"Authentication Method" serialized:"authentication_method"`
}

func newUserCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "user",
		Short:       "Manage users.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &userCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(newInvitationCommand(prerunner))
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}
