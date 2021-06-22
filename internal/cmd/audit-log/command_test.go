package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

var (
	timeNow = time.Now()
	getSpec = mds.AuditLogConfigSpec{
		Destinations: mds.AuditLogConfigDestinations{
			BootstrapServers: []string{"one:8090"},
			Topics: map[string]mds.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 10 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mds.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes: &map[string]mds.AuditLogConfigRouteCategories{},
		Metadata: &mds.AuditLogConfigMetadata{
			ResourceVersion: "one",
			UpdatedAt:       &timeNow,
		},
	}
	putSpec = mds.AuditLogConfigSpec{
		Destinations: mds.AuditLogConfigDestinations{
			BootstrapServers: []string{"two:8090"},
			Topics: map[string]mds.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 20 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mds.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes: &map[string]mds.AuditLogConfigRouteCategories{},
		Metadata: &mds.AuditLogConfigMetadata{
			ResourceVersion: "one",
			UpdatedAt:       &timeNow,
		},
	}
	putResponseSpec = mds.AuditLogConfigSpec{
		Destinations: mds.AuditLogConfigDestinations{
			BootstrapServers: []string{"localhost:8090"},
			Topics: map[string]mds.AuditLogConfigDestinationConfig{
				"confluent-audit-log-events": {
					RetentionMs: 30 * 24 * 60 * 60 * 1000,
				},
			},
		},
		ExcludedPrincipals: &[]string{},
		DefaultTopics: mds.AuditLogConfigDefaultTopics{
			Allowed: "confluent-audit-log-events",
			Denied:  "confluent-audit-log-events",
		},
		Routes:   &map[string]mds.AuditLogConfigRouteCategories{},
		Metadata: &mds.AuditLogConfigMetadata{},
	}
)

type AuditConfigTestSuite struct {
	suite.Suite
	conf    *v3.Config
	mockApi mds.AuditLogConfigurationApi
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
	Input  interface{}
	Result interface{}
}

func (suite *AuditConfigTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedConfluentConfigMock()
}

func (suite *AuditConfigTestSuite) TearDownSuite() {
}

func StripTimestamp(obj interface{}) interface{} {
	spec, castOk := obj.(mds.AuditLogConfigSpec)
	if castOk {
		return mds.AuditLogConfigSpec{
			Destinations:       spec.Destinations,
			ExcludedPrincipals: spec.ExcludedPrincipals,
			DefaultTopics:      spec.DefaultTopics,
			Routes:             spec.Routes,
			Metadata: &mds.AuditLogConfigMetadata{
				ResourceVersion: spec.Metadata.ResourceVersion,
			},
		}
	} else {
		return obj
	}
}

func (suite *AuditConfigTestSuite) mockCmdReceiver(expect chan MockCall, expectedFunc ApiFunc, expectedInput interface{}) (interface{}, error) {
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
		GetConfigFunc: func(ctx context.Context) (mds.AuditLogConfigSpec, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, GetConfig, nil)
			if err != nil {
				return mds.AuditLogConfigSpec{}, nil, nil
			}
			castResult, ok := result.(mds.AuditLogConfigSpec)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for GetConfig")
				return mds.AuditLogConfigSpec{}, nil, nil
			}
		},
		ListRoutesFunc: func(ctx context.Context, opts *mds.ListRoutesOpts) (mds.AuditLogConfigListRoutesResponse, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, ListRoutes, opts)
			if err != nil {
				return mds.AuditLogConfigListRoutesResponse{}, nil, nil
			}
			castResult, ok := result.(mds.AuditLogConfigListRoutesResponse)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for ListRoutes")
				return mds.AuditLogConfigListRoutesResponse{}, nil, nil
			}
		},
		PutConfigFunc: func(ctx context.Context, spec mds.AuditLogConfigSpec) (mds.AuditLogConfigSpec, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, PutConfig, spec)
			if err != nil {
				return mds.AuditLogConfigSpec{}, nil, nil
			}
			castResult, ok := result.(mds.AuditLogConfigSpec)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for PutConfig")
				return mds.AuditLogConfigSpec{}, nil, nil
			}
		},
		ResolveResourceRouteFunc: func(ctx context.Context, opts *mds.ResolveResourceRouteOpts) (mds.AuditLogConfigResolveResourceRouteResponse, *http.Response, error) {
			result, err := suite.mockCmdReceiver(expect, ResolveResourceRoute, opts)
			if err != nil {
				return mds.AuditLogConfigResolveResourceRouteResponse{}, nil, nil
			}
			castResult, ok := result.(mds.AuditLogConfigResolveResourceRouteResponse)
			if ok {
				return castResult, nil, nil
			} else {
				assert.Fail(suite.T(), "unexpected result type for ResolveResourceRoute")
				return mds.AuditLogConfigResolveResourceRouteResponse{}, nil, nil
			}
		},
	}
	mdsClient := mds.NewAPIClient(mds.NewConfiguration())
	mdsClient.AuditLogConfigurationApi = suite.mockApi
	return New("confluent", cliMock.NewPreRunnerMock(nil, mdsClient, nil, suite.conf))
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
	bothToDevNull := mds.AuditLogConfigRouteCategoryTopics{Allowed: &devNull, Denied: &devNull}
	authorizeToDevNull := mds.AuditLogConfigRouteCategories{Authorize: &bothToDevNull}

	expect := make(chan MockCall, 10)
	expect <- MockCall{
		Func: ListRoutes,
		Input: &mds.ListRoutesOpts{
			Q: optional.NewString("crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test")},
		Result: mds.AuditLogConfigListRoutesResponse{
			DefaultTopics: mds.AuditLogConfigDefaultTopics{
				Allowed: "confluent-audit-log-events",
				Denied:  "confluent-audit-log-events",
			},
			Routes: &map[string]mds.AuditLogConfigRouteCategories{
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
		Input: &mds.ResolveResourceRouteOpts{
			Crn: optional.NewString("crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test")},
		Result: mds.AuditLogConfigResolveResourceRouteResponse{
			Route: "default",
			Categories: mds.AuditLogConfigRouteCategories{
				Management: &mds.AuditLogConfigRouteCategoryTopics{Allowed: &defaultTopic, Denied: &defaultTopic},
				Authorize:  &mds.AuditLogConfigRouteCategoryTopics{Allowed: &defaultTopic, Denied: &defaultTopic},
				Produce:    &mds.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
				Consume:    &mds.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
				Describe:   &mds.AuditLogConfigRouteCategoryTopics{Allowed: &devNullTopic, Denied: &devNullTopic},
			},
		},
	}
	cmd := suite.newMockCmd(expect)
	cmd.SetArgs([]string{"route", "lookup", "crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test"})
	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(expect))

}

func writeToTempFile(spec mds.AuditLogConfigSpec) (f *os.File, err error) {
	fileBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	file, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		return file, err
	}
	_, err = file.Write(fileBytes)
	if err != nil {
		return file, err
	}
	if err = file.Sync(); err != nil {
		return file, err
	}
	return file, nil
}
