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
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	mdsmock "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2/mock"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	mds2mock "github.com/confluentinc/mds-sdk-go/mdsv2alpha1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
)

var (
	errNotFound = fmt.Errorf("user \"notfound@email.com\" not found")
)

const (
	env123 = "env-123"
)

var v2RoleBindingMock = &mdsmock.RoleBindingsIamV2Api{
	CreateIamV2RoleBindingFunc: func(_ context.Context) mdsv2.ApiCreateIamV2RoleBindingRequest {
		return mdsv2.ApiCreateIamV2RoleBindingRequest{}
	},
	CreateIamV2RoleBindingExecuteFunc: func(_ mdsv2.ApiCreateIamV2RoleBindingRequest) (mdsv2.IamV2RoleBinding, *http.Response, error) {
		return mdsv2.IamV2RoleBinding{}, &http.Response{StatusCode: http.StatusOK}, nil
	},
	ListIamV2RoleBindingsFunc: func(_ context.Context) mdsv2.ApiListIamV2RoleBindingsRequest {
		return mdsv2.ApiListIamV2RoleBindingsRequest{}
	},
	ListIamV2RoleBindingsExecuteFunc: func(_ mdsv2.ApiListIamV2RoleBindingsRequest) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
		return mdsv2.IamV2RoleBindingList{
			Data: []mdsv2.IamV2RoleBinding{
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("1"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("ResourceOwner"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=org-resource-id/cloud-cluster=lkc-123/ksql=ksql-9999"),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("2"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("ResourceOwner"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=org-resource-id/cloud-cluster=lkc-123/schema-registry=sr-777"),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("3"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("OrganizationAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=" + v1.MockOrgResourceId),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("4"),
					Principal:  mdsv2.PtrString("User:u-xyz"),
					RoleName:   mdsv2.PtrString("OrganizationAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=" + v1.MockOrgResourceId),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("5"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("OrganizationAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=" + v1.MockOrgResourceId),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("6"),
					Principal:  mdsv2.PtrString("User:notfound@email.com"),
					RoleName:   mdsv2.PtrString("OrganizationAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=" + v1.MockOrgResourceId),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("7"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("EnvironmentAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=org-resource-id/environment=" + v1.MockEnvironmentId),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("8"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("EnvironmentAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=org-resource-id/environment=" + env123),
				},
				mdsv2.IamV2RoleBinding{
					Id:         mdsv2.PtrString("9"),
					Principal:  mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("ResourceOwner"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=org-resource-id/environment=env-123/cloud-cluster=lkc-123/kafka=lkc-123"),
				},
			}}, nil, nil
	},
	DeleteIamV2RoleBindingFunc: func(_ context.Context, _ string) mdsv2.ApiDeleteIamV2RoleBindingRequest {
		return mdsv2.ApiDeleteIamV2RoleBindingRequest{}
	},
	DeleteIamV2RoleBindingExecuteFunc: func(_ mdsv2.ApiDeleteIamV2RoleBindingRequest) (mdsv2.IamV2RoleBinding, *http.Response, error) {
		return mdsv2.IamV2RoleBinding{}, &http.Response{StatusCode: http.StatusOK}, nil
	},
}

var v2UserMock = &iammock.UsersIamV2Api{
	ListIamV2UsersFunc: func(_ context.Context) iamv2.ApiListIamV2UsersRequest {
		return iamv2.ApiListIamV2UsersRequest{}
	},
	ListIamV2UsersExecuteFunc: func(_ iamv2.ApiListIamV2UsersRequest) (iamv2.IamV2UserList, *http.Response, error) {
		user := iamv2.IamV2User{
			Email: iamv2.PtrString("test@email.com"),
			Id:    iamv2.PtrString(v1.MockUserResourceId),
		}
		return iamv2.IamV2UserList{Data: []iamv2.IamV2User{user}}, nil, nil
	},
	GetIamV2UserFunc: func(_ context.Context, _ string) iamv2.ApiGetIamV2UserRequest {
		return iamv2.ApiGetIamV2UserRequest{}
	},
	GetIamV2UserExecuteFunc: func(_ iamv2.ApiGetIamV2UserRequest) (iamv2.IamV2User, *http.Response, error) {
		return iamv2.IamV2User{
			Email: iamv2.PtrString("test@email.com"),
			Id:    iamv2.PtrString(v1.MockUserResourceId),
		}, nil, nil
	},
}

var v2ServiceAccountMock = &iammock.ServiceAccountsIamV2Api{
	ListIamV2ServiceAccountsFunc: func(_ context.Context) iamv2.ApiListIamV2ServiceAccountsRequest {
		return iamv2.ApiListIamV2ServiceAccountsRequest{}
	},
	ListIamV2ServiceAccountsExecuteFunc: func(_ iamv2.ApiListIamV2ServiceAccountsRequest) (iamv2.IamV2ServiceAccountList, *http.Response, error) {
		serviceAccount := iamv2.IamV2ServiceAccount{DisplayName: iamv2.PtrString("One Great Service"), Id: iamv2.PtrString("sa-123456")}
		return iamv2.IamV2ServiceAccountList{Data: []iamv2.IamV2ServiceAccount{serviceAccount}}, nil, nil
	},
}

type roleBindingTest struct {
	args      []string
	principal string
	roleName  string
	scope     mdsv2alpha1.Scope
	err       error
}

type myRoleBindingTest struct {
	mockRoleBindingsResult mdsv2.IamV2RoleBindingList
	mockedListUserResult   iamv2.IamV2UserList
	expected               []listDisplay
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

	v2Client := &ccloudv2.Client{
		IamClient: &iamv2.APIClient{UsersIamV2Api: v2UserMock, ServiceAccountsIamV2Api: v2ServiceAccountMock},
		MdsClient: &mdsv2.APIClient{RoleBindingsIamV2Api: v2RoleBindingMock},
		AuthToken: "auth-token",
	}

	return New(suite.conf, climock.NewPreRunnerMdsV2Mock(nil, v2Client, mdsClient, suite.conf))
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
		err:  errNotFound,
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
				copy := expectedListCmdArgs{
					tc.principal, tc.roleName, tc.scope,
				}
				fmt.Println("")
				expect <- copy
			}()
			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		} else {
			err := cmd.Execute()
			assert.Equal(suite.T(), tc.err, err)
		}
	}
}

func (suite *RoleBindingTestSuite) newMockIamListRoleBindingCmd(mockRoleBindingsResult chan mdsv2.IamV2RoleBindingList, mockListUserResult chan iamv2.IamV2UserList) *cobra.Command {
	v2RoleBindingMock := &mdsmock.RoleBindingsIamV2Api{
		ListIamV2RoleBindingsFunc: func(_ context.Context) mdsv2.ApiListIamV2RoleBindingsRequest {
			return mdsv2.ApiListIamV2RoleBindingsRequest{}
		},
		ListIamV2RoleBindingsExecuteFunc: func(_ mdsv2.ApiListIamV2RoleBindingsRequest) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
			return <-mockRoleBindingsResult, nil, nil
		},
	}

	v2UserMock := &iammock.UsersIamV2Api{
		ListIamV2UsersFunc: func(ctx context.Context) iamv2.ApiListIamV2UsersRequest {
			return iamv2.ApiListIamV2UsersRequest{}
		},
		ListIamV2UsersExecuteFunc: func(r iamv2.ApiListIamV2UsersRequest) (iamv2.IamV2UserList, *http.Response, error) {
			return <-mockListUserResult, nil, nil
		},
	}

	v2Client := &ccloudv2.Client{
		IamClient: &iamv2.APIClient{UsersIamV2Api: v2UserMock},
		MdsClient: &mdsv2.APIClient{RoleBindingsIamV2Api: v2RoleBindingMock},
		AuthToken: "auth-token",
	}
	return New(suite.conf, climock.NewPreRunnerMdsV2Mock(nil, v2Client, nil, suite.conf))
}

