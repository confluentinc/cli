package s3

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
)

var S3ReleaseNotesFile = "release-notes.rst"

type PublicRepo struct {
	*PublicRepoParams
	// @VisibleForTesting
	endpoint string
	fs       pio.FileSystem
	goos     string
	goarch   string
}

type PublicRepoParams struct {
	S3BinBucket             string
	S3BinRegion             string
	S3BinPrefixFmt          string
	S3ReleaseNotesPrefixFmt string
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
		goos:             update.GetOs(),
		goarch:           runtime.GOARCH,
	}
}

func (r *PublicRepo) GetLatestMajorAndMinorVersion(name string, current *version.Version) (*version.Version, *version.Version, error) {
	versions, err := r.GetAvailableBinaryVersions(name)
	if err != nil {
		return nil, nil, errors.Wrapf(err, errors.GetBinaryVersionsErrorMsg)
	}

	// The index of the largest available version. This may be a major version update.
	majorIdx := len(versions) - 1
	if majorIdx < 0 {
		return nil, nil, errors.New(errors.GetBinaryVersionsErrorMsg)
	}
	major := versions[majorIdx]
	if current.Segments()[0] == major.Segments()[0] {
		major = nil
	}

	// The index of the largest available minor version. This will not be a major version update.
	nextMajorNum := current.Segments()[0] + 1
	nextMajorVer, _ := version.NewVersion(fmt.Sprintf("%d.0.0", nextMajorNum))

	// Find the first major version update and go back one. If there is no major version update, this will simply be the last index.
	minorIdx := sort.Search(len(versions), func(idx int) bool {
		return versions[idx].GreaterThanOrEqual(nextMajorVer)
	}) - 1

	if minorIdx < 0 {
		return nil, nil, errors.New(errors.GetBinaryVersionsErrorMsg)
	}
	minor := versions[minorIdx]

	return major, minor, nil
}

func (r *PublicRepo) GetAvailableBinaryVersions(name string) (version.Collection, error) {
	listBucketResult, err := r.getListBucketResultFromDir(fmt.Sprintf(r.S3BinPrefixFmt, name))
	if err != nil {
		return nil, err
	}
	availableVersions, err := r.getMatchedBinaryVersionsFromListBucketResult(listBucketResult, name)
	if err != nil {
		return nil, err
	}
	if len(availableVersions) == 0 {
		return nil, errors.New(errors.NoVersionsErrorMsg)
	}
	return availableVersions, nil
}

func (r *PublicRepo) getListBucketResultFromDir(s3DirPrefix string) (*ListBucketResult, error) {
	url := fmt.Sprintf("%s?prefix=%s/", r.endpoint, s3DirPrefix)
	log.CliLogger.Debugf("Getting available versions from %s", url)

	var results []ListBucketResult
	more := true

	for more {
		resp, err := r.getHttpResponse(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var result ListBucketResult
		if err := xml.Unmarshal(body, &result); err != nil {
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
		result.Contents = append(result.Contents, r.Contents...)
	}

	return &result, nil
}

func (r *PublicRepo) getMatchedBinaryVersionsFromListBucketResult(result *ListBucketResult, name string) (version.Collection, error) {
	objectKey, _ := NewPrefixedKey(fmt.Sprintf(r.S3BinPrefixFmt, name), "_", true)
	objectKey.goos = r.goos
	objectKey.goarch = r.goarch

	var versions version.Collection
	for _, v := range result.Contents {
		match, foundVersion, err := objectKey.ParseVersion(v.Key, name)
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

func (r *PublicRepo) GetLatestReleaseNotesVersions(name, currentVersion string) (version.Collection, error) {
	versions, err := r.GetAvailableReleaseNotesVersions(name)
	if err != nil {
		return nil, errors.Wrapf(err, errors.GetReleaseNotesVersionsErrorMsg)
	}

	current, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil, err
	}

	idx := sort.Search(len(versions), func(i int) bool { return versions[i].GreaterThan(current) })

	return versions[idx:], nil
}

func (r *PublicRepo) GetAvailableReleaseNotesVersions(name string) (version.Collection, error) {
	listBucketResult, err := r.getListBucketResultFromDir(fmt.Sprintf(r.S3ReleaseNotesPrefixFmt, name))
	if err != nil {
		return nil, err
	}
	availableVersions := r.getMatchedReleaseNotesVersionsFromListBucketResult(name, listBucketResult)
	if len(availableVersions) == 0 {
		return nil, errors.New(errors.NoVersionsErrorMsg)
	}
	return availableVersions, nil
}

func (r *PublicRepo) getMatchedReleaseNotesVersionsFromListBucketResult(name string, result *ListBucketResult) version.Collection {
	var versions version.Collection
	for _, v := range result.Contents {
		match, foundVersion := r.parseMatchedReleaseNotesVersion(name, v.Key)
		if match {
			versions = append(versions, foundVersion)
		}
	}
	sort.Sort(versions)
	return versions
}

func (r *PublicRepo) parseMatchedReleaseNotesVersion(name, key string) (bool, *version.Version) {
	if !strings.HasPrefix(key, fmt.Sprintf(r.S3ReleaseNotesPrefixFmt, name)) {
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

func (r *PublicRepo) DownloadVersion(name, version, downloadDir string) ([]byte, error) {
	objectKey, _ := NewPrefixedKey(fmt.Sprintf(r.S3BinPrefixFmt, name), "_", true)
	objectKey.goos = r.goos
	objectKey.goarch = r.goarch

	s3URL := objectKey.URLFor(name, version)
	downloadVersion := r.getDownloadVersion(s3URL)

	resp, err := r.getHttpResponse(downloadVersion)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (r *PublicRepo) DownloadReleaseNotes(name, version string) (string, error) {
	downloadURL := fmt.Sprintf("%s/%s/%s/%s", r.endpoint, fmt.Sprintf(r.S3ReleaseNotesPrefixFmt, name), version, S3ReleaseNotesFile)
	resp, err := r.getHttpResponse(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (r *PublicRepo) DownloadChecksums(name, version string) (string, error) {
	cliBinDir := fmt.Sprintf(r.S3BinPrefixFmt, name)
	checksumFileName := fmt.Sprintf("%s_%s_checksums.txt", name, version)
	downloadURL := fmt.Sprintf("%s/%s/%s/%s", r.endpoint, cliBinDir, version, checksumFileName)
	resp, err := r.getHttpResponse(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.CliLogger.Tracef("Downloaded the following checksums for version %s:\n%s", version, string(body))

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
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			log.CliLogger.Tracef("Response from AWS: %s", string(body))
		}
		return nil, errors.Errorf(errors.UnexpectedS3ResponseErrorMsg, resp.Status)
	}
	return resp, nil
}

func (r *PublicRepo) getDownloadVersion(s3URL string) string {
	return fmt.Sprintf("%s/%s", r.endpoint, s3URL)
}
