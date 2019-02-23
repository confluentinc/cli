package update

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/atrox/homedir"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/mattn/go-isatty"

	"github.com/confluentinc/cli/log"
)

type Params struct {
	S3BinBucket     string
	S3BinRegion     string
	S3BinPrefix     string
	UpdateCheckFile string
	Credentials     *credentials.Credentials
	Logger          *log.Logger
}

type Client struct {
	*Params
	session *session.Session
	s3svc   *s3.S3
}

// NewUpdateClient returns a client for updating CLI binaries
func NewUpdateClient(params *Params) (*Client, error) {
	if err := validate(params); err != nil {
		return nil, err
	}
	s := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(params.S3BinRegion),
		Credentials: params.Credentials,
	}))
	return &Client{
		Params:  params,
		session: s,
		s3svc:   s3.New(s),
	}, nil
}

func validate(params *Params) error {
	var err *multierror.Error
	if params.S3BinRegion == "" {
		err = multierror.Append(err, fmt.Errorf("missing required parameter: S3BinRegion"))
	}
	if params.S3BinBucket == "" {
		err = multierror.Append(err, fmt.Errorf("missing required parameter: S3BinBucket"))
	}
	if params.S3BinPrefix == "" {
		err = multierror.Append(err, fmt.Errorf("missing required parameter: S3BinPrefix"))
	}
	return err.ErrorOrNil()
}

func GetCredentials(allProfiles []string) (*credentials.Credentials, error) {
	envProfile := os.Getenv("AWS_PROFILE")
	if envProfile != "" {
		allProfiles = append(allProfiles, envProfile)
	}

	var creds *credentials.Credentials
	var allErrors *multierror.Error
	for _, profile := range allProfiles {
		profileCreds := credentials.NewSharedCredentials("", profile)
		val, err := profileCreds.Get()
		if err != nil {
			allErrors = multierror.Append(allErrors, fmt.Errorf("error while finding creds: %s", err))
			continue
		}

		if val.AccessKeyID == "" {
			allErrors = multierror.Append(allErrors, fmt.Errorf("error: access key id is empty for %s", profile))
			continue
		}

		if profileCreds.IsExpired() {
			allErrors = multierror.Append(allErrors, fmt.Errorf("error: aws creds in profile %s are expired", profile))
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

// Check for new versions in s3 bucket
func (c *Client) CheckForUpdates(name string, currentVersion string) (updateAvailable bool, latestVersion string, err error) {
	availableVersions, err := c.GetAvailableVersions(name)
	if err != nil {
		return false, "", err
	}

	mostRecentVersion := availableVersions[len(availableVersions)-1]

	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, "", fmt.Errorf("unable to parse %s version %s - %s", name, currentVersion, err)
	}

	if currVersion.LessThan(mostRecentVersion) {
		return true, mostRecentVersion.String(), nil
	}

	return false, currentVersion, nil
}

func (c *Client) GetAvailableVersions(name string) (version.Collection, error) {
	result, err := c.s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(c.S3BinBucket),
		Prefix: aws.String(c.S3BinPrefix+"/"),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing s3 bucket %s", err)
	}

	var availableVersions version.Collection
	for _, r := range result.Contents {
		// Format: S3BinPrefix/NAME-v0.0.0-OS-ARCH
		split := strings.Split(*r.Key, "-")

		// Skip files that don't match our naming standards for binaries
		if len(split) != 4 {
			continue
		}

		// Skip non-matching binaries
		if split[0] != fmt.Sprintf("%s/%s", c.S3BinPrefix, name) {
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

		v, err := version.NewVersion(split[1])
		if err != nil {
			c.Logger.Warnf("WARNING: Unable to parse version %s - %s", split[1], err)
			continue
		}
		availableVersions = append(availableVersions, v)
	}

	if len(availableVersions) <= 0 {
		return nil, fmt.Errorf("no versions found, that's pretty weird")
	}

	sort.Sort(availableVersions)

	return availableVersions, nil
}

func (c *Client) PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool {
	if confirm && !isatty.IsTerminal(os.Stdout.Fd()) {
		c.Logger.Warn("WARN: disable confirm as stdout is not a tty")
		confirm = false
	}

	c.Logger.Printf("New version of %s is available", name)
	c.Logger.Printf("Current Version: %s", currVersion)
	c.Logger.Printf("Latest Version:  %s", latestVersion)

	if confirm {
		for {
			fmt.Print("Do you want to download an install this update? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')

			choice := string([]byte(input)[0])
			switch choice {
			case "y":
				return true
			case "n":
				return false
			default:
				fmt.Printf("%s is not a valid choice", choice)
				continue
			}
		}
	}

	return false
}

func (c *Client) UpdateBinary(name, version, path string) error {
	downloadDir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadDir)

	newBin, err := c.DownloadVersion(version, downloadDir)
	if err != nil {
		return err
	}

	fmt.Println(path)
	fmt.Println(newBin)
	//err = copyFile(newBin, path)
	//if err != nil {
	//	return err
	//}
	//
	//if err := os.Chmod(oldBin, 0755); err != nil {
	//	return err
	//}

	return nil
}

func (c *Client) DownloadVersion(version string, downloadDir string) (string, error) {
	c.Logger.Printf("Downloading cpd version %s...", version)
	startTime := time.Now()

	binName := fmt.Sprintf("cpd-v%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	downloader := s3manager.NewDownloader(c.session)

	downloadBinPath := filepath.Join(downloadDir, binName)
	downloadBin, err := os.Create(downloadBinPath)
	if err != nil {
		return "", err
	}

	bytes, err := downloader.Download(downloadBin, &s3.GetObjectInput{
		Bucket: aws.String(S3BinBucket),
		Key:    aws.String(fmt.Sprintf("cpd/%s", binName)),
	})
	if err != nil {
		return "", err
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := time.Since(startTime).Seconds()
	c.Logger.Printf("Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)", mb, timeSpent, mb/timeSpent)

	return downloadBinPath, nil
}

func (c *Client) TouchUpdateCheckFile() error {
	updateFile, err := homedir.Expand(UpdateCheckFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(updateFile); os.IsNotExist(err) {
		if f, err := os.Create(updateFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := os.Chtimes(updateFile, time.Now(), time.Now()); err != nil {
		return err
	}
	return nil
}

// copyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and removes
// the dst if it exists.
func copyFile(src, dst string) error {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	return err
}

func formatError(profiles []string, origErrors error) error {
	var newErrors *multierror.Error
	if e, ok := (origErrors).(*multierror.Error); ok {
		newErrors = multierror.Append(newErrors, fmt.Errorf("failed to find aws credentials in profiles: %s",
			strings.Join(AWSProfiles, ", ")),
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
			newErrors = multierror.Append(newErrors, fmt.Errorf("  %s", strings.Replace(errMsg.Error(), "\n", " ", -1)))
		}
	}
	return newErrors.ErrorOrNil()
}