var myRoleBindingListTests = []myRoleBindingTest{
	{
		mockRoleBindingsResult: mdsv2.IamV2RoleBindingList{
			Data: []mdsv2.IamV2RoleBinding{
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:u-epo7ml"),
					RoleName:   mdsv2.PtrString("MetricsViewer"),
					CrnPattern: mdsv2.PtrString("crn_pattern")},
			},
		},
		mockedListUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{iamv2.IamV2User{
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
		mockRoleBindingsResult: mdsv2.IamV2RoleBindingList{
			Data: []mdsv2.IamV2RoleBinding{
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("MetricsViewer"),
					CrnPattern: mdsv2.PtrString("crn_pattern")},
			},
		},
		mockedListUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{iamv2.IamV2User{
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
		mockRoleBindingsResult: mdsv2.IamV2RoleBindingList{
			Data: []mdsv2.IamV2RoleBinding{
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:sa-123"),
					RoleName:   mdsv2.PtrString("MetricsViewer"),
					CrnPattern: mdsv2.PtrString("crn_pattern")},
			},
		},
		mockedListUserResult: iamv2.IamV2UserList{},
		expected: []listDisplay{
			{
				Principal: "User:sa-123",
				Role:      "MetricsViewer",
			},
		},
	},
	// Multiple role bindings at various scopes
	{
		mockRoleBindingsResult: mdsv2.IamV2RoleBindingList{
			Data: []mdsv2.IamV2RoleBinding{
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("OrganizationAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet"),
				},
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("EnvironmentAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet/environment=Cyberdyne"),
				},
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("CloudClusterAdmin"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet/environment=Cyberdyne/cloud-cluster=t1000"),
				},
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("ResourceOwner"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet/environment=Cyberdyne/cloud-cluster=t1000/kafka=t1000/topic=connor"),
				},
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("ResourceOwner"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet/environment=Cyberdyne/cloud-cluster=t1000/kafka=t1000/topic=john*"),
				},
				mdsv2.IamV2RoleBinding{Principal: mdsv2.PtrString("User:" + v1.MockUserResourceId),
					RoleName:   mdsv2.PtrString("DeveloperRead"),
					CrnPattern: mdsv2.PtrString("crn://confluent.cloud/organization=Skynet/environment=Cyberdyne/cloud-cluster=t1000/connector=c1000"),
				},
			},
		},
		mockedListUserResult: iamv2.IamV2UserList{
			Data: []iamv2.IamV2User{iamv2.IamV2User{
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
				CloudCluster:   "t1000",
				ClusterType:    "",
				LogicalCluster: "",
				ResourceType:   "Connector",
				Name:           "c1000",
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
				PatternType:    "PREFIXED",
			},
		},
	},
}

