package s3

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	pio "github.com/confluentinc/cli/internal/pkg/io"
)

type PublicRepo struct {
	*PublicRepoParams
	// @VisibleForTesting
	endpoint string
	fs       pio.FileSystem
	goos     string
	goarch   string
}

type PublicRepoParams struct {
	S3BinBucket string
	S3BinRegion string
	S3BinPrefix string
	S3ObjectKey ObjectKey
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
		fs:               &pio.RealFileSystem{},
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received unexpected response from S3: %s", resp.Status)
	}

	var result ListBucketResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	var availableVersions version.Collection
	for _, v := range result.Contents {
		match, foundVersion, err := r.S3ObjectKey.ParseVersion(v.Key, name)
		if err != nil {
			return nil, err
		}
		if !match {
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
	s3URL := r.S3ObjectKey.URLFor(name, version)
	downloadVersion := fmt.Sprintf("%s/%s", r.endpoint, s3URL)

	resp, err := http.Get(downloadVersion)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			r.Logger.Tracef("Response from AWS: %s", string(body))
		}
		return "", 0, errors.Errorf("received unexpected response from S3: %s", resp.Status)
	}

	binName := fmt.Sprintf("%s-v%s-%s-%s", name, version, r.goos, r.goarch)
	downloadBinPath := filepath.Join(downloadDir, binName)

	downloadBin, err := r.fs.Create(downloadBinPath)
	if err != nil {
		return "", 0, err
	}
	defer downloadBin.Close()

	bytes, err := r.fs.Copy(downloadBin, resp.Body)
	if err != nil {
		return "", 0, err
	}

	return downloadBinPath, bytes, nil
}
