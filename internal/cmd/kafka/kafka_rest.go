package kafka

import (
	"encoding/json"
	"fmt"
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

const kafkaPort = "8090"

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func handleCommonKafkaRestClientErrors(url string, err error) error {
	switch err.(type) {
	case *neturl.Error: // Handle errors with request url
		if e, ok := err.(*neturl.Error); ok {
			return errors.Errorf(errors.InvalidFlagValueWithWrappedErrorErrorMsg, url, "url", e.Err)
		}
	case kafkarestv3.GenericOpenAPIError:
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkaRestV3Error
			err = json.Unmarshal(openAPIError.Body(), &decodedError)
			if err != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			return fmt.Errorf("Kafka REST Proxy API failed with: error msg: %v error code: %v", decodedError.Message, decodedError.Code)
		}
	}
	return err
}

func bootstrapServersToRestURL(bootstrap string) (string, error) {
	bootstrapServers := strings.Split(bootstrap, ",")

	server := bootstrapServers[0]
	serverLength := len(server)
	if _, err := strconv.Atoi(server[serverLength-4:]); err == nil && serverLength > 5 && server[serverLength-5] == ':' {
		//TODO: change to https when config is fixed
		return "http://" + server[:serverLength-4] + kafkaPort + "/kafka/v3", nil
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

// Converts ACLBindings to Kafka Rest get parameters
func convertAclBindingToGetParams(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsGetOpts {
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

// Converts ACLBindings to Kafka Rest post parameters
func convertAclBindingToPostParams(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsPostOpts {
	var aclRequestData kafkarestv3.CreateAclRequestData

	if acl.Pattern.ResourceType.String() != "UNKNOWN" {
		aclRequestData.ResourceType = kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String())
	}

	if acl.Pattern.PatternType.String() != "UNKNOWN" {
		aclRequestData.PatternType = kafkarestv3.AclPatternType(acl.Pattern.PatternType.String())
	}

	// TODO: ResourceName not specified in SDK, check with Rigel
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

// Converts ACLFilters to Kafka Rest get parameters
func convertAclFilterToGetParams(acl *schedv1.ACLFilter) kafkarestv3.ClustersClusterIdAclsGetOpts {
	var kafkaRestConfig kafkarestv3.ClustersClusterIdAclsGetOpts

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

// Converts ACLFilters to Kafka Rest post parameters
func convertAclFilterToPostParams(acl *schedv1.ACLFilter) kafkarestv3.ClustersClusterIdAclsDeleteOpts {
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
