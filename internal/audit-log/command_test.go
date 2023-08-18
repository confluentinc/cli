package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	climock "github.com/confluentinc/cli/v3/mock"
	"github.com/confluentinc/cli/v3/pkg/config"
)

var (
	timeNow = time.Now()
	getSpec = mdsv1.AuditLogConfigSpec{
		Destinations: mdsv1.AuditLogConfigDestinations{
			BootstrapServers: []string{"one:8090"},
			Topics: map[string]mdsv1.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 10 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes: &map[string]mdsv1.AuditLogConfigRouteCategories{},
		Metadata: &mdsv1.AuditLogConfigMetadata{
			ResourceVersion: "one",
			UpdatedAt:       &timeNow,
		},
	}
	putSpec = mdsv1.AuditLogConfigSpec{
		Destinations: mdsv1.AuditLogConfigDestinations{
			BootstrapServers: []string{"two:8090"},
			Topics: map[string]mdsv1.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 20 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes: &map[string]mdsv1.AuditLogConfigRouteCategories{},
		Metadata: &mdsv1.AuditLogConfigMetadata{
			ResourceVersion: "one",
			UpdatedAt:       &timeNow,
		},
	}
	putResponseSpec = mdsv1.AuditLogConfigSpec{
		Destinations: mdsv1.AuditLogConfigDestinations{
			BootstrapServers: []string{"localhost:8090"},
			Topics: map[string]mdsv1.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 30 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes:   &map[string]mdsv1.AuditLogConfigRouteCategories{},
		Metadata: &mdsv1.AuditLogConfigMetadata{},
	}
)

type AuditConfigTestSuite struct {
	suite.Suite
	conf    *config.Config
	mockApi mdsv1.AuditLogConfigurationApi
}

type ApiFunc string

const (
	GetConfig            ApiFunc = "GetConfig"
	PutConfig            ApiFunc = "PutConfig"
	ListRoutes           ApiFunc = "ListRoutes"
	ResolveResourceRoute ApiFunc = "ResolveResourceRoute"
)

type MockCall struct {
	Func   ApiFunc
	Input  any
	Result any
}

func (suite *AuditConfigTestSuite) SetupSuite() {
	suite.conf = config.AuthenticatedOnPremConfigMock()
}

func (suite *AuditConfigTestSuite) TearDownSuite() {
}

func StripTimestamp(obj any) any {
	spec, castOk := obj.(mdsv1.AuditLogConfigSpec)
	if castOk {
		return mdsv1.AuditLogConfigSpec{
			Destinations:       spec.Destinations,
			ExcludedPrincipals: spec.ExcludedPrincipals,
			DefaultTopics:      spec.DefaultTopics,
			Routes:             spec.Routes,
			Metadata: &mdsv1.AuditLogConfigMetadata{
				ResourceVersion: spec.Metadata.ResourceVersion,
			},
		}
	} else {
		return obj
	}
}

func (suite *AuditConfigTestSuite) mockCmdReceiver(expect chan MockCall, expectedFunc ApiFunc, expectedInput any) (any, error) {
	if !assert.Greater(suite.T(), len(expect), 0) {
		return nil, fmt.Errorf("unexpected call to %#v", expectedFunc)
	}
	mockCall := <-expect
	if !assert.Equal(suite.T(), expectedFunc, mockCall.Func) {
		return nil, fmt.Errorf("unexpected call to %#v", expectedFunc)
	}
	if !assert.Equal(suite.T(), StripTimestamp(expectedInput), StripTimestamp(mockCall.Input)) {
		return nil, fmt.Errorf("unexpected input to %#v", expectedFunc)
	}
	return mockCall.Result, nil
}

func (suite *AuditConfigTestSuite) newMockCmd(expect chan MockCall) *cobra.Command {
	suite.mockApi = &mock.AuditLogConfigurationApi{
		GetConfigFunc: func(ctx context.Context) (mdsv1.AuditLogConfigSpec, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, GetConfig, nil)
			if err != nil {
				return mdsv1.AuditLogConfigSpec{}, nil, nil
			}
			castResult, ok := result.(mdsv1.AuditLogConfigSpec)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for GetConfig")
				return mdsv1.AuditLogConfigSpec{}, nil, nil
			}
		},
		ListRoutesFunc: func(ctx context.Context, opts *mdsv1.ListRoutesOpts) (mdsv1.AuditLogConfigListRoutesResponse, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, ListRoutes, opts)
			if err != nil {
				return mdsv1.AuditLogConfigListRoutesResponse{}, nil, nil
			}
			castResult, ok := result.(mdsv1.AuditLogConfigListRoutesResponse)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for ListRoutes")
				return mdsv1.AuditLogConfigListRoutesResponse{}, nil, nil
			}
		},
		PutConfigFunc: func(ctx context.Context, spec mdsv1.AuditLogConfigSpec) (mdsv1.AuditLogConfigSpec, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, PutConfig, spec)
			if err != nil {
				return mdsv1.AuditLogConfigSpec{}, nil, nil
			}
			castResult, ok := result.(mdsv1.AuditLogConfigSpec)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for PutConfig")
				return mdsv1.AuditLogConfigSpec{}, nil, nil
			}
		},
		ResolveResourceRouteFunc: func(ctx context.Context, opts *mdsv1.ResolveResourceRouteOpts) (mdsv1.AuditLogConfigResolveResourceRouteResponse, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, ResolveResourceRoute, opts)
			if err != nil {
				return mdsv1.AuditLogConfigResolveResourceRouteResponse{}, nil, nil
			}
			castResult, ok := result.(mdsv1.AuditLogConfigResolveResourceRouteResponse)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for ResolveResourceRoute")
				return mdsv1.AuditLogConfigResolveResourceRouteResponse{}, nil, nil
			}
		},
	}
	mdsClient := mdsv1.NewAPIClient(mdsv1.NewConfiguration())
	mdsClient.AuditLogConfigurationApi = suite.mockApi
	return New(climock.NewPreRunnerMock(nil, nil, mdsClient, nil, suite.conf))
}

func TestAuditConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AuditConfigTestSuite))
}

func (suite *AuditConfigTestSuite) TestAuditConfigDescribe() {
	expect := make(chan MockCall, 10)
	expect <- MockCall{GetConfig, nil, getSpec}
	cmd := suite.newMockCmd(expect)
	cmd.SetArgs([]string{"config", "describe"})
	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))
}

func (suite *AuditConfigTestSuite) TestAuditConfigUpdate() {
	tempFile, err := writeToTempFile(putSpec)
	if tempFile != nil {
		defer os.Remove(tempFile.Name())
	}
	if err != nil {
		assert.Fail(suite.T(), err.Error())
		return
	}
	expect := make(chan MockCall, 10)
	expect <- MockCall{PutConfig, putSpec, putResponseSpec}
	mockCmd := suite.newMockCmd(expect)
	mockCmd.SetArgs([]string{"config", "update", "--file", tempFile.Name()})
	err = mockCmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))
}

func (suite *AuditConfigTestSuite) TestAuditConfigUpdateForce() {
	tempFile, err := writeToTempFile(putSpec)
	if tempFile != nil {
		defer os.Remove(tempFile.Name())
	}
	if err != nil {
		assert.Fail(suite.T(), err.Error())
		return
	}
	expect := make(chan MockCall, 10)
	expect <- MockCall{GetConfig, nil, getSpec}
	expect <- MockCall{PutConfig, putSpec, putResponseSpec}
	mockCmd := suite.newMockCmd(expect)
	mockCmd.SetArgs([]string{"config", "update", "--force", "--file", tempFile.Name()})
	err = mockCmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))
}

func (suite *AuditConfigTestSuite) TestAuditConfigRouteList() {
	devNull := ""
	bothToDevNull := mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &devNull, Denied: &devNull}
	authorizeToDevNull := mdsv1.AuditLogConfigRouteCategories{Authorize: &bothToDevNull}

	expect := make(chan MockCall, 10)
	expect <- MockCall{
		Func: ListRoutes,
		Input: &mdsv1.ListRoutesOpts{
			Q: optional.NewString("crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test"),
		},
		Result: mdsv1.AuditLogConfigListRoutesResponse{
			DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
				Allowed: "confluent-audit-log-events",
				Denied:  "confluent-audit-log-events",
			},
			Routes: &map[string]mdsv1.AuditLogConfigRouteCategories{
				"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test/connector=from-db4": authorizeToDevNull,
				"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test/connector=*":        authorizeToDevNull,
				"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=*/connector=*":              authorizeToDevNull,
				"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-*":                       authorizeToDevNull,
				"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=*":                          authorizeToDevNull,
				"crn://mds1.example.com/kafka=*/connect=qa-*":                                            authorizeToDevNull,
				"crn://mds1.example.com/kafka=*/connect=qa-*/connector=*":                                authorizeToDevNull,
			},
		},
	}
	cmd := suite.newMockCmd(expect)
	cmd.SetArgs([]string{"route", "list", "--resource", "crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test"})
	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))
}

func (suite *AuditConfigTestSuite) TestAuditConfigRouteLookup() {
	defaultTopic := "confluent-audit-log-events"
	devNullTopic := ""
	expect := make(chan MockCall, 10)
	expect <- MockCall{
		Func: ResolveResourceRoute,
		Input: &mdsv1.ResolveResourceRouteOpts{
			Crn: optional.NewString("crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test"),
		},
		Result: mdsv1.AuditLogConfigResolveResourceRouteResponse{
			Route: "default",
			Categories: mdsv1.AuditLogConfigRouteCategories{
				Management: &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &defaultTopic, Denied: &defaultTopic},
				Authorize:  &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &defaultTopic, Denied: &defaultTopic},
				Produce:    &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
				Consume:    &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
				Describe:   &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
			},
		},
	}
	cmd := suite.newMockCmd(expect)
	cmd.SetArgs([]string{"route", "lookup", "crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test"})
	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))
}

func writeToTempFile(spec mdsv1.AuditLogConfigSpec) (*os.File, error) {
	fileBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(os.TempDir(), "test")
	if err != nil {
		return file, err
	}
	if _, err := file.Write(fileBytes); err != nil {
		return file, err
	}
	if err := file.Sync(); err != nil {
		return file, err
	}
	return file, nil
}
