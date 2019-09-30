package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Metadata service allows introspecting details from a Confluent cluster.
// This is different from MDS but I can't think of a better name. Please help! The endpoint is even /v1/metadata/id !
// This doesn't really belong in ccloud-sdk-go or mds-sdk-go. Where does it belong? :(
// Maybe the Scope object belongs in github.com/confluentinc/crn?
type Metadata interface {
	DescribeCluster(ctx context.Context, url string) (*ClusterMetadata, error)
}

type ClusterMetadata struct {
	ID    string `json:"id"`
	Scope *Scope `json:"scope"`
}

type Scope struct {
	// Path defines the scope hierarchy, where each Path element is a key in the Clusters map
	Path []string `json:"path"`
	// Clusters defines all the key-value pairs needed to uniquely identify a scope
	Clusters map[string]string `json:"clusters"`
}

type MetadataService struct {
	client    *http.Client
	userAgent string
	logger    *log.Logger
}

func NewMetadataService(client *http.Client, userAgent string, logger *log.Logger) *MetadataService {
	return &MetadataService{
		client:    client,
		userAgent: userAgent,
		logger:    logger,
	}
}

type command struct {
	*cobra.Command
	config *config.Config
	client Metadata
}

func (s *MetadataService) DescribeCluster(ctx context.Context, url string) (*ClusterMetadata, error) {
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
		return nil, fmt.Errorf("unable to fetch cluster metadata: %s - %s", resp.Status, body)
	}
	meta := &ClusterMetadata{}
	err = json.Unmarshal(body, meta)
	return meta, err
}

// New returns the Cobra command for `cluster`.
func New(prerunner pcmd.PreRunner, config *config.Config, client Metadata) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "cluster",
			Short:             "Retrieve metadata about Confluent clusters.",
			PersistentPreRunE: prerunner.Anonymous(),
		},
		config: config,
		client: client,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a Confluent cluster.",
		RunE:  c.describe,
		Args:  cobra.NoArgs,
	}
	describeCmd.Flags().String("url", "", "URL to a Confluent cluster.")
	check(describeCmd.MarkFlagRequired("url"))
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)
}

type Tuple struct {
	Key   string
	Value string
}

var (
	listFields = []string{"Key", "Value"}
	listLabels = []string{"Key", "Value"}
)

func (c *command) describe(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return nil
	}

	meta, err := c.client.DescribeCluster(context.Background(), url)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	fmt.Println(meta)

	var data [][]string
	for _, element := range meta.Scope.Path {
		id := meta.Scope.Clusters[element]
		data = append(data, printer.ToRow(&Tuple{Key: element, Value: id}, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)

	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
