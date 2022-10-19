package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	ccv2sdkmock "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2/mock"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	mds2mock "github.com/confluentinc/mds-sdk-go/mdsv2alpha1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
)

var (
	errUserNotFound = errors.Errorf(errors.InvalidEmailErrorMsg, "notfound@email.com")
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
	mockedListIamUserResult iamv2.IamV2UserList
	expected                []listDisplay
}

type expectedListCmdArgs struct {
	principal string
	roleName  string
	scope     mdsv2alpha1.Scope
}

type RoleBindingTestSuite struct {
	suite.Suite
	conf *v1.Config
}

func (suite *RoleBindingTestSuite) SetupSuite() {
	os.Setenv("XX_DATAPLANE_3_ENABLE", "1")
	suite.conf = v1.AuthenticatedCloudConfigMock()
	v1.AddEnvironmentToConfigMock(suite.conf, env123, env123)
}

func (suite *RoleBindingTestSuite) newMockIamRoleBindingCmd(expect chan expectedListCmdArgs, message string) *cobra.Command {
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
	iamUserMock := &iammock.UsersIamV2Api{
		GetIamV2UserFunc: func(ctx context.Context, id string) iamv2.ApiGetIamV2UserRequest {
			return iamv2.ApiGetIamV2UserRequest{}
		},
		GetIamV2UserExecuteFunc: func(r iamv2.ApiGetIamV2UserRequest) (iamv2.IamV2User, *http.Response, error) {
			return iamv2.IamV2User{
				Email: iamv2.PtrString("test@email.com"),
				Id:    iamv2.PtrString(v1.MockUserResourceId),
			}, nil, nil
		},
		ListIamV2UsersFunc: func(ctx context.Context) iamv2.ApiListIamV2UsersRequest {
			return iamv2.ApiListIamV2UsersRequest{}
		},
		ListIamV2UsersExecuteFunc: func(r iamv2.ApiListIamV2UsersRequest) (iamv2.IamV2UserList, *http.Response, error) {
			return iamv2.IamV2UserList{Data: []iamv2.IamV2User{
				{
					Email: iamv2.PtrString("test@email.com"),
					Id:    iamv2.PtrString(v1.MockUserResourceId),
				},
			}}, nil, nil
		},
	}
	iamServiceAccountMock := &iammock.ServiceAccountsIamV2Api{
		GetIamV2ServiceAccountFunc: func(ctx context.Context, id string) iamv2.ApiGetIamV2ServiceAccountRequest {
			return iamv2.ApiGetIamV2ServiceAccountRequest{}
		},
		GetIamV2ServiceAccountExecuteFunc: func(r iamv2.ApiGetIamV2ServiceAccountRequest) (iamv2.IamV2ServiceAccount, *http.Response, error) {
			return iamv2.IamV2ServiceAccount{
				DisplayName: iamv2.PtrString("One Great Service"),
				Id:          iamv2.PtrString("User:sa-123456"),
			}, nil, nil
		},
		ListIamV2ServiceAccountsFunc: func(ctx context.Context) iamv2.ApiListIamV2ServiceAccountsRequest {
			return iamv2.ApiListIamV2ServiceAccountsRequest{}
		},
		ListIamV2ServiceAccountsExecuteFunc: func(r iamv2.ApiListIamV2ServiceAccountsRequest) (iamv2.IamV2ServiceAccountList, *http.Response, error) {
			return iamv2.IamV2ServiceAccountList{Data: []iamv2.IamV2ServiceAccount{
				{
					DisplayName: iamv2.PtrString("One Great Service"),
					Id:          iamv2.PtrString("User:sa-123456"),
				},
			}}, nil, nil
		},
	}
	providerMock := &ccv2sdkmock.IdentityProvidersIamV2Api{
		ListIamV2IdentityProvidersFunc: func(_ context.Context) identityproviderv2.ApiListIamV2IdentityProvidersRequest {
			return identityproviderv2.ApiListIamV2IdentityProvidersRequest{}
		},
		ListIamV2IdentityProvidersExecuteFunc: func(_ identityproviderv2.ApiListIamV2IdentityProvidersRequest) (identityproviderv2.IamV2IdentityProviderList, *http.Response, error) {
			id := "op-01"
			prov := identityproviderv2.IamV2IdentityProvider{Id: &id, DisplayName: &id}
			return identityproviderv2.IamV2IdentityProviderList{Data: []identityproviderv2.IamV2IdentityProvider{prov}}, nil, nil
		},
	}
	poolMock := &ccv2sdkmock.IdentityPoolsIamV2Api{
		ListIamV2IdentityPoolsFunc: func(_ context.Context, _ string) identityproviderv2.ApiListIamV2IdentityPoolsRequest {
			return identityproviderv2.ApiListIamV2IdentityPoolsRequest{}
		},
		ListIamV2IdentityPoolsExecuteFunc: func(_ identityproviderv2.ApiListIamV2IdentityPoolsRequest) (identityproviderv2.IamV2IdentityPoolList, *http.Response, error) {
			id := "pool-01"
			pool := identityproviderv2.IamV2IdentityPool{Id: &id, DisplayName: &id}
			return identityproviderv2.IamV2IdentityPoolList{Data: []identityproviderv2.IamV2IdentityPool{pool}}, nil, nil
		},
	}
	v2client := &ccloudv2.Client{
		AuthToken: "auth-token",
		IamClient: &iamv2.APIClient{
			UsersIamV2Api:           iamUserMock,
			ServiceAccountsIamV2Api: iamServiceAccountMock,
		},
		IdentityProviderClient: &identityproviderv2.APIClient{
			IdentityPoolsIamV2Api:     poolMock,
			IdentityProvidersIamV2Api: providerMock,
		},
	}
	return New(suite.conf, climock.NewPreRunnerMdsV2Mock(nil, v2client, mdsClient, suite.conf))
}

