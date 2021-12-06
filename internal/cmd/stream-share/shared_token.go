package stream_share

import (
	"fmt"
	"os"

	v1 "github.com/confluentinc/cdx-schema/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

type sharedTokenCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner       pcmd.PreRunner
	analyticsClient analytics.Client
}

// NewSharedTokenCommand returns the sub-command object to perform operations on shared token
func NewSharedTokenCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *sharedTokenCommand {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "shared-token",
			Short: "Perform operations on shared token.",
		}, prerunner)
	cmd := &sharedTokenCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}
	cmd.init()
	return cmd
}

func (c *sharedTokenCommand) init() {
	// create sub-command
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a shared token for a specific topic and recipient.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
	}
	createCmd.Flags().String("consumer-email", "", "Email of consumer of stream share.")
	createCmd.Flags().String("topic", "", "Topic of stream share.")
	createCmd.Flags().String("cluster", "", "Cluster of stream share.")
	_ = createCmd.MarkFlagRequired("consumer-email")
	_ = createCmd.MarkFlagRequired("topic")
	_ = createCmd.MarkFlagRequired("cluster")
	c.AddCommand(createCmd)

	// redeem sub-command
	redeemCmd := &cobra.Command{
		Use:   "redeem",
		Short: "Redeem a stream share token.",
		Long:  "Redeem a stream share token to access a specific topic. Creates a config file at the path specified by the output flag.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.redeem),
	}
	redeemCmd.Flags().String("token", "", "Token received from the producer of topic data.")
	redeemCmd.Flags().String("output", "./consumer.config", "Optional path for config file.")
	_ = createCmd.MarkFlagRequired("token")
	c.AddCommand(redeemCmd)
}

func (c *sharedTokenCommand) redeem(cmd *cobra.Command, _ []string) error {
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return err
	} else if token == "" {
		return errors.New(errors.TokenEmptyErrorMsg)
	}

	outputPath, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	redeemedToken, err := c.Client.StreamShare.RedeemSharedToken(token)
	if err != nil {
		return err
	}

	utils.Println(cmd, "Token redeemed successfully. Use the generated output file to consume from the kafka topic.")

	err = c.createConfigFile(redeemedToken, outputPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *sharedTokenCommand) createConfigFile(redeemedToken *v1.CdxV1RedeemToken, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	var topic, groupPrefix string
	for _, r := range redeemedToken.GetResources() {
		if r.CdxV1SharedGroup != nil {
			groupPrefix = r.CdxV1SharedGroup.GroupPrefix
		}
		if r.CdxV1SharedTopic != nil {
			topic = r.CdxV1SharedTopic.Topic
		}
	}

	_, err = f.WriteString(fmt.Sprintf(
		"bootstrap.servers=%s\n"+
			"sasl.username=%s\n"+
			"sasl.password=%s\n"+
			"topic=%s\n"+
			"group.id=%s.go_demo_group_1\n"+
			"sasl.mechanisms=PLAIN\n"+
			"security.protocol=SASL_SSL\n"+
			"auto.offset.reset=latest",
		redeemedToken.GetKafkaBootstrapUrl(),
		redeemedToken.GetApikey(),
		redeemedToken.GetSecret(),
		topic,
		groupPrefix,
	))
	if err != nil {
		return err
	}

	return nil
}

func (c *sharedTokenCommand) create(cmd *cobra.Command, _ []string) error {
	email, err := cmd.Flags().GetString("consumer-email")
	if err != nil {
		return err
	}
	matched := utils.ValidateEmail(email)
	if !matched {
		return errors.New(errors.BadEmailFormatErrorMsg)
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	} else if topic == "" {
		return errors.New(errors.TopicEmptyErrorMsg)
	}

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	} else if cluster == "" {
		return errors.New(errors.ClusterEmptyErrorMsg)
	}

	createSharedTokenRequest := &v1.CdxV1CreateSharedTokenRequest{
		EnvironmentId:  stringToPtr(c.EnvironmentId()),
		KafkaClusterId: &cluster,
		ConsumerRestriction: &v1.CdxV1CreateSharedTokenRequestConsumerRestrictionOneOf{
			CdxV1Email: &v1.CdxV1Email{
				Kind:  "Email",
				Email: &email,
			},
		},
		Resources: &[]v1.CdxV1SharedResource{
			{
				Crn:       fmt.Sprintf("crn://confluent.cloud/kafka=%s/topic=%s", cluster, topic),
				Operation: "READ",
			},
		},
	}

	sharedToken, err := c.Client.StreamShare.CreateSharedToken(createSharedTokenRequest)
	if err != nil {
		return err
	}
	utils.Println(cmd, fmt.Sprintf("Stream share id: %s created successfully", sharedToken.StreamShare.Id))
	utils.Println(cmd, fmt.Sprintf("To share topic '%s' with %s, use the following token:", topic, email))
	utils.Println(cmd, *sharedToken.Token)

	return nil
}

func stringToPtr(s string) *string {
	return &s
}
