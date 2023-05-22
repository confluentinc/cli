package hub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type Manifest struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Version  string    `json:"version"`
	Owner    Owner     `json:"owner"`
	Archive  Archive   `json:"archive"`
	Licenses []License `json:"license"`
}

type Owner struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Archive struct {
	Url  string `json:"url"`
	Md5  string `json:"md5"`
	Sha1 string `json:"sha1"`
}

type License struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (c *Client) GetRemoteManifest(owner, name, version string) (*Manifest, error) {
	manifestUrl := fmt.Sprintf("%s/api/plugins/%s/%s", c.URL, owner, name)
	if version != "latest" {
		manifestUrl = fmt.Sprintf("%s/versions/%s", manifestUrl, version)
	}

	req, err := http.NewRequest(http.MethodGet, manifestUrl, nil)
	if err != nil {
		return nil, err
	}

	r, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
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

	pluginManifest := new(Manifest)
	if err := json.Unmarshal(body, &pluginManifest); err != nil {
		return nil, err
	}

	return pluginManifest, nil
}

func (c *Client) GetRemoteArchive(pluginManifest *Manifest) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, pluginManifest.Archive.Url, nil)
	if err != nil {
		return nil, err
	}

	r, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve archive from Confuent Hub")
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
