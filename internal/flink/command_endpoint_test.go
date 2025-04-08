package flink

import (
	"reflect"
	"testing"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"
)

func TestBuildCloudRegionKeyFilterMapFromPrivateLinkAttachments(t *testing.T) {
	// Helper function to create a private link attachment with specified cloud and region
	createPlatt := func(cloud, region string) networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment {
		spec := networkingprivatelinkv1.NewNetworkingV1PrivateLinkAttachmentSpec()
		spec.SetCloud(cloud)
		spec.SetRegion(region)

		platt := networkingprivatelinkv1.NewNetworkingV1PrivateLinkAttachment()
		platt.SetSpec(*spec)
		return *platt
	}

	tests := []struct {
		name     string
		platts   []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment
		expected map[CloudRegionKey]bool
	}{
		{
			name:     "Empty slice",
			platts:   []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{},
			expected: map[CloudRegionKey]bool{},
		},
		{
			name: "Single attachment",
			platts: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				createPlatt("AWS", "us-east-1"),
			},
			expected: map[CloudRegionKey]bool{
				{cloud: "AWS", region: "us-east-1"}: true,
			},
		},
		{
			name: "Multiple unique attachments",
			platts: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				createPlatt("AWS", "us-east-1"),
				createPlatt("GCP", "us-central1"),
				createPlatt("AZURE", "eastus"),
			},
			expected: map[CloudRegionKey]bool{
				{cloud: "AWS", region: "us-east-1"}:   true,
				{cloud: "GCP", region: "us-central1"}: true,
				{cloud: "AZURE", region: "eastus"}:    true,
			},
		},
		{
			name: "Duplicate cloud/region combinations",
			platts: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				createPlatt("AWS", "us-east-1"),
				createPlatt("AWS", "us-east-1"),
				createPlatt("AWS", "us-west-1"),
			},
			expected: map[CloudRegionKey]bool{
				{cloud: "AWS", region: "us-east-1"}: true,
				{cloud: "AWS", region: "us-west-1"}: true,
			},
		},
		{
			name: "Empty cloud or region values are skipped",
			platts: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				createPlatt("", "us-east-1"),
				createPlatt("AWS", ""),
				createPlatt("", ""),
				createPlatt("GCP", "us-central1"),
			},
			expected: map[CloudRegionKey]bool{
				{cloud: "GCP", region: "us-central1"}: true,
			},
		},
		{
			name: "Mix of valid and invalid entries",
			platts: []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
				createPlatt("AWS", "us-east-1"),
				createPlatt("", "eu-west-1"),
				createPlatt("AZURE", "eastus"),
				createPlatt("GCP", ""),
			},
			expected: map[CloudRegionKey]bool{
				{cloud: "AWS", region: "us-east-1"}: true,
				{cloud: "AZURE", region: "eastus"}:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(tt.platts)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("buildCloudRegionKeyFilterMapFromPrivateLinkAttachments() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
