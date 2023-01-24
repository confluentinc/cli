package kafka

import (
	"context"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
)

type region struct {
	CloudId    string `human:"Cloud ID" serialized:"cloud_id"`
	CloudName  string `human:"Cloud Name" serialized:"cloud_name"`
	RegionId   string `human:"Region ID" serialized:"region_id"`
	RegionName string `human:"Region Name" serialized:"region_name"`
}

func ListRegions(client *ccloudv1.Client, cloud string) ([]*region, error) {
	metadataList, err := client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return nil, err
	}

	var regions []*region

	for _, metadata := range metadataList {
		if cloud != "" && cloud != metadata.GetId() {
			continue
		}

		for _, r := range metadata.GetRegions() {
			if !r.GetIsSchedulable() {
				continue
			}

			regions = append(regions, &region{
				CloudId:    metadata.GetId(),
				CloudName:  metadata.GetName(),
				RegionId:   r.GetId(),
				RegionName: r.GetName(),
			})
		}
	}

	return regions, nil
}
