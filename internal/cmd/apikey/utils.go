package apikey

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	kafkaResourceType = "kafka"
	srResourceType    = "schema-registry"
	ksqlResourceType  = "ksql"
)

func (c *command) resolveResourceID(cmd *cobra.Command, args []string) (resourceType string, accId string, clusterId string, currentKey string, err error) {
	resource, err := cmd.Flags().GetString(resourceFlagName)
	if resource == "" || err != nil {
		return "", "", "", "", err
	}
	// If resource is schema registry
	if strings.HasPrefix(resource, "lsrc-") {
		src, err := pcmd.GetSchemaRegistry(cmd, c.ch)
		if err != nil {
			return "", "", "", "", err
		}
		if src == nil {
			return "", "", "", "", errors.ErrNoSrEnabled
		}
		clusterInContext, _ := c.config.SchemaRegistryCluster()
		if clusterInContext == nil || clusterInContext.SrCredentials == nil {
			currentKey = ""
		} else {
			currentKey = clusterInContext.SrCredentials.Key
		}
		return srResourceType, src.AccountId, src.Id, currentKey, nil

	} else if strings.HasPrefix(resource, "lksqlc-") {
		ksql, err := pcmd.GetKSQL(cmd, c.ch)
		if err != nil {
			return "", "", "", "", err
		}
		if ksql == nil {
			return "", "", "", "", errors.ErrNoKSQL
		}
		return ksqlResourceType, ksql.AccountId, ksql.Id, "", nil
	} else {
		kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, resourceFlagName)
		if err != nil {
			return "", "", "", "", err
		}
		return kafkaResourceType, c.config.Auth.Account.Id, kcc.ID, kcc.APIKey, nil
	}
}
