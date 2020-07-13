//go:generate sh -c "go run github.com/travisjeffery/mocker/cmd/mocker --prefix \"\" --dst ../mock/s3api.go --pkg mock \"$(go list -f '{{ .Dir }}' -m github.com/aws/aws-sdk-go)/service/s3/s3iface/interface.go\" S3API"
//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/Downloader.go --pkg mock private.go Downloader
package s3

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"

	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Downloader is the iface for s3manager.Downloader for testing DownloadVersion without actually connecting to s3
type Downloader interface {
	Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error)
}

// credsFactory for testing getCredentials without actually changing your ~/.aws/credentials file
type credsFactory interface {
	newProvider(profile string) credentials.Provider
}

type PrivateRepoParams struct {
	S3BinBucket string
	S3BinRegion string
	S3BinPrefix string
	S3ObjectKey ObjectKey
	AWSProfiles []string
	Logger      *log.Logger
	// @VisibleForTesting
	creds        credsFactory
	s3svc        s3iface.S3API
	s3downloader Downloader
	fs           pio.FileSystem
}

type PrivateRepo struct {
	*PrivateRepoParams
	// @VisibleForTesting
	goos   string
	goarch string
}

func NewPrivateRepo(params *PrivateRepoParams) (*PrivateRepo, error) {
	if err := validate(params); err != nil {
		return nil, err
	}

	if params.creds == nil {
		params.creds = &sharedCredsFactory{}
	}
	creds, err := getCredentials(params.creds, params.AWSProfiles)
	if err != nil {
		return nil, err
	}

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(params.S3BinRegion),
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}
	if params.s3svc == nil {
		params.s3svc = s3.New(s)
	}
	if params.s3downloader == nil {
		params.s3downloader = s3manager.NewDownloader(s)
	}
	return &PrivateRepo{
		PrivateRepoParams: params,
	}, nil
}

func validate(params *PrivateRepoParams) error {
	var err *multierror.Error
	if params.S3BinRegion == "" {
		err = multierror.Append(err, errors.Errorf(errors.MissingRequiredParamErrorMsg, "S3BinRegion"))
	}
	if params.S3BinBucket == "" {
		err = multierror.Append(err, errors.Errorf(errors.MissingRequiredParamErrorMsg, "S3BinBucket"))
	}
	if params.S3BinPrefix == "" {
		err = multierror.Append(err, errors.Errorf(errors.MissingRequiredParamErrorMsg, "S3BinPrefix"))
	}
	return err.ErrorOrNil()
}

func (r *PrivateRepo) GetAvailableVersions(name string) (version.Collection, error) {
	result, err := r.s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(r.S3BinBucket),
		Prefix: aws.String(r.S3BinPrefix + "/"),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.ListingS3BucketErrorMsg)
	}

	var availableVersions version.Collection
	for _, c := range result.Contents {
		match, foundVersion, err := r.S3ObjectKey.ParseVersion(*c.Key, name)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		availableVersions = append(availableVersions, foundVersion)
	}

	if len(availableVersions) <= 0 {
		return nil, errors.New(errors.NoVersionsErrorMsg)
	}

	sort.Sort(availableVersions)

	return availableVersions, nil
}

func (r *PrivateRepo) DownloadVersion(name, version, downloadDir string) (string, int64, error) {
	binName := fmt.Sprintf("%s-v%s-%s-%s", name, version, r.goos, r.goarch)
	downloadBinPath := filepath.Join(downloadDir, binName)
	downloadBin, err := r.fs.Create(downloadBinPath)
	if err != nil {
		return "", 0, err
	}
	defer downloadBin.Close()

	s3URL := r.S3ObjectKey.URLFor(name, version)
	bytes, err := r.s3downloader.Download(downloadBin, &s3.GetObjectInput{
		Bucket: aws.String(r.S3BinBucket),
		Key:    aws.String(s3URL),
	})
	if err != nil {
		return "", 0, err
	}

	return downloadBinPath, bytes, nil
}

func getCredentials(cf credsFactory, allProfiles []string) (*credentials.Credentials, error) {
	envProfile := os.Getenv("AWS_PROFILE")
	if envProfile != "" {
		allProfiles = append(allProfiles, envProfile)
	}
	if len(allProfiles) == 0 {
		allProfiles = append(allProfiles, "default")
	}

	var creds *credentials.Credentials
	var allErrors *multierror.Error
	for _, profile := range allProfiles {
		profileCreds := credentials.NewCredentials(cf.newProvider(profile))
		val, err := profileCreds.Get()
		if err != nil {
			allErrors = multierror.Append(allErrors, errors.Wrap(err, errors.FindingCredsErrorMsg))
			continue
		}

		if val.AccessKeyID == "" {
			allErrors = multierror.Append(allErrors, errors.Errorf(errors.EmptyAccessKeyIDErrorMsg, profile))
			continue
		}

		if profileCreds.IsExpired() {
			allErrors = multierror.Append(allErrors, errors.Errorf(errors.AWSCredsExpiredErrorMsg, profile))
			continue
		}

		creds = profileCreds
		break
	}

	if creds == nil {
		return nil, formatError(allProfiles, allErrors)
	}
	return creds, nil
}

func formatError(profiles []string, origErrors error) error {
	var newErrors *multierror.Error
	if e, ok := (origErrors).(*multierror.Error); ok {
		newErrors = multierror.Append(newErrors, errors.Errorf(errors.FindAWSCredsErrorMsg,
			strings.Join(profiles, ", ")),
		)
		for _, errMsg := range e.Errors {
			/*
				aws error puts a newline into the message; idk why but it looks
				ugly so remove it

				2019/01/17 09:25:40 failed to find aws credentials in profiles: confluent-dev, confluent, default
				2019/01/17 09:25:40   error while finding creds: SharedCredsLoad: failed to get profile
				caused by: section 'confluent-dev' does not exist
				2019/01/17 09:25:40   error while finding creds: SharedCredsLoad: failed to get profile
				caused by: section 'confluent' does not exist
				2019/01/17 09:25:40   error while finding creds: SharedCredsLoad: failed to get profile
				caused by: section 'default' does not exist
				2019/01/17 09:25:40 Checking for updates...

				vs

				2019/01/17 09:27:12 failed to find aws credentials in profiles: confluent-dev, confluent, default
				2019/01/17 09:27:12   error while finding creds: SharedCredsLoad: failed to get profile caused by: section 'confluent-dev' does not exist
				2019/01/17 09:27:12   error while finding creds: SharedCredsLoad: failed to get profile caused by: section 'confluent' does not exist
				2019/01/17 09:27:12   error while finding creds: SharedCredsLoad: failed to get profile caused by: section 'default' does not exist
				2019/01/17 09:27:12 Checking for updates...
			*/
			newErrors = multierror.Append(newErrors, errors.Errorf("  %s", strings.ReplaceAll(errMsg.Error(), "\n", " ")))
		}
	}
	return newErrors.ErrorOrNil()
}

type sharedCredsFactory struct{}

func (f *sharedCredsFactory) newProvider(profile string) credentials.Provider {
	// credentials.NewSharedProvider does this internally...
	return &credentials.SharedCredentialsProvider{
		Filename: "",
		Profile:  profile,
	}
}
