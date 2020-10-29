package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/form"
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
	Prompt form.Prompt
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
			value, err = r.Prompt.ReadLineMasked()
		} else {
			value, err = r.Prompt.ReadLine()
		}
		if err != nil {
			return "", err
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
		value, err = r.Prompt.ReadLine()
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
	KSQLResourceType  = "ksql"
	CloudResourceType = "cloud"
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
	if strings.HasPrefix(resourceId, "lsrc-") {
		// Resource is schema registry.
		resourceType = SrResourceType
	} else if strings.HasPrefix(resourceId, "lksqlc-") {
		resourceType = KSQLResourceType
	} else if resourceId == CloudResourceType {
		resourceType = CloudResourceType
		resourceId = ""
	} else {
		// Resource is Kafka cluster.
		resourceType = KafkaResourceType
	}
	return resourceType, resourceId, nil
}