func TestRoleBindingTestSuite(t *testing.T) {
	suite.Run(t, new(RoleBindingTestSuite))
}

var roleBindingListTests = []roleBindingTest{
	{
		args:      []string{"--current-user"},
		principal: "User:" + v1.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId},
		principal: "User:" + v1.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:u-xyz"},
		principal: "User:u-xyz",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:test@email.com"},
		principal: "User:" + v1.MockUserResourceId,
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args: []string{"--principal", "User:notfound@email.com"},
		err:  errUserNotFound,
	},
	{
		args:     []string{"--role", "OrganizationAdmin"},
		roleName: "OrganizationAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:     []string{"--role", "EnvironmentAdmin", "--current-env"},
		roleName: "EnvironmentAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId, "environment=" + v1.MockEnvironmentId}},
	},
	{
		args:     []string{"--role", "EnvironmentAdmin", "--environment", "env-123"},
		roleName: "EnvironmentAdmin",
		scope:    mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId, "environment=env-123"}},
	},
	{
		args:      []string{"--current-user", "--environment", "env-123", "--kafka-cluster-id", "lkc-123"},
		principal: "User:" + v1.MockUserResourceId,
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{KafkaCluster: "lkc-123"},
		},
	},
	{
		args:      []string{"--current-user", "--environment", "env-123", "--cloud-cluster", "lkc-123", "--ksql-cluster-id", "ksql-9999"},
		principal: "User:" + v1.MockUserResourceId,
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{KsqlCluster: "ksql-9999"},
		},
	},
	{
		args:      []string{"--current-user", "--environment", "env-123", "--cloud-cluster", "lkc-123", "--schema-registry-cluster-id", "sr-777"},
		principal: "User:" + v1.MockUserResourceId,
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{SchemaRegistryCluster: "sr-777"},
		},
	},
}

func (suite *RoleBindingTestSuite) TestRoleBindingsList() {
	expect := make(chan expectedListCmdArgs)
	for _, tc := range roleBindingListTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, fmt.Sprintf("%v", tc.args))
		cmd.SetArgs(append([]string{"rbac", "role-binding", "list"}, tc.args...))

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
			assert.Equal(suite.T(), tc.err.Error(), err.Error())
		}
	}
}

