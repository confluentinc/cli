package update

import (
	"fmt"
	"io"
	"net/http"

	"github.com/confluentinc/cli/v4/pkg/httpclient"
	"github.com/confluentinc/cli/v4/pkg/log"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

type Client struct {
	url    string
	client *http.Client
}

func NewClient(isTest bool, cfg httpclient.Config) *Client {
	url := "https://packages.confluent.io"
	if isTest {
		url = testserver.TestPackagesUrl.String()
	}

	return &Client{url: url, client: httpclient.New(cfg)}
}

func (c *Client) GetBinaries() (string, error) {
	return c.get(fmt.Sprintf("%s/confluent-cli/binaries/", c.url))
}

func (c *Client) GetBinary(version, filename string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/confluent-cli/binaries/%s/%s", c.url, version, filename)
	log.CliLogger.Tracef("GET %s", url)

	resp, err := httpclient.New(httpclient.LargeDownload).Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetBinaryChecksums(version string) (string, error) {
	return c.get(fmt.Sprintf("%s/confluent-cli/binaries/%s/confluent_checksums.txt", c.url, version))
}

func (c *Client) GetReleaseNotes(version string) (string, error) {
	return c.get(fmt.Sprintf("%s/confluent-cli/release-notes/%s/release-notes.rst", c.url, version))
}

func (c *Client) get(url string) (string, error) {
	log.CliLogger.Tracef("GET %s", url)

	resp, err := c.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
