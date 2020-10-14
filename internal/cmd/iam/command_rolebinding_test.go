package iam

import (
	"context"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	mock2 "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	mds2mock "github.com/confluentinc/mds-sdk-go/mdsv2alpha1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

var roleBindingListPatterns = []struct {
	args      []string
	principal string
	roleName  string
	scope     mdsv2alpha1.Scope
}{
	{
		args:      []string{"--current-user"},
		principal: "User:" + v3.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:" + v3.MockUserResourceId},
		principal: "User:" + v3.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:u-xyz"},
		principal: "User:u-xyz",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:     []string{"--role", "OrganizationAdmin"},
		roleName: "OrganizationAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:     []string{"--role", "EnvironmentAdmin", "--environment", "env-123"},
		roleName: "EnvironmentAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=env-123"}},
	},
}

type expectedListCmdArgs struct {
	principal string
	roleName  string
	scope     mdsv2alpha1.Scope
}

type RoleBindingTestSuite struct {
	suite.Suite
	conf *v3.Config
}

func (suite *RoleBindingTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
}

func (suite *RoleBindingTestSuite) newMockIamRoleBindingCmd(expect chan interface{}, message string) *cobra.Command {

	mdsClient := mdsv2alpha1.NewAPIClient(mdsv2alpha1.NewConfiguration())
	mdsClient.RBACRoleBindingSummariesApi = &mds2mock.RBACRoleBindingSummariesApi{
		MyRoleBindingsFunc: func(ctx context.Context, principal string, scope mdsv2alpha1.Scope) ([]mdsv2alpha1.ScopeRoleBindingMapping, *http.Response, error) {
			assert.Equal(suite.T(), expectedListCmdArgs{principal, "", scope}, <-expect, message)
			return nil, nil, nil
		},
		LookupPrincipalsWithRoleFunc: func(ctx context.Context, roleName string, scope mdsv2alpha1.Scope) ([]string, *http.Response, error) {
			assert.Equal(suite.T(), expectedListCmdArgs{"", roleName, scope}, <-expect, message)
			return nil, nil, nil
		},
	}
	return New("ccloud", mock2.NewPreRunnerMdsV2Mock(nil, mdsClient, suite.conf))
}

func TestRoleBindingTestSuite(t *testing.T) {
	suite.Run(t, new(RoleBindingTestSuite))
}

func (suite *RoleBindingTestSuite) TestRoleBindingsList() {
	expect := make(chan interface{})
	for _, roleBindingListPattern := range roleBindingListPatterns {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rolebinding", "list"}, roleBindingListPattern.args...))

		go func() {
			expect <- expectedListCmdArgs{
				roleBindingListPattern.principal,
				roleBindingListPattern.roleName,
				roleBindingListPattern.scope,
			}

		}()

		err := cmd.Execute()
		assert.Nil(suite.T(), err)
	}
}
