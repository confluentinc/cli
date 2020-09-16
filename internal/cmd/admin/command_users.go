package admin

import (
	"context"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
	"regexp"
	"strconv"
)

var (
	listFields          = []string{"Id", "Email", "FirstName", "LastName", "Deactivated", "ResourceId"}
	humanLabels         = []string{"ID", "Email", "First Name", "Last Name", "Deactivated", "Resource ID"}
	humanLabelMap		= map[string]string{
		"Id": "ID",
		"Email":"Email",
		"FirstName":"First Name",
		"LastName":"Last Name",
		"Deactivated":"Deactivated",
		"ResourceId": "Resource ID",
	}
	structuredLabels    = []string{"id", "email", "first_name", "last_name", "deactivated", "resource_id"}
	structuredLabelMap		= map[string]string{
		"Id": "id",
		"Email":"email",
		"FirstName":"first_name",
		"LastName":"last_name",
		"Deactivated":"deactivated",
		"ResourceId": "resource_id",
	}
)

type userCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func NewUsersCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &userCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "users",
				Short: "Manage an organization's users.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}
	c.AddCommand(c.newUserDescribeCommand())
	c.AddCommand(c.newUserListCommand())
	c.AddCommand(c.newUserInviteCommand())
	c.AddCommand(c.newUserDeleteCommand())
	return c.Command
}

func (c userCommand) newUserDescribeCommand() (describeCmd *cobra.Command) {
	describeCmd = &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	return
}

func (c userCommand) describe(cmd *cobra.Command, args []string) error {
	userId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return err
	}
	user, err := c.Client.User.Describe(context.Background(), &orgv1.User{
		Id:                   int32(userId),
		OrganizationId:       c.State.Auth.User.OrganizationId,
	})
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, user, listFields, humanLabelMap, structuredLabelMap)
}

func (c userCommand) newUserListCommand() (listCmd *cobra.Command) {
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List an organization's users.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	return
}

func (c userCommand) list(cmd *cobra.Command, _ []string) error {
	type userStruct struct {
		Id    		int32
		Email  		string
		FirstName   string
		LastName 	string
		Deactivated bool
		ResourceId 	string
	}
	users, err := c.Client.User.List(context.Background())
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}
	for _, user := range users {
		outputWriter.AddElement(&userStruct{
			Id: user.Id,
			Email: user.Email,
			FirstName: user.FirstName,
			LastName: user.LastName,
			Deactivated: user.Deactivated,
			ResourceId: user.ResourceId,
		})
	}
	return outputWriter.Out()
}

func (c userCommand) newUserInviteCommand() (createCmd *cobra.Command) {
	createCmd = &cobra.Command {
		Use:   "invite <email>",
		Short: "Invite a user to join your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.invite),
	}
	return
}

func (c userCommand) invite(cmd *cobra.Command, args []string) error {
	email := args[0]
	matched := validateEmail(email)
	if !matched {
		return errors.New("invalid email structure")
	}
	newUser := &orgv1.User{Email: email, OrganizationId: c.State.Auth.Organization.Id}
	user, err := c.Client.User.Invite(context.Background(), newUser)
	if err != nil {
		return err
	}
	pcmd.Println(cmd, "An email invitation has been sent to " + user.Email)
	return nil
}

func validateEmail(email string) bool {
	rgxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	matched := rgxEmail.MatchString(email)
	return matched
}

func (c userCommand) newUserDeleteCommand() (deleteCmd *cobra.Command) {
	deleteCmd = &cobra.Command {
		Use:   "delete <id>",
		Short: "Delete a user from your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	return
}

func (c userCommand) delete(cmd *cobra.Command, args []string) error {
	userId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return err
	}
	c.Client.User.Delete(context.Background(), &orgv1.User{
		Id: 			int32(userId),
		OrganizationId: c.State.Auth.Organization.Id,
	})
	pcmd.Println(cmd, "Successfully deleted user.")
	return nil
}
