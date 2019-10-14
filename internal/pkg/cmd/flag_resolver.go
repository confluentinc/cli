package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	ErrUnexpectedStdinPipe = fmt.Errorf("unexpected stdin pipe")
	ErrNoValueSpecified    = fmt.Errorf("no value specified")
	ErrNoPipe              = fmt.Errorf("no pipe")
)

const (
	kafkaResourceType = "kafka"
	srResourceType    = "schema-registry"
)

// FlagResolver reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
type FlagResolver interface {
	ValueFrom(source string, prompt string, secure bool) (string, error)
	ResolveFlags(cmd *cobra.Command, cfg *config.Config) error
}

type FlagResolverImpl struct {
	Prompt Prompt
	Out    io.Writer
}

// ValueFrom reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
func (r *FlagResolverImpl) ValueFrom(source string, prompt string, secure bool) (value string, err error) {
	// Interactively prompt
	if source == "" {
		if prompt == "" {
			return "", ErrNoValueSpecified
		}
		if yes, err := r.Prompt.IsPipe(); err != nil {
			return "", err
		} else if yes {
			return "", ErrUnexpectedStdinPipe
		}

		_, err = fmt.Fprintf(r.Out, prompt)
		if err != nil {
			return "", err
		}
		if secure {
			valueByte, err := r.Prompt.ReadPassword()
			if err != nil {
				return "", err
			}
			value = string(valueByte)
		} else {
			value, err = r.Prompt.ReadString('\n')
		}
		_, err = fmt.Fprintf(r.Out, "\n")
		if err != nil {
			return "", err
		}
		return value, err
	}

	// Read from stdin pipe
	if source == "-" {
		if yes, err := r.Prompt.IsPipe(); err != nil {
			return "", err
		} else if !yes {
			return "", ErrNoPipe
		}
		value, err = r.Prompt.ReadString('\n')
		if err != nil {
			return "", err
		}
		// To remove the final \n
		return value[0 : len(value)-1], nil
	}

	// Read from a file
	if source[0] == '@' {
		filePath := source[1:]
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(b), err
	}

	return source, nil
}

func (r *FlagResolverImpl) ResolveFlags(cmd *cobra.Command, cfg *config.Config) error {
	err := resolveContextFlag(cmd, cfg)
	if err != nil {
		return err
	}
	err = resolveClusterFlag(cmd, cfg)
	if err != nil {
		return err
	}
	err = resolveEnvironmentFlag(cmd, cfg)
	if err != nil {
		return err
	}
	return nil
}

func resolveContextFlag(cmd *cobra.Command, cfg *config.Config) error {
	if cmd.Flags().Changed("context") {
		name, err := cmd.Flags().GetString("context")
		if err != nil {
			return err
		}
		_, err = cfg.FindContext(name)
		if err != nil {
			return err
		}
		cfg.UserSpecifiedContext = name
	}
	return nil
}

func resolveClusterFlag(cmd *cobra.Command, cfg *config.Config) error {
	if cmd.Flags().Changed("cluster") {
		clusterId, err := cmd.Flags().GetString("cluster")
		if err != nil {
			return err
		}
		context := cfg.Context()
		if context == nil {
			return errors.ErrNoContext
		}
		if _, ok := context.KafkaClusters[clusterId]; !ok {
			return fmt.Errorf("kafka cluster '%s' does not exist under context '%s'", clusterId, context.Name)
		}
		context.UserSpecifiedCluster = clusterId
	}
	return nil
}

func resolveEnvironmentFlag(cmd *cobra.Command, cfg *config.Config) error {
	if cmd.Flags().Changed("environment") {
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}
		context := cfg.Context()
		if context == nil {
			return errors.ErrNoContext
		}
		for _, account := range context.State.Auth.Accounts {
			if account.Id == environment {
				context.State.Auth.Account = account
				return nil
			}
		}
		return fmt.Errorf("environment with id '%s' not found in context '%s'", environment, context.Name)
	}
	return nil
}

func resolveResourceID(cmd *cobra.Command, cfg *config.Config) (resourceType string, accId string,
	clusterId string, currentKey string, err error) {
	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return "", "", "", "", err
	}
	// Resource is schema registry.
	if strings.HasPrefix(resource, "lsrc-") {
		// TODO: Set SR cluster in resource.
		
		//src, err := pcmd.GetSchemaRegistry(cmd, c.ch)
		//if err != nil {
		//	return "", "", "", "", err
		//}
		//if src == nil {
		//	return "", "", "", "", errors.ErrNoSrEnabled
		//}
		////clusterInContext, _ := c.config.SchemaRegistryCluster()
		//if clusterInContext == nil || clusterInContext.SrCredentials == nil {
		//	currentKey = ""
		//} else {
		//	currentKey = clusterInContext.SrCredentials.Key
		//}
		//return srResourceType, src.AccountId, src.Id, currentKey, nil
	} else {
		// Resource is Kafka cluster.
		kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, "resource")
		if err != nil {
			return "", "", "", "", err
		}
		state, err := c.config.AuthenticatedState()
		if err != nil {
			return "", "", "", "", err
		}
		return kafkaResourceType, state.Auth.Account.Id, kcc.ID, kcc.APIKey, nil
	}
}
