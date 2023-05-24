package hub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/confluentinc/cli/internal/pkg/cpstructs"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func (c *Client) GetRemoteManifest(owner, name, version string) (*cpstructs.Manifest, error) {
	manifestUrl := fmt.Sprintf("%s/api/plugins/%s/%s", c.URL, owner, name)
	if version != "latest" {
		manifestUrl = fmt.Sprintf("%s/versions/%s", manifestUrl, version)
	}

	req, err := http.NewRequest(http.MethodGet, manifestUrl, nil)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Tracef("\n%s\n", string(dump))
	}

	r, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if c.Debug {
		dump, err := httputil.DumpResponse(r, true)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Tracef("\n%s\n", string(dump))
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		response := make(map[string]interface{})
		_ = json.Unmarshal(body, &response)
		if errorMessage, ok := response["message"]; ok {
			return nil, errors.Errorf("failed to read manifest file from Confluent Hub: %s", errorMessage)
		}
		return nil, errors.Errorf("failed to read manifest file from Confluent Hub")
	}

	pluginManifest := new(cpstructs.Manifest)
	if err := json.Unmarshal(body, &pluginManifest); err != nil {
		return nil, err
	}

	return pluginManifest, nil
}

func (c *Client) GetRemoteArchive(pluginManifest *cpstructs.Manifest) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, pluginManifest.Archive.Url, nil)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Tracef("\n%s\n", string(dump))
	}

	r, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve archive from Confuent Hub")
	}

	return io.ReadAll(r.Body)
}
