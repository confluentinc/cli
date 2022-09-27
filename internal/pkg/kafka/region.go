package kafka

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

type region struct {
	CloudId    string `human:"Cloud ID" structured:"cloud_id"`
	CloudName  string `human:"Cloud Name" structured:"cloud_name"`
	RegionId   string `human:"Region ID" structured:"region_id"`
	RegionName string `human:"Region Name" structured:"region_name"`
}

var Clouds = []string{"aws", "azure", "gcp"}

func ListRegions(client *ccloud.Client, cloud string) ([]*region, error) {
	metadataList, err := client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return nil, err
	}

	var regions []*region

	for _, metadata := range metadataList {
		if cloud != "" && cloud != metadata.Id {
			continue
		}

		for _, r := range metadata.Regions {
			if !r.IsSchedulable {
				continue
			}

			regions = append(regions, &region{
				CloudId:    metadata.Id,
				CloudName:  metadata.Name,
				RegionId:   r.Id,
				RegionName: r.Name,
			})
		}
	}

	return regions, nil
}
