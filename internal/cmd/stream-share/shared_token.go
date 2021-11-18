package stream_share

import (
	"fmt"
	"github.com/confluentinc/cdx-schema/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

type sharedTokenCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner       pcmd.PreRunner
	logger          *log.Logger
	analyticsClient analytics.Client
}

// NewSharedTokenCommand returns the sub-command object to perform operations on shared token
func NewSharedTokenCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *sharedTokenCommand {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "shared-token",
			Short: "Perform operations on shared token",
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
	createCmd.Flags().String("consumer_email", "", "Email id of consumer of stream share")
	createCmd.Flags().String("topic", "", "Topic of stream share")
	createCmd.Flags().String("cluster", "", "Cluster of stream share")
	c.AddCommand(createCmd)

	// redeem sub-command
	redeemCmd := &cobra.Command{
		Use:   "redeem",
		Short: "Redeem the shared token to access a specific topic",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.redeem),
	}
	redeemCmd.Flags().String("token", "", "Token received from the producer of topic data")
	c.AddCommand(redeemCmd)
}

func (c *sharedTokenCommand) redeem(cmd *cobra.Command, _ []string) error {
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return err
	} else if token == "" {
		return errors.New(errors.TokenEmptyErrorMsg)
	}

	redeemedToken, err := c.Client.StreamShare.RedeemSharedToken(token)
	if err != nil {
		return err
	}

	utils.Println(cmd, "Token redeemed successfully. Use the following information to consume from the kafka topic:")

	utils.Println(cmd, fmt.Sprintf("Bootstrap URL: %s", redeemedToken.GetKafkaBootstrapUrl()))
	utils.Println(cmd, fmt.Sprintf("API Key: %s", redeemedToken.GetApikey()))
	utils.Println(cmd, fmt.Sprintf("API Key Secret: %s", redeemedToken.GetSecret()))

	for _, r := range redeemedToken.GetResources() {
		if r.CdxV1SharedGroup != nil {
			utils.Println(cmd, fmt.Sprintf("Shared Consumer Group Prefix: %s", r.CdxV1SharedGroup.GroupPrefix))
		}
		if r.CdxV1SharedTopic != nil {
			utils.Println(cmd, fmt.Sprintf("Shared Topic: %s", r.CdxV1SharedTopic.Topic))
		}
	}

	return nil
}

func (c *sharedTokenCommand) create(cmd *cobra.Command, _ []string) error {
	email, err := cmd.Flags().GetString("consumer_email")
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
	utils.Println(cmd, fmt.Sprintf("To share topic '%s' with %s, use the following token:", topic, email))
	utils.Println(cmd, *sharedToken.Token)

	return nil
}

func stringToPtr(s string) *string {
	return &s
}
