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
	S3KeyParser KeyParser
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
	resp, err := http.Get(listVersions)
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
		found, foundVersion, err := r.S3KeyParser.Validate(v.Key)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		availableVersions = append(availableVersions, foundVersion)
	}

	if len(availableVersions) <= 0 {
		return nil, fmt.Errorf("no versions found, that's pretty weird")
	}

	sort.Sort(availableVersions)

	return availableVersions, nil
}

func (r *PublicRepo) DownloadVersion(name, version, downloadDir string) (string, int64, error) {
	binName := fmt.Sprintf("%s-v%s-%s-%s", name, version, r.goos, r.goarch)
	downloadVersion := fmt.Sprintf("%s/%s/%s/%s", r.endpoint, r.S3BinPrefix, version, binName)

	resp, err := http.Get(downloadVersion)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

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
