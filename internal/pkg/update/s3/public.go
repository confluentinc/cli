package s3

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
)

var (
	S3ReleaseNotesFile = "release-notes.rst"
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
	S3BinBucket          string
	S3BinRegion          string
	S3BinPrefix          string
	S3ReleaseNotesPrefix string
	S3ObjectKey          ObjectKey
	Logger               *log.Logger
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

func (r *PublicRepo) GetLatestBinaryVersion(name string) (*version.Version, error) {
	availableVersions, err := r.GetAvailableBinaryVersions(name)
	if err != nil {
		return nil, errors.Wrapf(err, errors.GetBinaryVersionsErrorMsg)
	}
	return availableVersions[len(availableVersions)-1], nil
}

func (r *PublicRepo) GetAvailableBinaryVersions(name string) (version.Collection, error) {
	listBucketResult, err := r.getListBucketResultFromDir(r.S3BinPrefix)
	if err != nil {
		return nil, err
	}
	availableVersions, err := r.getMatchedBinaryVersionsFromListBucketResult(listBucketResult, name)
	if err != nil {
		return nil, err
	}
	if len(availableVersions) <= 0 {
		return nil, errors.New(errors.NoVersionsErrorMsg)
	}
	return availableVersions, nil
}

func (r *PublicRepo) getListBucketResultFromDir(s3DirPrefix string) (*ListBucketResult, error) {
	url := fmt.Sprintf("%s?prefix=%s/", r.endpoint, s3DirPrefix)
	r.Logger.Debugf("Getting available versions from %s", url)

	var results []ListBucketResult
	more := true

	for more {
		resp, err := r.getHttpResponse(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var result ListBucketResult
		err = xml.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		// ListBucketResult paginates results
		if result.IsTruncated {
			// Last key is the "marker" used as the starting point for next page of results
			marker := result.Contents[len(result.Contents)-1].Key
			url = fmt.Sprintf("%s?prefix=%s/&marker=%s", r.endpoint, s3DirPrefix, marker)
		} else {
			more = false
		}
	}

	// Concatenate paginated results here so rest of code doesn't have to think about pagination
	result := results[0] // copy most properties from results[0]
	result.IsTruncated = false
	for _, r := range results[1:] { // skip results[0]
		result.Contents = append(result.Contents, r.Contents[1:]...) // don't duplicate "marker" entry
	}

	return &result, nil
}

func (r *PublicRepo) getMatchedBinaryVersionsFromListBucketResult(result *ListBucketResult, name string) (version.Collection, error) {
	var versions version.Collection
	for _, v := range result.Contents {
		match, foundVersion, err := r.S3ObjectKey.ParseVersion(v.Key, name)
		if err != nil {
			return nil, err
		}
		if match {
			versions = append(versions, foundVersion)
		}
	}
	sort.Sort(versions)
	return versions, nil
}

func (r *PublicRepo) GetLatestReleaseNotesVersion() (*version.Version, error) {
	availableVersions, err := r.GetAvailableReleaseNotesVersions()
	if err != nil {
		return nil, errors.Wrapf(err, errors.GetReleaseNotesVersionsErrorMsg)
	}
	return availableVersions[len(availableVersions)-1], nil
}

func (r *PublicRepo) GetAvailableReleaseNotesVersions() (version.Collection, error) {
	listBucketResult, err := r.getListBucketResultFromDir(r.S3ReleaseNotesPrefix)
	if err != nil {
		return nil, err
	}
	availableVersions, err := r.getMatchedReleaseNotesVersionsFromListBucketResult(listBucketResult)
	if err != nil {
		return nil, err
	}
	if len(availableVersions) <= 0 {
		return nil, errors.New(errors.NoVersionsErrorMsg)
	}
	return availableVersions, nil
}

func (r *PublicRepo) getMatchedReleaseNotesVersionsFromListBucketResult(result *ListBucketResult) (version.Collection, error) {
	var versions version.Collection
	for _, v := range result.Contents {
		match, foundVersion := r.parseMatchedReleaseNotesVersion(v.Key)
		if match {
			versions = append(versions, foundVersion)
		}
	}
	sort.Sort(versions)
	return versions, nil
}

func (r *PublicRepo) parseMatchedReleaseNotesVersion(key string) (match bool, ver *version.Version) {
	if !strings.HasPrefix(key, r.S3ReleaseNotesPrefix) {
		return false, nil
	}
	split := strings.Split(key, "/")
	if split[len(split)-1] != S3ReleaseNotesFile {
		return false, nil
	}
	ver, err := version.NewSemver(split[2])
	if err != nil {
		return false, nil
	}
	return true, ver
}

func (r *PublicRepo) DownloadVersion(name, version, downloadDir string) (string, int64, error) {
	s3URL := r.S3ObjectKey.URLFor(name, version)
	downloadVersion := fmt.Sprintf("%s/%s", r.endpoint, s3URL)

	resp, err := r.getHttpResponse(downloadVersion)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

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

func (r *PublicRepo) DownloadReleaseNotes(version string) (string, error) {
	downloadURL := fmt.Sprintf("%s/%s/%s/%s", r.endpoint, r.S3ReleaseNotesPrefix, version, S3ReleaseNotesFile)
	resp, err := r.getHttpResponse(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// must close the response afterwards
func (r *PublicRepo) getHttpResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			r.Logger.Tracef("Response from AWS: %s", string(body))
		}
		return nil, errors.Errorf(errors.UnexpectedS3ResponseErrorMsg, resp.Status)
	}
	return resp, nil
}
