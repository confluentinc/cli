package cluster

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type describeCommand struct {
	*pcmd.CLICommand
	client metadata
}

type metadata interface {
	DescribeCluster(url, caCertPath string) (*ScopedId, error)
}

type out struct {
	Crn   string     `json:"crn" yaml:"crn"`
	Scope []scopeOut `json:"scope" yaml:"scope"`
}

type scopeOut struct {
	Type string `human:"Type" json:"type" yaml:"type"`
	ID   string `human:"ID" json:"id" yaml:"id"`
}

func newDescribeCommand(prerunner pcmd.PreRunner, userAgent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a Kafka cluster.",
		Long:  fmt.Sprintf("Describe a Kafka cluster. Environment variable `%s` can replace the `--url` flag, and `%s` can replace the `--ca-cert-path` flag.", pauth.ConfluentPlatformMDSURL, pauth.ConfluentPlatformCACertPath),
		Args:  cobra.NoArgs,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Discover the cluster ID and Kafka ID for Connect.",
				Code: "confluent cluster describe --url http://localhost:8083",
			},
		),
	}

	c := &describeCommand{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		client:     newScopedIdService(userAgent),
	}

	c.RunE = c.describe

	c.Flags().String("url", "", "URL to a Confluent cluster.")
	c.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
	pcmd.AddOutputFlag(c.Command)

	return c.Command
}

func (c *describeCommand) describe(cmd *cobra.Command, _ []string) error {
	url, err := getURL(cmd)
	if err != nil {
		return err
	}

	caCertPath, err := getCACertPath(cmd)
	if err != nil {
		return err
	}

	meta, err := c.client.DescribeCluster(url, caCertPath)
	if err != nil {
		return err
	}

	return printDescribe(cmd, meta)
}

func getURL(cmd *cobra.Command) (string, error) {
	// Order of precedence: flags > env vars
	if url, err := cmd.Flags().GetString("url"); url != "" || err != nil {
		return url, err
	}

	if url := pauth.GetEnvWithFallback(pauth.ConfluentPlatformMDSURL, pauth.DeprecatedConfluentPlatformMDSURL); url != "" {
		return url, nil
	}

	return "", errors.New(errors.MdsUrlNotFoundSuggestions)
}

func getCACertPath(cmd *cobra.Command) (string, error) {
	// Order of precedence: flags > env vars
	if path, err := cmd.Flags().GetString("ca-cert-path"); path != "" || err != nil {
		return path, err
	}

	return pauth.GetEnvWithFallback(pauth.ConfluentPlatformCACertPath, pauth.DeprecatedConfluentPlatformCACertPath), nil
}

func printDescribe(cmd *cobra.Command, meta *ScopedId) error {
	var types []string
	for name := range meta.Scope.Clusters {
		types = append(types, name)
	}
	sort.Strings(types) // since we don't have hierarchy info, just display in alphabetical order

	if output.GetFormat(cmd).IsSerialized() {
		out := &out{
			Crn:   meta.ID,
			Scope: make([]scopeOut, len(types)),
		}
		for i, name := range types {
			out.Scope[i] = scopeOut{
				Type: name,
				ID:   meta.Scope.Clusters[name],
			}
		}
		return output.SerializedOutput(cmd, out)
	}

	if meta.ID != "" {
		utils.Printf(cmd, "Confluent Resource Name: %s\n\n", meta.ID)
	}

	utils.Println(cmd, "Scope:")
	list := output.NewList(cmd)
	for _, name := range types {
		list.Add(&scopeOut{
			Type: name,
			ID:   meta.Scope.Clusters[name],
		})
	}
	return list.Print()
}
