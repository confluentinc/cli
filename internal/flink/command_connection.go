package flink

import (
	"fmt"
	"slices"
	"time"

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
	cmd.Flags().String("auth-type", "", fmt.Sprintf("Specify authentication type for the type : %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["auth-type"], "or")))
	cmd.Flags().String("bearer-token", "", fmt.Sprintf("Specify bearer token for BEARER authentication type: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["bearer-token"], "or")))
	cmd.Flags().String("oauth2-token-endpoint", "", fmt.Sprintf("Specify oauth2 token endpoint: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["oauth2-token-endpoint"], "or")))
	cmd.Flags().String("oauth2-client-id", "", fmt.Sprintf("Specify oauth2 client id: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["oauth2-client-id"], "or")))
	cmd.Flags().String("oauth2-client-secret", "", fmt.Sprintf("Specify oauth2 client secret: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["oauth2-client-secret"], "or")))
	cmd.Flags().String("oauth2-scope", "", fmt.Sprintf("Specify oauth2 scope: %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionSecretTypeMapping["oauth2-scope"], "or")))
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

	secretMap := map[string]string{}
	var requiredSecretKeys []string
	var optionalSecretKeys []string

	dynamicKey, hasDynamicKey := flink.ConnectionTypeDynamicKeyMapping[connectionType]
	if hasDynamicKey {
		dynamicKeyValue, err := cmd.Flags().GetString(dynamicKey)
		if err != nil {
			return nil, err
		}
		if dynamicKeyValue == "" {
			return nil, fmt.Errorf("must provide %s for connection %s", dynamicKey, connectionType)
		}

		backendKey, ok := flink.ConnectionSecretBackendKeyMapping[dynamicKey]
		if !ok {
			return nil, fmt.Errorf(`backend key not found for "%s"`, dynamicKey)
		}
		secretMap[backendKey] = dynamicKeyValue

		requiredSecretKeys, exists := flink.ConnectionDynamicRequiredSecretMapping[connectionType][dynamicKeyValue]
		if !exists {
			validTypes := make([]string, 0, len(flink.ConnectionDynamicRequiredSecretMapping[connectionType]))
			for k := range flink.ConnectionDynamicRequiredSecretMapping[connectionType] {
				validTypes = append(validTypes, k)
			}
			return nil, errors.NewErrorWithSuggestions(
				fmt.Sprintf("invalid %s %s for connection %s", dynamicKey, dynamicKeyValue, connectionType),
				fmt.Sprintf("Valid types are %s.", utils.ArrayToCommaDelimitedString(validTypes, "or")),
			)
		}

		allPossibleKeys, exists := flink.ConnectionDynamicSecretMapping[connectionType][dynamicKeyValue]
		if exists {
			connectionSecrets = append(connectionSecrets, allPossibleKeys...)
		}

		for key := range flink.ConnectionSecretTypeMapping {
			secret, err := cmd.Flags().GetString(key)
			if err != nil {
				return nil, err
			}
			if secret != "" && !slices.Contains(connectionSecrets, key) {
				return nil, errors.NewErrorWithSuggestions(
					fmt.Sprintf("%s is invalid for connection %s with %s %s.", key, connectionType, dynamicKey, dynamicKeyValue),
					fmt.Sprintf("Valid secret types are %s.", utils.ArrayToCommaDelimitedString(connectionSecrets, "or")),
				)
			}
		}

		for _, secretKey := range allPossibleKeys {
			if !slices.Contains(requiredSecretKeys, secretKey) {
				optionalSecretKeys = append(optionalSecretKeys, secretKey)
			}
		}

		for _, requiredKey := range requiredSecretKeys {
			secret, err := cmd.Flags().GetString(requiredKey)
			if err != nil {
				return nil, err
			}
			if secret == "" {
				return nil, fmt.Errorf("must provide %s for %s %s on connection %s", requiredKey, dynamicKey, dynamicKeyValue, connectionType)
			}
			backendKey, ok := flink.ConnectionSecretBackendKeyMapping[requiredKey]
			if !ok {
				return nil, fmt.Errorf(`backend key not found for "%s"`, requiredKey)
			}
			secretMap[backendKey] = secret
		}

		for _, optionalKey := range optionalSecretKeys {
			secret, err := cmd.Flags().GetString(optionalKey)
			if err != nil {
				return nil, err
			}
			if secret != "" {
				backendKey, ok := flink.ConnectionSecretBackendKeyMapping[optionalKey]
				if !ok {
					return nil, fmt.Errorf(`backend key not found for "%s"`, optionalKey)
				}
				secretMap[backendKey] = secret
			}
		}
	} else {

		for key := range flink.ConnectionSecretTypeMapping {
			secret, err := cmd.Flags().GetString(key)
			if err != nil {
				return nil, err
			}
			if secret != "" && !slices.Contains(connectionSecrets, key) {
				return nil, errors.NewErrorWithSuggestions(
					fmt.Sprintf("%s is invalid for connection %s.", key, connectionType),
					fmt.Sprintf("Valid secret types are %s.", utils.ArrayToCommaDelimitedString(connectionSecrets, "or")),
				)
			}
		}

		requiredSecretKeys = flink.ConnectionRequiredSecretMapping[connectionType]
		for _, secretKey := range flink.ConnectionTypeSecretMapping[connectionType] {
			if !slices.Contains(requiredSecretKeys, secretKey) {
				optionalSecretKeys = append(optionalSecretKeys, secretKey)
			}
		}

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

		for _, optionalKey := range optionalSecretKeys {
			secret, err := cmd.Flags().GetString(optionalKey)
			if err != nil {
				return nil, err
			}
			if secret != "" {
				backendKey, ok := flink.ConnectionSecretBackendKeyMapping[optionalKey]
				if !ok {
					return nil, fmt.Errorf("backend key not found for %s", optionalKey)
				}
				secretMap[backendKey] = secret
			}
		}
	}

	return secretMap, nil
}
