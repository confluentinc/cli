package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"

	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	mds2mock "github.com/confluentinc/mds-sdk-go/mdsv2alpha1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	mock2 "github.com/confluentinc/cli/mock"
)

var (
	errNotFound = fmt.Errorf("user not found")
)

const (
	env123 = "env-123"
)

type roleBindingTest struct {
	args      []string
	principal string
	roleName  string
	scope     mdsv2alpha1.Scope
	err       error
}

type myRoleBindingTest struct {
	scopeRoleBindingMapping []mdsv2alpha1.ScopeRoleBindingMapping
	mockedListUserResult    []*v1.User
	expected                []listDisplay
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
	v3.AddEnvironmentToConfigMock(suite.conf, env123, env123)
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
	mdsClient.RBACRoleBindingCRUDApi = &mds2mock.RBACRoleBindingCRUDApi{
		AddRoleForPrincipalFunc: func(ctx context.Context, principal, roleName string, scope mdsv2alpha1.Scope) (*http.Response, error) {
			assert.Equal(suite.T(), expectedListCmdArgs{principal, roleName, scope}, <-expect, message)
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
		DeleteRoleForPrincipalFunc: func(ctx context.Context, principal, roleName string, scope mdsv2alpha1.Scope) (*http.Response, error) {
			assert.Equal(suite.T(), expectedListCmdArgs{principal, roleName, scope}, <-expect, message)
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}
	userMock := &ccsdkmock.User{
		DescribeFunc: func(arg0 context.Context, arg1 *v1.User) (user *v1.User, e error) {
			if arg1.Email == "test@email.com" {
				return &v1.User{
					Email:      "test@email.com",
					ResourceId: v3.MockUserResourceId,
				}, nil
			} else if arg1.Email == "notfound@email.com" || arg1.ResourceId == "u-noemail" {
				return nil, errNotFound
			} else {
				return &v1.User{
					Email:      arg1.ResourceId + "@email.com",
					ResourceId: arg1.ResourceId,
				}, nil
			}
		},
		ListFunc: func(arg0 context.Context) ([]*v1.User, error) {
			return []*v1.User{{
				Email:      "test@email.com",
				ResourceId: v3.MockUserResourceId,
			}}, nil
		},
	}
	client := &ccloud.Client{
		User: userMock,
	}
	return New("ccloud", mock2.NewPreRunnerMdsV2Mock(client, mdsClient, suite.conf))
}

func TestRoleBindingTestSuite(t *testing.T) {
	suite.Run(t, new(RoleBindingTestSuite))
}

var roleBindingListTests = []roleBindingTest{
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
		args:      []string{"--principal", "User:test@email.com"},
		principal: "User:" + v3.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args: []string{"--principal", "User:notfound@email.com"},
		err:  errNotFound,
	},
	{
		args:     []string{"--role", "OrganizationAdmin"},
		roleName: "OrganizationAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:     []string{"--role", "EnvironmentAdmin", "--current-env"},
		roleName: "EnvironmentAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=" + v3.MockEnvironmentId}},
	},
	{
		args:     []string{"--role", "EnvironmentAdmin", "--environment", "env-123"},
		roleName: "EnvironmentAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=env-123"}},
	},
}

func (suite *RoleBindingTestSuite) TestRoleBindingsList() {
	expect := make(chan interface{})
	for _, tc := range roleBindingListTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rolebinding", "list"}, tc.args...))

		if tc.err == nil {
			go func() {
				expect <- expectedListCmdArgs{
					tc.principal,
					tc.roleName,
					tc.scope,
				}

			}()
			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		} else {
			// error case
			err := cmd.Execute()
			assert.Equal(suite.T(), tc.err, err)
		}
	}
}

func (suite *RoleBindingTestSuite) newMockIamListRoleBindingCmd(
	mockeRoleBindingsResult chan []mdsv2alpha1.ScopeRoleBindingMapping,
	mockeddListUserResult chan []*v1.User,
) *cobra.Command {
	// Mock MDS Client
	mdsClient := mdsv2alpha1.NewAPIClient(mdsv2alpha1.NewConfiguration())
	mdsClient.RBACRoleBindingSummariesApi = &mds2mock.RBACRoleBindingSummariesApi{
		MyRoleBindingsFunc: func(ctx context.Context, principal string, scope mdsv2alpha1.Scope) ([]mdsv2alpha1.ScopeRoleBindingMapping, *http.Response, error) {
			return <-mockeRoleBindingsResult, nil, nil
		},
	}

	// Mock User Client
	userMock := &ccsdkmock.User{
		ListFunc: func(arg0 context.Context) ([]*v1.User, error) {
			return <-mockeddListUserResult, nil
		},
	}
	client := &ccloud.Client{
		User: userMock,
	}
	return New("ccloud", mock2.NewPreRunnerMdsV2Mock(client, mdsClient, suite.conf))
}

