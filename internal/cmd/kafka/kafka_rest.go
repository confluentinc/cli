package kafka

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/dghubble/sling"
)

const kafkaRestPort = "8090"

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func kafkaRestHttpError(httpResp *http.Response) error {
	return errors.NewErrorWithSuggestions(
		fmt.Sprintf(errors.KafkaRestErrorMsg, httpResp.Request.Method, httpResp.Request.URL, httpResp.Status),
		errors.InternalServerErrorSuggestions)
}

func parseOpenAPIError(err error) (*kafkaRestV3Error, error) {
	if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
		var decodedError kafkaRestV3Error
		err = json.Unmarshal(openAPIError.Body(), &decodedError)
		if err != nil {
			return nil, err
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}

func kafkaRestError(url string, err error, httpResp *http.Response) error {
	switch err.(type) {
	case *neturl.Error:
		if e, ok := err.(*neturl.Error); ok {
			return errors.Errorf(errors.InvalidFlagValueWithWrappedErrorErrorMsg, url, "url", e.Err)
		}
	case kafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := parseOpenAPIError(err)
		if parseErr == nil {
			return fmt.Errorf("REST request failed: %v (%v)", openAPIError.Message, openAPIError.Code)
		}
		if httpResp != nil && httpResp.StatusCode >= 400 {
			return kafkaRestHttpError(httpResp)
		}
		return errors.NewErrorWithSuggestions(errors.UnknownErrorMsg, errors.InternalServerErrorSuggestions)
	}
	return err
}

func bootstrapServersToRestURL(bootstrap string) (string, error) {
	bootstrapServers := strings.Split(bootstrap, ",")

	server := bootstrapServers[0]
	serverLength := len(server)
	if _, err := strconv.Atoi(server[serverLength-4:]); err == nil && serverLength > 5 && server[serverLength-5] == ':' {
		return "https://" + server[:serverLength-4] + kafkaRestPort + "/kafka/v3", nil
	}

	return "", errors.New(errors.InvalidBootstrapServerErrorMsg)
}

func getAccessToken(authenticatedState *v2.ContextState, server string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	_, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).Body(strings.NewReader("{}")).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}

	return responses.Token, nil
}

// Converts ACLBinding to Kafka REST GET parameters
func aclBindingToClustersClusterIdAclsGetOpts(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsGetOpts {
	var kafkaRestConfig kafkarestv3.ClustersClusterIdAclsGetOpts

	if acl.Pattern.ResourceType.String() != "UNKNOWN" {
		kafkaRestConfig.ResourceType = optional.NewInterface(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	kafkaRestConfig.ResourceName = optional.NewString(acl.Pattern.Name)

	if acl.Pattern.PatternType.String() != "UNKNOWN" {
		kafkaRestConfig.PatternType = optional.NewInterface(kafkarestv3.AclPatternType(acl.Pattern.PatternType.String()))
	}

	kafkaRestConfig.Principal = optional.NewString(acl.Entry.Principal)
	kafkaRestConfig.Host = optional.NewString(acl.Entry.Host)

	if acl.Entry.Operation.String() != "UNKNOWN" {
		kafkaRestConfig.Operation = optional.NewInterface(kafkarestv3.AclOperation(acl.Entry.Operation.String()))
	}

	if acl.Entry.PermissionType.String() != "UNKNOWN" {
		kafkaRestConfig.Permission = optional.NewInterface(kafkarestv3.AclPermission(acl.Entry.PermissionType.String()))
	}

	return kafkaRestConfig
}

// Converts ACLBinding to Kafka REST POST parameters
func aclBindingToClustersClusterIdAclsPostOpts(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsPostOpts {
	var aclRequestData kafkarestv3.CreateAclRequestData

	if acl.Pattern.ResourceType.String() != "UNKNOWN" {
		aclRequestData.ResourceType = kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String())
	}

	if acl.Pattern.PatternType.String() != "UNKNOWN" {
		aclRequestData.PatternType = kafkarestv3.AclPatternType(acl.Pattern.PatternType.String())
	}

	aclRequestData.ResourceName = acl.Pattern.Name
	aclRequestData.Principal = acl.Entry.Principal
	aclRequestData.Host = acl.Entry.Host

	if acl.Entry.Operation.String() != "UNKNOWN" {
		aclRequestData.Operation = kafkarestv3.AclOperation(acl.Entry.Operation.String())
	}

	if acl.Entry.PermissionType.String() != "UNKNOWN" {
		aclRequestData.Permission = kafkarestv3.AclPermission(acl.Entry.PermissionType.String())
	}

	var kafkaRestConfig kafkarestv3.ClustersClusterIdAclsPostOpts
	kafkaRestConfig.CreateAclRequestData = optional.NewInterface(aclRequestData)

	return kafkaRestConfig
}

// Converts ACLFilter to Kafka REST DELETE parameters
func aclFilterToClustersClusterIdAclsDeleteOpts(acl *schedv1.ACLFilter) kafkarestv3.ClustersClusterIdAclsDeleteOpts {
	var kafkaRestConfig kafkarestv3.ClustersClusterIdAclsDeleteOpts

	if acl.PatternFilter.ResourceType.String() != "UNKNOWN" {
		kafkaRestConfig.ResourceType = optional.NewInterface(kafkarestv3.AclResourceType(acl.PatternFilter.ResourceType.String()))
	}

	kafkaRestConfig.ResourceName = optional.NewString(acl.PatternFilter.Name)

	if acl.PatternFilter.PatternType.String() != "UNKNOWN" {
		kafkaRestConfig.PatternType = optional.NewInterface(kafkarestv3.AclPatternType(acl.PatternFilter.PatternType.String()))
	}

	kafkaRestConfig.Principal = optional.NewString(acl.EntryFilter.Principal)
	kafkaRestConfig.Host = optional.NewString(acl.EntryFilter.Host)

	if acl.EntryFilter.Operation.String() != "UNKNOWN" {
		kafkaRestConfig.Operation = optional.NewInterface(kafkarestv3.AclOperation(acl.EntryFilter.Operation.String()))
	}

	if acl.EntryFilter.PermissionType.String() != "UNKNOWN" {
		kafkaRestConfig.Permission = optional.NewInterface(kafkarestv3.AclPermission(acl.EntryFilter.PermissionType.String()))
	}

	return kafkaRestConfig
}