func (suite *RoleBindingTestSuite) TestMyRoleBindingsList() {
	mockeRoleBindingsResult := make(chan mdsv2.IamV2RoleBindingList)
	mockeListUserResult := make(chan iamv2.IamV2UserList)
	for _, tc := range myRoleBindingListTests {
		cmd := suite.newMockIamListRoleBindingCmd(mockeRoleBindingsResult, mockeListUserResult)

		go func() {
			mockeRoleBindingsResult <- tc.mockRoleBindingsResult
			mockeListUserResult <- tc.mockedListUserResult
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
		err:  errNotFound,
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
		args:      []string{"--principal", "User:" + v1.MockUserResourceId, "--role", "ResourceOwner", "--environment", env123, "--cloud-cluster", "lkc-123", "--kafka-cluster-id", "lkc-123"},
		principal: "User:" + v1.MockUserResourceId,
		roleName:  "ResourceOwner",
		scope: mdsv2alpha1.Scope{
			Path:     []string{"organization=" + v1.MockOrgResourceId, "environment=env-123", "cloud-cluster=lkc-123"},
			Clusters: mdsv2alpha1.ScopeClusters{KafkaCluster: "lkc-123"},
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
				copy := expectedListCmdArgs{
					tc.principal, tc.roleName, tc.scope,
				}
				fmt.Println("")
				expect <- copy
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
	expect := make(chan expectedListCmdArgs)
	for _, tc := range roleBindingCreateDeleteTests {
		cmd := suite.newMockIamRoleBindingCmd(expect, "")
		cmd.SetArgs(append([]string{"rbac", "role-binding", "delete"}, tc.args...))

		if tc.err == nil {
			go func() {
				copy := expectedListCmdArgs{
					tc.principal, tc.roleName, tc.scope,
				}
				fmt.Println("")
				expect <- copy
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

func TestRoleBindingTestSuite(t *testing.T) {
	suite.Run(t, new(RoleBindingTestSuite))
}