var myRoleBindingListTests = []myRoleBindingTest{
	// Principal whose email address is NOT known will be returned without an email address
	{
		scopeRoleBindingMapping: []mdsv2alpha1.ScopeRoleBindingMapping{
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:u-epo7ml": {
						"MetricsViewer": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
		},
		mockedListUserResult: []*v1.User{{
			Email:      "test@email.com",
			ResourceId: v3.MockUserResourceId,
		}},
		expected: []listDisplay{
			{
				Principal: "User:u-epo7ml",
				Role:      "MetricsViewer",
				Name:      "Skynet",
			},
		},
	},
	// Principal whose email address is known will be returned with eamil address
	{
		scopeRoleBindingMapping: []mdsv2alpha1.ScopeRoleBindingMapping{
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v3.MockUserResourceId: {
						"MetricsViewer": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
		},
		mockedListUserResult: []*v1.User{{
			Email:      "test@email.com",
			ResourceId: v3.MockUserResourceId,
		}},
		expected: []listDisplay{
			{
				Principal: "User:u-123",
				Role:      "MetricsViewer",
				Name:      "Skynet",
				Email:     "test@email.com",
			},
		},
	},
	// Service Account
	{
		scopeRoleBindingMapping: []mdsv2alpha1.ScopeRoleBindingMapping{
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:sa-123": {
						"MetricsViewer": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
		},
		mockedListUserResult: []*v1.User{},
		expected: []listDisplay{
			{
				Principal: "User:sa-123",
				Role:      "MetricsViewer",
				Name:      "Skynet",
			},
		},
	},
}

func (suite *RoleBindingTestSuite) TestMyRoleBindingsList() {
	mockeRoleBindingsResult := make(chan []mdsv2alpha1.ScopeRoleBindingMapping)
	mockeListUserResult := make(chan []*v1.User)
	for _, tc := range myRoleBindingListTests {
		cmd := suite.newMockIamListRoleBindingCmd(mockeRoleBindingsResult, mockeListUserResult)

		go func() {
			mockeRoleBindingsResult <- tc.scopeRoleBindingMapping
			mockeListUserResult <- tc.mockedListUserResult
		}()
		output, err := pcmd.ExecuteCommand(cmd, "rolebinding", "list", "--current-user", "-ojson")
		assert.Nil(suite.T(), err)
		var actual []listDisplay
		err = json.Unmarshal([]byte(output), &actual)
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), tc.expected, actual)
	}
}

var roleBindingCreateDeleteTests = []roleBindingTest{
	{
		args:      []string{"--principal", "User:" + v3.MockUserResourceId, "--role", "OrganizationAdmin"},
		principal: "User:" + v3.MockUserResourceId,
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:u-xyz", "--role", "OrganizationAdmin"},
		principal: "User:u-xyz",
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:test@email.com", "--role", "OrganizationAdmin"},
		principal: "User:" + v3.MockUserResourceId,
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId}},
	},
	{
		args: []string{"--principal", "User:notfound@email.com", "--role", "OrganizationAdmin"},
		err:  errNotFound,
	},
	{
		args:      []string{"--principal", "User:" + v3.MockUserResourceId, "--role", "EnvironmentAdmin", "--current-env"},
		principal: "User:" + v3.MockUserResourceId,
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=" + v3.MockEnvironmentId}},
	},
	{
		args:      []string{"--principal", "User:" + v3.MockUserResourceId, "--role", "EnvironmentAdmin", "--environment", env123},
		principal: "User:" + v3.MockUserResourceId,
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=" + env123}},
	},
	{
		args:      []string{"--principal", "User:u-noemail", "--role", "EnvironmentAdmin", "--environment", v3.MockEnvironmentId},
		principal: "User:u-noemail",
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v3.MockOrgResourceId, "environment=" + v3.MockEnvironmentId}},
	},
}

func (suite *RoleBindingTestSuite) TestRoleBindingsCreate() {
	expect := make(chan interface{})
	for _, tc := range roleBindingCreateDeleteTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rolebinding", "create"}, tc.args...))

		if tc.err == nil {
			go func() {
				expect <- expectedListCmdArgs{
					tc.principal,
					tc.roleName,
					tc.scope,
				}

			}()
			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		} else {
			// error case
			err := cmd.Execute()
			assert.Equal(suite.T(), tc.err, err)
		}
	}
}

func (suite *RoleBindingTestSuite) TestRoleBindingsDelete() {
	expect := make(chan interface{})
	for _, tc := range roleBindingCreateDeleteTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rolebinding", "delete"}, tc.args...))

		if tc.err == nil {
			go func() {
				expect <- expectedListCmdArgs{
					tc.principal,
					tc.roleName,
					tc.scope,
				}

			}()
			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		} else {
			// error case
			err := cmd.Execute()
			assert.Equal(suite.T(), tc.err, err)
		}
	}
}
