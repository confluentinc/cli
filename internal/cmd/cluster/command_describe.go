package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type Metadata interface {
	DescribeCluster(ctx context.Context, url string) (*ScopedId, error)
}

type ScopedId struct {
	ID    string `json:"id"`
	Scope *Scope `json:"scope"`
}

type Scope struct {
	// Path defines the "outer scope" which isn't used yet. The hierarchy
	// isn't represented in the Scope object in practice today
	Path []string `json:"path"`
	// Clusters defines all the key-value pairs needed to uniquely identify a scope
	Clusters map[string]string `json:"clusters"`
}

type Element struct {
	Type string `json:"type" yaml:"type"`
	ID   string `json:"id" yaml:"id"`
}

// ScopedIdService allows introspecting details from a Confluent cluster.
// This is for querying the endpoint each CP service exposes at /v1/metadata/id.
type ScopedIdService struct {
	client    *http.Client
	userAgent string
	logger    *log.Logger
}

func NewScopedIdService(client *http.Client, userAgent string, logger *log.Logger) *ScopedIdService {
	return &ScopedIdService{
		client:    client,
		userAgent: userAgent,
		logger:    logger,
	}
}

var (
	describeFields = []string{"Type", "ID"}
	describeLabels = []string{"Type", "ID"}
)

type describeCommand struct {
	*pcmd.CLICommand
	client Metadata
}

// NewDescribeCommand returns the sub-command object for describing clusters through /v1/metadata/id
func NewDescribeCommand(prerunner pcmd.PreRunner, client Metadata) *cobra.Command {
	describeCmd := &describeCommand{
		CLICommand: pcmd.NewAnonymousCLICommand(&cobra.Command{
			Use:   "describe",
			Short: "Describe a Kafka cluster.",
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Discover the cluster ID and Kafka ID for Connect.",
					Code: "confluent cluster describe --url http://localhost:8083",
				},
			),
		}, prerunner),
		client: client,
	}
	describeCmd.Flags().String("url", "", "URL to a Confluent cluster.")
	check(describeCmd.MarkFlagRequired("url"))
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	describeCmd.RunE = pcmd.NewCLIRunE(describeCmd.describe)

	return describeCmd.Command
}

func (s *ScopedIdService) DescribeCluster(_ context.Context, url string) (*ScopedId, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/metadata/id", url), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(errors.FetchClusterMetadataErrorMsg, resp.Status, body)
	}
	meta := &ScopedId{}
	err = json.Unmarshal(body, meta)
	return meta, err
}

func (c *describeCommand) describe(cmd *cobra.Command, _ []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return nil
	}

	meta, err := c.client.DescribeCluster(context.Background(), url)
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	return printDescribe(cmd, meta, outputOption)
}

func printDescribe(cmd *cobra.Command, meta *ScopedId, format string) error {
	type StructuredDisplay struct {
		Crn   string    `json:"crn" yaml:"crn"`
		Scope []Element `json:"scope" yaml:"scope"`
	}
	structuredDisplay := &StructuredDisplay{}
	if meta.ID != "" {
		if format == output.Human.String() {
			pcmd.Printf(cmd, "Confluent Resource Name: %s\n\n", meta.ID)
		} else {
			structuredDisplay.Crn = meta.ID
		}
	}
	var types []string
	for name := range meta.Scope.Clusters {
		types = append(types, name)
	}
	sort.Strings(types) // since we don't have hierarchy info, just display in alphabetical order
	var data [][]string
	for _, name := range types {
		id := meta.Scope.Clusters[name]
		element := Element{Type: name, ID: id}
		if format == output.Human.String() {
			data = append(data, printer.ToRow(&element, describeFields))
		} else {
			structuredDisplay.Scope = append(structuredDisplay.Scope, element)
		}

	}
	if format == output.Human.String() {
		pcmd.Println(cmd, "Scope:")
		printer.RenderCollectionTable(data, describeLabels)
	} else {
		return output.StructuredOutput(format, structuredDisplay)
	}
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
