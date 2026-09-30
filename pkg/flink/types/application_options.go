package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

type ApplicationOptions struct {
	Cloud              bool
	UnsafeTrace        bool
	UserAgent          string
	EnvironmentId      string
	EnvironmentName    string
	OrganizationId     string // Cloud only
	Database           string
	ComputePoolId      string
	ServiceAccountId   string            // Cloud only
	FlinkConfiguration map[string]string // On-prem only
	Verbose            bool
	LSPBaseUrl         string // Cloud only
	GatewayUrl         string // Cloud only
	Context            *config.Context
}

func ParseApplicationOptionsFromSlices(
	configKeys, configValues []string,
) (*ApplicationOptions, error) {
	if len(configKeys) != len(configValues) {
		return nil, errors.NewErrorWithSuggestions(
			fmt.Sprintf(
				"number of config keys %d and config values %d don't match",
				len(configKeys),
				len(configValues),
			),
			"please provide the same number of config keys and values",
		)
	}
	availabaleFields := getAvailableConfigFields()

	// use reflections to find the fields in ApplicationOptions and assign the right value type to them
	appOptions := &ApplicationOptions{}
	appOptionsReflection := reflect.ValueOf(appOptions).Elem()
	for i, configKey := range configKeys {
		configValue := configValues[i]

		// find the field by name
		field := appOptionsReflection.FieldByName(configKey)
		if !field.IsValid() {
			return nil, errors.NewErrorWithSuggestions(
				fmt.Sprintf("config-key %s not found in ApplicationOptions", configKey),
				fmt.Sprintf(
					"double check the provided config-keys match the available fields %s",
					strings.Join(availabaleFields, ", "),
				),
			)
		}

		// convert the field value to the appropriate type
		switch field.Kind() {
		case reflect.Bool:
			value, err := strconv.ParseBool(configValue)
			if err != nil {
				return nil, fmt.Errorf("config-value %s cannot be parsed to bool", configValue)
			}
			field.SetBool(value)
		case reflect.String:
			field.SetString(configValue)
		default:
			return nil, fmt.Errorf(
				"field type %v of config-key %s is unsupported",
				field.Kind(),
				configKey,
			)
		}
	}

	return appOptions, nil
}

func getAvailableConfigFields() []string {
	appOptionsType := reflect.TypeOf(ApplicationOptions{})
	var availabaleFields []string
	for i := 0; i < appOptionsType.NumField(); i++ {
		field := appOptionsType.Field(i)
		// only allow string or bool fields
		if field.Type.Kind() == reflect.String || field.Type.Kind() == reflect.Bool {
			availabaleFields = append(availabaleFields, field.Name)
		}
	}
	return availabaleFields
}

func (a *ApplicationOptions) Validate() error {
	var missingOptions []string
	if a.GetEnvironmentId() == "" {
		missingOptions = append(missingOptions, "EnvironmentId")
	}
	if a.GetOrganizationId() == "" {
		missingOptions = append(missingOptions, "OrganizationId")
	}
	if a.GetComputePoolId() == "" {
		missingOptions = append(missingOptions, "ComputePoolId")
	}
	if a.GetGatewayUrl() == "" {
		missingOptions = append(missingOptions, "GatewayURL")
	}
	if len(missingOptions) > 0 {
		return fmt.Errorf("missing required config options: %s", strings.Join(missingOptions, ", "))
	}
	return nil
}

func (a *ApplicationOptions) GetUnsafeTrace() bool {
	if a != nil {
		return a.UnsafeTrace
	}
	return false
}

func (a *ApplicationOptions) GetUserAgent() string {
	if a != nil {
		return a.UserAgent
	}
	return ""
}

func (a *ApplicationOptions) GetEnvironmentId() string {
	if a != nil {
		return a.EnvironmentId
	}
	return ""
}

func (a *ApplicationOptions) GetEnvironmentName() string {
	if a != nil {
		return a.EnvironmentName
	}
	return ""
}

func (a *ApplicationOptions) GetOrganizationId() string {
	if a != nil {
		return a.OrganizationId
	}
	return ""
}

func (a *ApplicationOptions) GetDatabase() string {
	if a != nil {
		return a.Database
	}
	return ""
}

func (a *ApplicationOptions) GetComputePoolId() string {
	if a != nil {
		return a.ComputePoolId
	}
	return ""
}

func (a *ApplicationOptions) GetFlinkConfiguration() map[string]string {
	if a != nil {
		return a.FlinkConfiguration
	}
	return nil
}

func (a *ApplicationOptions) GetServiceAccountId() string {
	if a != nil {
		return a.ServiceAccountId
	}
	return ""
}

func (a *ApplicationOptions) GetVerbose() bool {
	if a != nil {
		return a.Verbose
	}
	return false
}

func (a *ApplicationOptions) GetContext() *config.Context {
	if a != nil {
		return a.Context
	}
	return nil
}

func (a *ApplicationOptions) GetLSPBaseUrl() string {
	if a != nil {
		return a.LSPBaseUrl
	}
	return ""
}

func (a *ApplicationOptions) GetGatewayUrl() string {
	if a != nil {
		return a.GatewayUrl
	}
	return ""
}
