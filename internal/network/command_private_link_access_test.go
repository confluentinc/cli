package network

import (
	"testing"

	"github.com/stretchr/testify/assert"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"
)

func TestGetPrivateLinkAccessCloudWithAws(t *testing.T) {
	cloud, err := getPrivateLinkAccessCloud(networkingv1.NetworkingV1PrivateLinkAccess{
		Id: networkingv1.PtrString("pla-123456"),
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Cloud: &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
				NetworkingV1AwsPrivateLinkAccess: &networkingv1.NetworkingV1AwsPrivateLinkAccess{
					Kind:    "AwsPrivateLinkAccess",
					Account: "012345678901",
				},
			},
			DisplayName: networkingv1.PtrString("aws-pla"),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
			Phase: "READY",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, CloudAws, cloud)
}

func TestGetPrivateLinkAccessCloudWithAzure(t *testing.T) {
	cloud, err := getPrivateLinkAccessCloud(networkingv1.NetworkingV1PrivateLinkAccess{
		Id: networkingv1.PtrString("pla-123456"),
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Cloud: &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
				NetworkingV1AzurePrivateLinkAccess: &networkingv1.NetworkingV1AzurePrivateLinkAccess{
					Kind:         "AzurePrivateLinkAccess",
					Subscription: "1234abcd-12ab-34cd-1234-123456abcdef",
				},
			},
			DisplayName: networkingv1.PtrString("aws-pla"),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
			Phase: "READY",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, CloudAzure, cloud)
}

func TestGetPrivateLinkAccessCloudWithGcp(t *testing.T) {
	cloud, err := getPrivateLinkAccessCloud(networkingv1.NetworkingV1PrivateLinkAccess{
		Id: networkingv1.PtrString("pla-123456"),
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Cloud: &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
				NetworkingV1GcpPrivateServiceConnectAccess: &networkingv1.NetworkingV1GcpPrivateServiceConnectAccess{
					Kind:    "GcpPrivateServiceConnectAccess",
					Project: "temp-gear-123456",
				},
			},
			DisplayName: networkingv1.PtrString("aws-pla"),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
			Phase: "READY",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, CloudGcp, cloud)
}

func TestGetPrivateLinkAccessCloudWithEmptyValue(t *testing.T) {
	_, err := getPrivateLinkAccessCloud(networkingv1.NetworkingV1PrivateLinkAccess{})
	assert.Error(t, err)
}
