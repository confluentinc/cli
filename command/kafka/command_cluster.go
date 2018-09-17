package kafka

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codyaray/go-printer"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/common"
	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/shared"
)

var (
	listFields      = []string{"Id", "Name", "ServiceProvider", "Region", "Durability", "Status"}
	listLabels      = []string{"Id", "Name", "Provider", "Region", "Durability", "Status"}
	describeFields  = []string{"Id", "Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "PricePerHour"}
	describeRenames = map[string]string{"NetworkIngress": "Ingress", "NetworkEgress": "Egress", "ServiceProvider": "Provider"}
)

type clusterCommand struct {
	*cobra.Command
	config *shared.Config
	kafka  Kafka
}

// NewClusterCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewClusterCommand(config *shared.Config, kafka Kafka) *cobra.Command {
	cmd := &clusterCommand{
		Command: &cobra.Command{
			Use:   "cluster",
			Short: "Manage kafka clusters.",
		},
		config: config,
		kafka:  kafka,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		RunE:  c.list,
	})

	createCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a Kafka cluster.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().String("cloud", "", "aws or gcp")
	createCmd.Flags().String("region", "", "a valid region in the given cloud")
	// default to smallest size allowed
	createCmd.Flags().Int32("ingress", 1, "network ingress in MB/s")
	createCmd.Flags().Int32("egress", 1, "network egress in MB/s")
	createCmd.Flags().Int32("storage", 500, "total usable data storage in GB")
	createCmd.Flags().Bool("multizone", false, "use multiple zones for high availability")
	check(createCmd.MarkFlagRequired("cloud"))
	check(createCmd.MarkFlagRequired("region"))
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "describe ID",
		Short: "Describe a Kafka cluster.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "update ID",
		Short: "Update a Kafka cluster.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete a Kafka cluster.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Auth a Kafka cluster.",
		RunE:  c.auth,
	})
	c.AddCommand(&cobra.Command{
		Use:   "use ID",
		Short: "Set a Kafka cluster in the CLI context.",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *clusterCommand) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.kafka.List(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	var data [][]string
	for _, cluster := range clusters {
		data = append(data, printer.ToRow(cluster, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return common.HandleError(err)
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return common.HandleError(err)
	}
	ingress, err := cmd.Flags().GetInt32("ingress")
	if err != nil {
		return common.HandleError(err)
	}
	egress, err := cmd.Flags().GetInt32("egress")
	if err != nil {
		return common.HandleError(err)
	}
	storage, err := cmd.Flags().GetInt32("storage")
	if err != nil {
		return common.HandleError(err)
	}
	multizone, err := cmd.Flags().GetBool("multizone")
	if err != nil {
		return common.HandleError(err)
	}
	durability := schedv1.Durability_LOW
	if multizone {
		durability = schedv1.Durability_HIGH
	}
	config := &schedv1.KafkaClusterConfig{
		AccountId:       c.config.Auth.Account.Id,
		Name:            args[0],
		ServiceProvider: cloud,
		Region:          region,
		NetworkIngress:  ingress,
		NetworkEgress:   egress,
		Storage:         storage,
		Durability:      durability,
	}
	cluster, err := c.kafka.Create(context.Background(), config)
	if err != nil {
		// TODO: don't swallow validation errors (reportedly separately)
		return common.HandleError(err)
	}
	printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
	return nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	cluster, err := c.kafka.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
	return nil
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	err := c.kafka.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Printf("The kafka cluster %s has been deleted.\n", args[0])
	return nil
}

func (c *clusterCommand) auth(cmd *cobra.Command, args []string) error {
	cfg, err := c.config.Context()
	if err != nil {
		return common.HandleError(err)
	}
	cluster, known := c.config.Platforms[cfg.Platform].KafkaClusters[cfg.Kafka]
	if known {
		fmt.Printf("Kafka Cluster: %s\n", cfg.Kafka)
		fmt.Printf("Bootstrap Servers: %s\n", cluster.Bootstrap)
		fmt.Printf("API Key: %s\n", cluster.APIKey)
		fmt.Printf("API Secret: %s\n", cluster.APISecret)
		return nil
	}

	userProvidingKey, err := userHasKey(cfg.Kafka)
	if err != nil {
		return common.HandleError(err)
	}

	var key, secret string
	if userProvidingKey {
		key, secret, err = promptForKafkaCreds()
	} else {
		key, secret, err = c.createKafkaCreds(cfg.Kafka)
	}
	if err != nil {
		return common.HandleError(err)
	}

	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id, Id: cfg.Kafka}
	kc, err := c.kafka.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}

	if c.config.Platforms[cfg.Platform].KafkaClusters == nil {
		c.config.Platforms[cfg.Platform].KafkaClusters = map[string]shared.KafkaClusterConfig{}
	}
	c.config.Platforms[cfg.Platform].KafkaClusters[cfg.Kafka] = shared.KafkaClusterConfig{
		Bootstrap: strings.TrimPrefix(kc.Endpoint, "SASL_SSL://"),
		APIKey:    key,
		APISecret: secret,
	}
	return c.config.Save()
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	cfg, err := c.config.Context()
	if err != nil {
		return common.HandleError(err)
	}
	cfg.Kafka = args[0]
	return c.config.Save()
}

//
// Helper functions
//

func userHasKey(kafkaClusterID string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Do you have an API key for %s? [N/y] ", kafkaClusterID)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	r := strings.TrimSpace(response)
	return r == "" || r[0] == 'y' || r[0] == 'Y', nil
}

func promptForKafkaCreds() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("API Key: ")
	key, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("API Secret: ")
	byteSecret, err := terminal.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", "", err
	}
	secret := string(byteSecret)

	return strings.TrimSpace(key), strings.TrimSpace(secret), nil
}

func (c *clusterCommand) createKafkaCreds(kafkaClusterID string) (string, string, error) {
	client := chttp.NewClientWithJWT(context.Background(), c.config.AuthToken, c.config.AuthURL, c.config.Logger)
	key, _, err := client.APIKey.Create(&orgv1.ApiKey{
		UserId:    c.config.Auth.User.Id,
		ClusterId: kafkaClusterID,
	})
	if err != nil {
		return "", "", shared.ConvertAPIError(err)
	}
	fmt.Println("Okay, we've created an API key. If needed, you can see it with `confluent kafka auth`.")
	return key.Key, key.Secret, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
