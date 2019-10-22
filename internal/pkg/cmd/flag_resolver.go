package cmd

import (
	"context"
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
		return errors.HandleCommon(err, cmd)
	}
	err = resolveClusterFlag(cmd, cfg)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = resolveEnvironmentFlag(cmd, cfg)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return resolveResourceId(cmd, cfg)
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
		ctx := cfg.Context()
		if ctx == nil {
			return errors.ErrNoContext
		}
		if _, err := ctx.FindKafkaCluster(clusterId); err != nil {
			return fmt.Errorf("kafka cluster '%s' does not exist under context '%s'", clusterId, ctx.Name)
		}
		ctx.UserSpecifiedKafkaCluster = clusterId
	}
	return nil
}

func resolveEnvironmentFlag(cmd *cobra.Command, cfg *config.Config) error {
	if cmd.Flags().Changed("environment") {
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}
		ctx := cfg.Context()
		if ctx == nil {
			return errors.ErrNoContext
		}
		for _, account := range ctx.State.Auth.Accounts {
			if account.Id == environment {
				ctx.State.Auth.Account = account
				return nil
			}
		}
		err = fmt.Errorf("environment with id '%s' not found in context '%s'", environment, ctx.Name)
		return err
	}
	return nil
}

func resolveResourceId(cmd *cobra.Command, cfg *config.Config) error {
	const resourceFlag = "resource"
	if !cmd.Flags().Changed(resourceFlag) {
		return nil
	}
	resource, err := cmd.Flags().GetString(resourceFlag)
	if err != nil {
		return err
	}
	ctx := cfg.Context()
	if ctx == nil {
		return errors.New("must have an existing context to use --resource flag")
	}
	// Resource is schema registry.
	if strings.HasPrefix(resource, "lsrc-") {
		for envId, srCluster := range ctx.SchemaRegistryClusters {
			if srCluster.Id == resource {
				ctx.UserSpecifiedSchemaRegistryEnvId = envId
			}
		}
		if ctx.UserSpecifiedSchemaRegistryEnvId == "" {
			// Query API by resource ID and env ID.
			state, err := ctx.AuthenticatedState()
			if err != nil {
				return err
			}
			accountId := state.Auth.Account.Id
			ctxClient := config.NewContextClient(ctx)
			srCluster, err := ctxClient.FetchSchemaRegistryById(context.Background(), resource, accountId)
			if err != nil {
				return err
			}
			cluster := &config.SchemaRegistryCluster{
				Id:                     srCluster.Id,
				SchemaRegistryEndpoint: srCluster.Endpoint,
				SrCredentials:          nil, // For now.
			}
			ctx.SchemaRegistryClusters[accountId] = cluster
			ctx.UserSpecifiedSchemaRegistryEnvId = accountId
			return ctx.Save()
		}
	} else {
		// Resource is Kafka cluster.
		return ctx.SetUserSpecifiedKafkaCluster(resource)
	}
	return nil
}
