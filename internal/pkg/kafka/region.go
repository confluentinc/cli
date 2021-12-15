package kafka

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

type region struct {
	CloudId    string
	CloudName  string
	RegionId   string
	RegionName string
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
