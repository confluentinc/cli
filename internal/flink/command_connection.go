package flink

import (
	"github.com/spf13/cobra"
	"strings"
)

const (
	OpenAI      string = "openai"
	AzureML     string = "azureml"
	AzureOpenAI string = "azureopenai"
	Bedrock     string = "bedrock"
	Sagemaker   string = "sagemaker"
	GoogleAI    string = "googleai"
	VertexAI    string = "vertexai"
	MongoDB     string = "mongodb"
	Elastic     string = "elastic"
	PineCone    string = "pinecone"
)

const (
	ApiKey       string = "api-key"
	AccessKey    string = "access-key"
	SecretKey    string = "secret-key"
	SessionToken string = "access-token"
	ServiceKey   string = "service-key"
	UserName     string = "username"
	Password     string = "password"
)

func (c *command) newConnectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage Flink connections.",
	}

	cmd.AddCommand(c.newConnectionCreateCommand())
	cmd.AddCommand(c.newConnectionDeleteCommand())
	cmd.AddCommand(c.newConnectionDescribeCommand())
	cmd.AddCommand(c.newConnectionListCommand())
	cmd.AddCommand(c.newConnectionUpdateCommand())

	return cmd
}

func (c *command) validConnectionArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	// TODO
	return []string{}
}

func AddConnectionSecretFlags(cmd *cobra.Command) {
	cmd.Flags().String(ApiKey, "", "Specify api key for the type: "+strings.Join(apiKeyConnectionTypes(), ", "))
	cmd.Flags().String(AccessKey, "", "Specify access key for the type: "+strings.Join(accessKeyConnectionTypes(), ", "))
	cmd.Flags().String(SecretKey, "", "Specify secret key for the type: "+strings.Join(secretKeyConnectionTypes(), ", "))
	cmd.Flags().String(SessionToken, "", "Specify session token for the type: "+strings.Join(sessionTokenConnectionTypes(), ", "))
	cmd.Flags().String(ServiceKey, "", "Specify service key for the type: "+strings.Join(serviceKeyConnectionTypes(), ", "))
	cmd.Flags().String(UserName, "", "Specify username for the type: "+strings.Join(usernameConnectionTypes(), ", "))
	cmd.Flags().String(Password, "", "Specify password for the type: "+strings.Join(passwordConnectionTypes(), ", "))
}

func AddConnectionSecretFlagChecks(cmd *cobra.Command) {
	cmd.MarkFlagsOneRequired(ApiKey, AccessKey, SecretKey, SessionToken, ServiceKey, UserName, Password)
	cmd.MarkFlagsRequiredTogether(AccessKey, SecretKey)
	cmd.MarkFlagsRequiredTogether(UserName, Password)
}

func supportedConnectionTypes() []string {
	return []string{OpenAI, AzureML, AzureOpenAI, Bedrock, Sagemaker, GoogleAI, VertexAI, Elastic, MongoDB, PineCone}
}

func apiKeyConnectionTypes() []string {
	return []string{OpenAI, AzureML, AzureOpenAI, GoogleAI, Elastic, PineCone}
}

func accessKeyConnectionTypes() []string {
	return []string{Bedrock, Sagemaker}
}

func secretKeyConnectionTypes() []string {
	return []string{Bedrock, Sagemaker}
}

func sessionTokenConnectionTypes() []string {
	return []string{Bedrock, Sagemaker}
}

func serviceKeyConnectionTypes() []string {
	return []string{VertexAI}
}

func usernameConnectionTypes() []string {
	return []string{MongoDB}
}

func passwordConnectionTypes() []string {
	return []string{MongoDB}
}
