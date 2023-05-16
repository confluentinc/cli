package connect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func getRemoteManifest(owner, name, version string) (*manifest, error) {
	manifestUrl := fmt.Sprintf("https://api.hub.confluent.io/api/plugins/%s/%s", owner, name)
	if version != "latest" {
		manifestUrl = fmt.Sprintf("%s/versions/%s", manifestUrl, version)
	}

	r, err := http.Get(manifestUrl)
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

	pluginManifest := new(manifest)
	if err := json.Unmarshal(body, &pluginManifest); err != nil {
		return nil, err
	}

	return pluginManifest, nil
}

func getRemoteArchive(pluginManifest *manifest) ([]byte, error) {
	r, err := http.Get(pluginManifest.Archive.Url)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve archive from Confuent Hub")
	}
	defer r.Body.Close()

	return io.ReadAll(r.Body)
}
