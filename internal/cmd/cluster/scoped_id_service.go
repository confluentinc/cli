package cluster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

// ScopedIdService allows introspecting details from a Confluent cluster.
// This is for querying the endpoint each CP service exposes at /v1/metadata/id.
type ScopedIdService struct {
	userAgent string
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

func newScopedIdService(userAgent string) *ScopedIdService {
	return &ScopedIdService{userAgent: userAgent}
}

func (s *ScopedIdService) DescribeCluster(url, caCertPath string) (*ScopedId, error) {
	var httpClient *http.Client
	if caCertPath != "" {
		var err error
		httpClient, err = utils.SelfSignedCertClientFromPath(caCertPath)
		if err != nil {
			return nil, err
		}
	} else {
		httpClient = utils.DefaultClient()
	}

	ctx := utils.GetContext()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/metadata/id", url), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