func (suite *RoleBindingTestSuite) newMockIamListRoleBindingCmd(mockRoleBindingsResult chan []mdsv2alpha1.ScopeRoleBindingMapping, mockListIamUserResult chan iamv2.IamV2UserList) *cobra.Command {
	// Mock MDS Client
	mdsClient := mdsv2alpha1.NewAPIClient(mdsv2alpha1.NewConfiguration())
	mdsClient.RBACRoleBindingSummariesApi = &mds2mock.RBACRoleBindingSummariesApi{
		MyRoleBindingsFunc: func(ctx context.Context, principal string, scope mdsv2alpha1.Scope) ([]mdsv2alpha1.ScopeRoleBindingMapping, *http.Response, error) {
			return <-mockRoleBindingsResult, nil, nil
		},
	}

	// Mock User Client
	iamUserMock := &iammock.UsersIamV2Api{
		ListIamV2UsersFunc: func(ctx context.Context) iamv2.ApiListIamV2UsersRequest {
			return iamv2.ApiListIamV2UsersRequest{}
		},
		ListIamV2UsersExecuteFunc: func(r iamv2.ApiListIamV2UsersRequest) (iamv2.IamV2UserList, *http.Response, error) {
			return <-mockListIamUserResult, nil, nil
		},
	}
	v2Client := &ccloudv2.Client{
		IamClient: &iamv2.APIClient{UsersIamV2Api: iamUserMock},
	}
	return New(suite.conf, climock.NewPreRunnerMdsV2Mock(nil, v2Client, mdsClient, suite.conf))
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
		mockedListIamUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{{
				Email: iamv2.PtrString("test@email.com"),
				Id:    iamv2.PtrString(v1.MockUserResourceId),
			}},
		},
		expected: []listDisplay{
			{
				Principal: "User:u-epo7ml",
				Role:      "MetricsViewer",
			},
		},
	},
	// Principal whose email address is known will be returned with email address
	{
		scopeRoleBindingMapping: []mdsv2alpha1.ScopeRoleBindingMapping{
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"MetricsViewer": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
		},
		mockedListIamUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{{
				Email: iamv2.PtrString("test@email.com"),
				Id:    iamv2.PtrString(v1.MockUserResourceId),
			}},
		},
		expected: []listDisplay{
			{
				Principal: "User:u-123",
				Role:      "MetricsViewer",
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
		mockedListIamUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{},
		},
		expected: []listDisplay{
			{
				Principal: "User:sa-123",
				Role:      "MetricsViewer",
			},
		},
	},
	// Multiple role bindings at various scopes
	{
		scopeRoleBindingMapping: []mdsv2alpha1.ScopeRoleBindingMapping{
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"OrganizationAdmin": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet", "environment=Cyberdyne"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"EnvironmentAdmin": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet", "environment=Cyberdyne", "cloud-cluster=t1000"},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"CloudClusterAdmin": []mdsv2alpha1.ResourcePattern{},
					},
				},
			},
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet", "environment=Cyberdyne", "cloud-cluster=t1000"},
					Clusters: mdsv2alpha1.ScopeClusters{
						KafkaCluster: "t1000",
					},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"ResourceOwner": []mdsv2alpha1.ResourcePattern{{
							ResourceType: "Topic",
							Name:         "connor",
							PatternType:  "LITERAL",
						}, {
							ResourceType: "Topic",
							Name:         "john",
							PatternType:  "PREFIX",
						}},
					},
				},
			},
			{
				Scope: mdsv2alpha1.Scope{
					Path: []string{"organization=Skynet", "environment=Cyberdyne"},
					Clusters: mdsv2alpha1.ScopeClusters{
						SchemaRegistryCluster: "sr1000",
					},
				},
				Rolebindings: map[string]map[string][]mdsv2alpha1.ResourcePattern{
					"User:" + v1.MockUserResourceId: {
						"DeveloperRead": []mdsv2alpha1.ResourcePattern{{
							ResourceType: "Subject",
							Name:         "terminators",
							PatternType:  "LITERAL",
						}},
					},
				},
			},
		},
		mockedListIamUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{{
				Email: iamv2.PtrString("test@email.com"),
				Id:    iamv2.PtrString(v1.MockUserResourceId),
			}},
		},
		expected: []listDisplay{
			{
				Principal:    "User:u-123",
				Role:         "CloudClusterAdmin",
				Email:        "test@email.com",
				Environment:  "Cyberdyne",
				CloudCluster: "t1000",
			},
			{
				Principal:      "User:u-123",
				Role:           "DeveloperRead",
				Email:          "test@email.com",
				Environment:    "Cyberdyne",
				ClusterType:    "Schema Registry",
				LogicalCluster: "sr1000",
				ResourceType:   "Subject",
				Name:           "terminators",
				PatternType:    "LITERAL",
			},
			{
				Principal:   "User:u-123",
				Role:        "EnvironmentAdmin",
				Email:       "test@email.com",
				Environment: "Cyberdyne",
			},
			{
				Principal: "User:u-123",
				Role:      "OrganizationAdmin",
				Email:     "test@email.com",
			},
			{
				Principal:      "User:u-123",
				Role:           "ResourceOwner",
				Email:          "test@email.com",
				Environment:    "Cyberdyne",
				CloudCluster:   "t1000",
				ClusterType:    "Kafka",
				LogicalCluster: "t1000",
				ResourceType:   "Topic",
				Name:           "connor",
				PatternType:    "LITERAL",
			},
			{
				Principal:      "User:u-123",
				Role:           "ResourceOwner",
				Email:          "test@email.com",
				Environment:    "Cyberdyne",
				CloudCluster:   "t1000",
				ClusterType:    "Kafka",
				LogicalCluster: "t1000",
				ResourceType:   "Topic",
				Name:           "john",
				PatternType:    "PREFIX",
			},
		},
	},
}

