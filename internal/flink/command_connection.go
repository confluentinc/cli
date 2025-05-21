package flink

import (
	"fmt"
	"slices"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const envNotFoundErrorMsg = "Failed to get environment '%s'. List available environments with `confluent environment list`."

type connectionOut struct {
	CreationDate time.Time `human:"Creation Date" serialized:"creation_date"`
	Name         string    `human:"Name" serialized:"name"`
	Environment  string    `human:"Environment" serialized:"environment"`
	Cloud        string    `human:"Cloud" serialized:"cloud"`
	Region       string    `human:"Region" serialized:"region"`
	Type         string    `human:"Type" serialized:"type"`
	Endpoint     string    `human:"Endpoint" serialized:"endpoint"`
	Data         string    `human:"Data" serialized:"data"`
	Status       string    `human:"Status" serialized:"status"`
	StatusDetail string    `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
}

func (c *command) newConnectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "connection",
		Short:       "Manage Flink connections.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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

	return c.validConnectionArgsMultiple(cmd, args)
}

func (c *command) validConnectionArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return nil
	}

	connections, err := client.ListConnections(environmentId, c.Context.GetCurrentOrganization(), "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connections))
	for i, connection := range connections {
		suggestions[i] = connection.GetName()
	}
	return suggestions
}

func AddConnectionSecretFlags(cmd *cobra.Command) {
	cmd.Flags().String("api-key", "", fmt.Sprintf("Specify API key for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["api-key"], "or")))
	cmd.Flags().String("aws-access-key", "", fmt.Sprintf("Specify access key for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["aws-access-key"], "or")))
	cmd.Flags().String("aws-secret-key", "", fmt.Sprintf("Specify secret key for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["aws-secret-key"], "or")))
	cmd.Flags().String("aws-session-token", "", fmt.Sprintf("Specify session token for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["aws-session-token"], "or")))
	cmd.Flags().String("service-key", "", fmt.Sprintf("Specify service key for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["service-key"], "or")))
	cmd.Flags().String("username", "", fmt.Sprintf("Specify username for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["username"], "or")))
	cmd.Flags().String("password", "", fmt.Sprintf("Specify password for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["password"], "or")))
	cmd.Flags().String("token", "", fmt.Sprintf("Specify bearer token for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["token"], "or")))
	cmd.Flags().String("token-endpoint", "", fmt.Sprintf("Specify OAuth2 token endpoint for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["token-endpoint"], "or")))
	cmd.Flags().String("client-id", "", fmt.Sprintf("Specify OAuth2 client ID for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["client-id"], "or")))
	cmd.Flags().String("client-secret", "", fmt.Sprintf("Specify OAuth2 client secret for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["client-secret"], "or")))
	cmd.Flags().String("scope", "", fmt.Sprintf("Specify OAuth2 scope for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["scope"], "or")))
	cmd.Flags().Bool("no-auth", false, fmt.Sprintf("Specify no authentication for the type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["no-auth"], "or")))
	cmd.MarkFlagsRequiredTogether("username", "password")
	cmd.MarkFlagsRequiredTogether("aws-access-key", "aws-secret-key")
	cmd.MarkFlagsRequiredTogether("token-endpoint", "client-id", "client-secret", "scope")
	cmd.MarkFlagsMutuallyExclusive("username", "client-id", "api-key", "token", "no-auth")
}

func validateConnectionType(connectionType string) error {
	if !slices.Contains(flink.ConnectionTypes, connectionType) {
		return errors.NewErrorWithSuggestions("invalid connection type "+connectionType, fmt.Sprintf("Specify the connection type as %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionTypes, "or")))
	}
	return nil
}

func validateConnectionSecrets(cmd *cobra.Command, connectionType string) (map[string]string, error) {
	var connectionSecrets []string
	connectionSecrets = append(connectionSecrets, flink.ConnectionTypeSecretMapping[connectionType]...)

	for key := range flink.ConnectionSecretTypeMapping {
		if key == "no-auth" {
			secret, err := cmd.Flags().GetBool(key)
			if err != nil {
				return nil, err
			}
			if secret && !slices.Contains(connectionSecrets, key) {
				return nil, errors.NewErrorWithSuggestions(fmt.Sprintf("%s is invalid for connection %s.", key, connectionType), fmt.Sprintf("Valid secret types are %s.", utils.ArrayToCommaDelimitedString(connectionSecrets, "or")))
			}
		} else {
			secret, err := cmd.Flags().GetString(key)
			if err != nil {
				return nil, err
			}
			if secret != "" && !slices.Contains(connectionSecrets, key) {
				return nil, errors.NewErrorWithSuggestions(fmt.Sprintf("%s is invalid for connection %s.", key, connectionType), fmt.Sprintf("Valid secret types are %s.", utils.ArrayToCommaDelimitedString(connectionSecrets, "or")))
			}
		}
	}

	requiredSecretKeys := flink.ConnectionRequiredSecretMapping[connectionType]
	var optionalSecretKeys []string
	for _, secretKey := range flink.ConnectionTypeSecretMapping[connectionType] {
		if !slices.Contains(requiredSecretKeys, secretKey) && secretKey != "no-auth" {
			optionalSecretKeys = append(optionalSecretKeys, secretKey)
		}
	}

	secretMap := map[string]string{}
	if noAuth, err := cmd.Flags().GetBool("no-auth"); err != nil {
		return nil, err
	} else if noAuth {
		secretMap["AUTH_TYPE"] = "NO_AUTH"
	} else {
		for _, requiredKey := range requiredSecretKeys {
			secret, err := cmd.Flags().GetString(requiredKey)
			if err != nil {
				return nil, err
			}
			if secret == "" {
				return nil, fmt.Errorf("must provide %s for type %s", requiredKey, connectionType)
			}
			backendKey, ok := flink.ConnectionSecretBackendKeyMapping[requiredKey]
			if !ok {
				return nil, fmt.Errorf(`backend key not found for "%s"`, requiredKey)
			}
			secretMap[backendKey] = secret
		}
	}

	for _, key := range lo.Keys(secretMap) {
		switch key {
		case "API_KEY":
			secretMap["AUTH_TYPE"] = key
		case "USERNAME", "PASSWORD":
			secretMap["AUTH_TYPE"] = "BASIC"
		case "TOKEN":
			secretMap["AUTH_TYPE"] = "BEARER"
		case "CLIENT_ID", "CLIENT_SECRET":
			secretMap["AUTH_TYPE"] = "OAUTH2"
		}
	}

	for _, optionalSecretKey := range optionalSecretKeys {
		secret, err := cmd.Flags().GetString(optionalSecretKey)
		if err != nil {
			return nil, err
		}

		backendKey, ok := flink.ConnectionSecretBackendKeyMapping[optionalSecretKey]
		if !ok {
			return nil, fmt.Errorf("backend key not found for %s", optionalSecretKey)
		}

		if secret != "" {
			secretMap[backendKey] = secret
		}
	}

	return secretMap, nil
}
