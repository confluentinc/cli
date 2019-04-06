package s3

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type PublicRepo struct {
	*PublicRepoParams
	// @VisibleForTesting
	endpoint string
	goos     string
	goarch   string
}

type PublicRepoParams struct {
	S3BinBucket string
	S3BinRegion string
	S3BinPrefix string
	Logger      *log.Logger
}

type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	Prefix         string         `xml:"Prefix"`
	MaxKeys        int32          `xml:"MaxKeys"`
	Delimiter      string         `xml:"Delimiter"`
	IsTruncated    bool           `xml:"IsTruncated"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
	Contents       []Object
}

type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

type Object struct {
	Key string `xml:"Key"`
}

func NewPublicRepo(params *PublicRepoParams) *PublicRepo {
	return &PublicRepo{
		PublicRepoParams: params,
		endpoint:         fmt.Sprintf("https://s3-%s.amazonaws.com/%s", params.S3BinRegion, params.S3BinBucket),
		goos:             runtime.GOOS,
		goarch:           runtime.GOARCH,
	}
}

func (r *PublicRepo) GetAvailableVersions(name string) (version.Collection, error) {
	listVersions := fmt.Sprintf("%s?prefix=%s/", r.endpoint, r.S3BinPrefix)
	r.Logger.Debugf("Getting available versions from %s", listVersions)
	resp, err := http.Get(r.endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r.Logger.Tracef("Response from AWS: %s", string(body))

	var result ListBucketResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	var availableVersions version.Collection
	for _, v := range result.Contents {
		// Format: S3BinPrefix/VERSION/NAME_VERSION_OS_ARCH
		split := strings.Split(v.Key, "_")

		// Skip files that don't match our naming standards for binaries
		if len(split) != 4 {
			continue
		}

		// Skip objects from other directories
		if !strings.HasPrefix(split[0], r.S3BinPrefix) {
			continue
		}

		// Skip binaries other than the requested one
		if !strings.HasSuffix(split[0], name) {
			continue
		}

		// Skip binaries not for this OS
		if split[2] != runtime.GOOS {
			continue
		}

		// Skip binaries not for this Arch
		if split[3] != runtime.GOARCH {
			continue
		}

		// Skip snapshot and dirty versions (which shouldn't be published, but accidents happen)
		if strings.Contains(split[1], "SNAPSHOT") {
			continue
		}
		if strings.Contains(split[1], "dirty") {
			continue
		}

		ver, err := version.NewVersion(split[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s version - %s", name, err)
		}
		availableVersions = append(availableVersions, ver)
	}

	if len(availableVersions) <= 0 {
		return nil, fmt.Errorf("no versions found, that's pretty weird")
	}

	sort.Sort(availableVersions)

	return availableVersions, nil
}

func (r *PublicRepo) DownloadVersion(name, version, downloadDir string) (string, int64, error) {
	downloadVersion := fmt.Sprintf("%s/%s/%s/%s_%s_%s_%s", r.endpoint, r.S3BinPrefix,
		version, name, version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(downloadVersion)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	binName := fmt.Sprintf("%s-v%s-%s-%s", name, version, runtime.GOOS, runtime.GOARCH)
	downloadBinPath := filepath.Join(downloadDir, binName)

	downloadBin, err := os.Create(downloadBinPath)
	if err != nil {
		return "", 0, err
	}
	defer downloadBin.Close()

	bytes, err := io.Copy(downloadBin, resp.Body)
	if err != nil {
		return "", 0, err
	}

	return downloadBinPath, bytes, nil
}