func (suite *RoleBindingTestSuite) TestMyRoleBindingsList() {
	mockRoleBindingsResult := make(chan []mdsv2alpha1.ScopeRoleBindingMapping)
	mockListIamUserResult := make(chan iamv2.IamV2UserList)
	for _, tc := range myRoleBindingListTests {
		cmd := suite.newMockIamListRoleBindingCmd(mockRoleBindingsResult, mockListIamUserResult)

		go func() {
			mockRoleBindingsResult <- tc.scopeRoleBindingMapping
			mockListIamUserResult <- tc.mockedListIamUserResult
		}()
		output, err := pcmd.ExecuteCommand(cmd, "rbac", "role-binding", "list", "--current-user", "-ojson")
		assert.Nil(suite.T(), err)
		var actual []listDisplay
		err = json.Unmarshal([]byte(output), &actual)
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), tc.expected, actual)
	}
}

var roleBindingCreateDeleteTests = []roleBindingTest{
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "OrganizationAdmin"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:u-xyz", "--role", "OrganizationAdmin"},
		principal: "User:u-xyz",
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args:      []string{"--principal", "User:test@email.com", "--role", "OrganizationAdmin"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "OrganizationAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId}},
	},
	{
		args: []string{"--principal", "User:notfound@email.com", "--role", "OrganizationAdmin"},
		err:  errUserNotFound,
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "EnvironmentAdmin", "--current-env"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId, "environment=" + v1.MockEnvironmentId}},
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "EnvironmentAdmin", "--environment", env123},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId, "environment=" + env123}},
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "ResourceOwner", "--environment", env123, "--kafka-cluster-id", "lkc-123"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "ResourceOwner",
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{KafkaCluster: "lkc-123"},
		},
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "ResourceOwner", "--environment", env123, "--cloud-cluster", "lkc-123", "--ksql-cluster-id", "ksql-9999"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "ResourceOwner",
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{KsqlCluster: "ksql-9999"},
		},
	},
	{
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "ResourceOwner", "--environment", env123, "--cloud-cluster", "lkc-123", "--schema-registry-cluster-id", "sr-777"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "ResourceOwner",
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{SchemaRegistryCluster: "sr-777"},
		},
	},
	{
		args:      []string{"--principal", "User:u-noemail", "--role", "EnvironmentAdmin", "--environment", v1.MockEnvironmentId},
		principal: "User:u-noemail",
		roleName:  "EnvironmentAdmin",
		scope:     mdsv2alpha1.Scope{Path: []string{"organization=" + v1.MockOrgResourceId, "environment=" + v1.MockEnvironmentId}},
	},
}

func (suite *RoleBindingTestSuite) TestRoleBindingsCreate() {
	expect := make(chan expectedListCmdArgs)
	for _, tc := range roleBindingCreateDeleteTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rbac", "role-binding", "create"}, tc.args...))

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
			assert.Equal(suite.T(), tc.err.Error(), err.Error())
		}
	}
}

func (suite *RoleBindingTestSuite) TestRoleBindingsDelete() {
	expect := make(chan expectedListCmdArgs)
	for _, tc := range roleBindingCreateDeleteTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rbac", "role-binding", "delete"}, tc.args...))

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
			assert.Equal(suite.T(), tc.err.Error(), err.Error())
		}
	}
}

func TestParseAndValidateResourcePattern_Prefixed(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:test", true)
	require.NoError(t, err)
	require.Equal(t, "PREFIXED", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Literal(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", false)
	require.NoError(t, err)
	require.Equal(t, "LITERAL", pattern.PatternType)
}

func TestParseAndValidateResourcePattern_Topic(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a", true)
	require.NoError(t, err)
	require.Equal(t, "Topic", pattern.ResourceType)
	require.Equal(t, "a", pattern.Name)
}

func TestParseAndValidateResourcePattern_TopicWithColon(t *testing.T) {
	pattern, err := parseAndValidateResourcePattern("Topic:a:b", true)
	require.NoError(t, err)
	require.Equal(t, "a:b", pattern.Name)
}

func TestParseAndValidateResourcePattern_ErrIncorrectResourceFormat(t *testing.T) {
	_, err := parseAndValidateResourcePattern("string with no colon", true)
	require.Error(t, err)
}
