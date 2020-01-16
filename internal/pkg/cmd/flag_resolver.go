package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
)

var (
	ErrUnexpectedStdinPipe = fmt.Errorf("unexpected stdin pipe")
	ErrNoValueSpecified    = fmt.Errorf("no value specified")
	ErrNoPipe              = fmt.Errorf("no pipe")
)

// FlagResolver reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
type FlagResolver interface {
	ValueFrom(source string, prompt string, secure bool) (string, error)
	ResolveContextFlag(cmd *cobra.Command) (string, error)
	ResolveClusterFlag(cmd *cobra.Command) (string, error)
	ResolveEnvironmentFlag(cmd *cobra.Command) (string, error)
	ResolveResourceId(cmd *cobra.Command) (resourceType string, resourceId string, err error)
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

func (r *FlagResolverImpl) ResolveContextFlag(cmd *cobra.Command) (string, error) {
	const contextFlag = "context"
	if cmd.Flags().Changed(contextFlag) {
		name, err := cmd.Flags().GetString(contextFlag)
		if err != nil {
			return "", err
		}
		return name, nil
	}
	return "", nil
}

func (r *FlagResolverImpl) ResolveClusterFlag(cmd *cobra.Command) (string, error) {
	const clusterFlag = "cluster"
	if cmd.Flags().Changed(clusterFlag) {
		clusterId, err := cmd.Flags().GetString(clusterFlag)
		if err != nil {
			return "", err
		}
		return clusterId, nil
		//if _, err := ctx.FindKafkaCluster(clusterId, client); err != nil {
		//	return fmt.Errorf("kafka cluster '%s' does not exist under context '%s'", clusterId, ctx.Name)
		//}
	}
	return "", nil
}

func (r *FlagResolverImpl) ResolveEnvironmentFlag(cmd *cobra.Command) (string, error) {
	const environmentFlag = "environment"
	if cmd.Flags().Changed(environmentFlag) {
		environment, err := cmd.Flags().GetString(environmentFlag)
		if err != nil {
			return "", err
		}
		return environment, err
	}
	return "", nil
}

const (
	KafkaResourceType = "kafka"
	SrResourceType    = "schema-registry"
)

func (r *FlagResolverImpl) ResolveResourceId(cmd *cobra.Command) (resourceType string, resourceId string, err error) {
	const resourceFlag = "resource"
	if !cmd.Flags().Changed(resourceFlag) {
		return "", "", nil
	}
	resourceId, err = cmd.Flags().GetString(resourceFlag)
	if err != nil {
		return "", "", err
	}
	//if ctx == nil {
	//	return "", "", errors.New("must have an existing context to use --resource flag")
	//}
	// Resource is schema registry.
	if strings.HasPrefix(resourceId, "lsrc-") {
		return SrResourceType, resourceId, nil
		//for envId, srCluster := range ctx.SchemaRegistryClusters {
		//	if srCluster.Id == resourceId {
		//		ctx.UserSpecifiedSchemaRegistryEnvId = envId
		//	}
		//}
		//if ctx.UserSpecifiedSchemaRegistryEnvId == "" {
		//	// Query API by resource ID and env ID.
		//	state, err := ctx.AuthenticatedState()
		//	if err != nil {
		//		return err
		//	}
		//	accountId := state.Auth.Account.Id
		//	ctxClient := config.NewContextClient(ctx, client)
		//	srCluster, err := ctxClient.FetchSchemaRegistryById(context.Background(), resourceId, accountId)
		//	if err != nil {
		//		return err
		//	}
		//	cluster := &config.SchemaRegistryCluster{
		//		Id:                     srCluster.Id,
		//		SchemaRegistryEndpoint: srCluster.Endpoint,
		//		SrCredentials:          nil, // For now.
		//	}
		//	ctx.SchemaRegistryClusters[accountId] = cluster
		//	ctx.UserSpecifiedSchemaRegistryEnvId = accountId
		//	return ctx.Save()
		//}
	} else {
		// Resource is Kafka cluster.
		return KafkaResourceType, resourceId, err
		//return ctx.SetUserSpecifiedKafkaCluster(resourceId, client)
	}
}
