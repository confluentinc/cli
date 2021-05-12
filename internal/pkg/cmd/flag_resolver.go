package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"

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
	ResolveApiKeyFlag(cmd *cobra.Command) (string, error)
	ResolveApiKeySecretFlag(cmd *cobra.Command) (string, error)
	ResolveOnPremKafkaRestFlags(cmd *cobra.Command) (*OnPremKafkaRestFlagValues, error)
	//ResolveCaCertPathFlag(cmd *cobra.Command) (string, error)
	//ResolveNoAuthFlag(cmd *cobra.Command) (bool, error)
	//ResolveClientCertAndKeyPathsFlag(cmd *cobra.Command) (string, string, error)
}

type FlagResolverImpl struct {
	Prompt form.Prompt
	Out    io.Writer
}

type OnPremKafkaRestFlagValues struct {
	url            string
	caCertPath     string
	clientCertPath string
	clientKeyPath  string
	noAuth         bool
	prompt         bool
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

func (r *FlagResolverImpl) ResolveApiKeyFlag(cmd *cobra.Command) (string, error) {
	const keyFlag = "api-key"
	if cmd.Flags().Changed(keyFlag) {
		key, err := cmd.Flags().GetString(keyFlag)
		if err != nil {
			return "", err
		}
		return key, nil
	}
	return "", nil
}

func (r *FlagResolverImpl) ResolveApiKeySecretFlag(cmd *cobra.Command) (string, error) {
	const secretFlag = "api-secret"
	if cmd.Flags().Changed(secretFlag) {
		secret, err := cmd.Flags().GetString(secretFlag)
		if err != nil {
			return "", err
		}
		return secret, nil
	}
	return "", nil
}

func (r *FlagResolverImpl) ResolveOnPremKafkaRestFlags(cmd *cobra.Command) (*OnPremKafkaRestFlagValues, error) {
	flagValues := new(OnPremKafkaRestFlagValues)

	url, err := r.resolveURLFlag(cmd)
	if err != nil {
		return flagValues, err
	}
	flagValues.url = url

	caCertPath, err := r.resolveCaCertPathFlag(cmd)
	if err != nil {
		return flagValues, err
	}
	flagValues.caCertPath = caCertPath

	clientCertPath, clientKeyPath, err := r.resolveClientCertAndKeyPathsFlag(cmd)
	if err != nil {
		return flagValues, err
	}
	flagValues.clientCertPath = clientCertPath
	flagValues.clientKeyPath = clientKeyPath

	noAuth, err := r.resolveNoAuthFlag(cmd)
	if err != nil {
		return flagValues, err
	}
	flagValues.noAuth = noAuth

	prompt, err := r.resolvePromptFlag(cmd)
	if err != nil {
		return flagValues, err
	}
	flagValues.prompt = prompt

	return flagValues, nil
}

func (r *FlagResolverImpl) resolveURLFlag(cmd *cobra.Command) (string, error) {
	const url = "url"
	if cmd.Flags().Changed(url) {
		url, err := cmd.Flags().GetString(url)
		if err != nil {
			return "", err
		}
		return url, nil
	}
	return "", nil
}

func (r *FlagResolverImpl) resolveCaCertPathFlag(cmd *cobra.Command) (string, error) {
	const certPathFlag = "ca-cert-path"
	if cmd.Flags().Changed(certPathFlag) {
		path, err := cmd.Flags().GetString(certPathFlag)
		if err != nil {
			return "", err
		}
		return path, nil
	}
	return "", nil
}

func (r *FlagResolverImpl) resolveClientCertAndKeyPathsFlag(cmd *cobra.Command) (string, string, error) {
	const clientCertPathFlag = "client-cert-path"
	var clientCertPath string
	var err error
	if cmd.Flags().Changed(clientCertPathFlag) {
		clientCertPath, err = cmd.Flags().GetString(clientCertPathFlag)
		if err != nil {
			return "", "", err
		}
	}
	const clientKeyPathFlag = "client-key-path"
	var clientKeyPath string
	if cmd.Flags().Changed(clientKeyPathFlag) {
		clientKeyPath, err = cmd.Flags().GetString(clientKeyPathFlag)
		if err != nil {
			return "", "", err
		}
	}
	if (clientCertPath != "" && clientKeyPath == "") || (clientCertPath == "" && clientKeyPath != "") {
		return "", "", errors.New(errors.NeedClientCertAndKeyPathsErrorMsg)
	}
	return clientCertPath, clientKeyPath, nil
}

func (r *FlagResolverImpl) resolveNoAuthFlag(cmd *cobra.Command) (bool, error) {
	const noAuthFlag = "no-auth"
	if cmd.Flags().Changed(noAuthFlag) {
		noAuth, err := cmd.Flags().GetBool(noAuthFlag)
		if err != nil {
			return false, err
		}
		return noAuth, nil
	}
	return false, nil
}

func (r *FlagResolverImpl) resolvePromptFlag(cmd *cobra.Command) (bool, error) {
	const prompt = "prompt"
	if cmd.Flags().Changed(prompt) {
		prompt, err := cmd.Flags().GetBool(prompt)
		if err != nil {
			return false, err
		}
		return prompt, nil
	}
	return false, nil
}
