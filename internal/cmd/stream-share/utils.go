package streamshare

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/crn"
)

const CrnCcloudAuthority = "confluent.cloud"

type privateLinkNetworkDetails struct {
	networkKind         string
	privateLinkDataType string
	privateLinkData     string
}

func getPrivateLinkNetworkDetails(network cdxv1.CdxV1Network) *privateLinkNetworkDetails {
	cloud := network.GetCloud()
	var details privateLinkNetworkDetails
	if cloud.CdxV1AwsNetwork != nil {
		details.networkKind = cloud.CdxV1AwsNetwork.Kind
		details.privateLinkDataType = "Private Link Endpoint Service"
		details.privateLinkData = cloud.CdxV1AwsNetwork.GetPrivateLinkEndpointService()
	} else if cloud.CdxV1AzureNetwork != nil {
		details.networkKind = cloud.CdxV1AzureNetwork.Kind
		details.privateLinkDataType = "Private Link Service Aliases"
		details.privateLinkData = fmt.Sprintf("%v", cloud.CdxV1AzureNetwork.GetPrivateLinkServiceAliases())
	} else if cloud.CdxV1GcpNetwork != nil {
		details.networkKind = cloud.CdxV1GcpNetwork.Kind
		details.privateLinkDataType = "Private Service Connect Service Attachments"
		details.privateLinkData = fmt.Sprintf("%v", cloud.CdxV1GcpNetwork.GetPrivateServiceConnectServiceAttachments())
	}
	return &details
}

func mapSubdomainsToList(m map[string]string) []string {
	var subdomains []string
	for k, v := range m {
		subdomains = append(subdomains, fmt.Sprintf(`%s="%s"`, k, v))
	}

	return subdomains
}

func confirmOptOut(cmd *cobra.Command) (bool, error) {
	f := form.New(
		form.Field{
			ID: "confirmation",
			Prompt: "Are you sure you want to disable Stream Sharing for your organization? " +
				"Existing shares in your organization will not be accessible if Stream Sharing is disabled.",
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(cmd, form.NewPrompt(os.Stdin)); err != nil {
		return false, errors.New(errors.FailedToReadInputErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}

func getSubjectsCRNFromSharedResources(sharedResources []cdxv1.CdxV1ProviderSharedResource) ([]string, error) {
	var crns []string
	for _, s := range sharedResources {
		for _, r := range s.GetResources() {
			crnObj, err := crn.NewFromString(r)
			if err != nil {
				return nil, err
			}
			for _, e := range crnObj.Elements {
				if e.ResourceType == resource.Subject {
					crns = append(crns, r)
				}
			}
		}
	}
	return crns, nil
}

func areSubjectsModified(newSubjectsCRN []string, existingSubjectsCRN []string) error {
	if len(newSubjectsCRN) != len(existingSubjectsCRN) {
		return errors.New(errors.SubjectsListUnmodifiableErrorMsg)
	}

	sort.Strings(newSubjectsCRN)
	sort.Strings(existingSubjectsCRN)

	for i, s := range existingSubjectsCRN {
		if s != newSubjectsCRN[i] {
			return errors.New(errors.SubjectsListUnmodifiableErrorMsg)
		}
	}
	return nil
}

func getTopicCrn(orgId, environment, srCluster, kafkaCluster, topic string) (string, error) {
	elements, err := crn.NewElements(
		crn.CcloudResourceType_ORGANIZATION, orgId,
		crn.CcloudResourceType_ENVIRONMENT, environment,
		crn.CcloudResourceType_SCHEMA_REGISTRY, srCluster,
		crn.CcloudResourceType_KAFKA, kafkaCluster,
		resource.Topic, topic,
	)
	if err != nil {
		return "", err
	}
	name := crn.ConfluentResourceName{
		Authority: CrnCcloudAuthority,
		Elements:  elements,
	}
	return name.String(), nil
}

func getSubjectCrn(orgId, environment, srCluster, subject string) (string, error) {
	elements, err := crn.NewElements(
		crn.CcloudResourceType_ORGANIZATION, orgId,
		crn.CcloudResourceType_ENVIRONMENT, environment,
		crn.CcloudResourceType_SCHEMA_REGISTRY, srCluster,
		resource.Subject, subject,
	)
	if err != nil {
		return "", err
	}
	name := crn.ConfluentResourceName{
		Authority: CrnCcloudAuthority,
		Elements:  elements,
	}
	return name.String(), nil
}
