package update

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/update/s3"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	S3BinBucket          = "confluent.cloud"
	S3BinRegion          = "us-west-2"
	S3BinPrefix          = "%s-cli/binaries"
	S3ReleaseNotesPrefix = "%s-cli/release-notes"
	CheckFileFmt         = "~/.%s/update_check"
	CheckInterval        = 24 * time.Hour
)

// NewClient returns a new update.Client configured for the CLI
func NewClient(cliName string, disableUpdateCheck bool, logger *log.Logger) (update.Client, error) {
	objectKey, err := s3.NewPrefixedKey(fmt.Sprintf(S3BinPrefix, cliName), "_", true)
	if err != nil {
		return nil, err
	}
	repo := s3.NewPublicRepo(&s3.PublicRepoParams{
		S3BinRegion:          S3BinRegion,
		S3BinBucket:          S3BinBucket,
		S3BinPrefix:          fmt.Sprintf(S3BinPrefix, cliName),
		S3ReleaseNotesPrefix: fmt.Sprintf(S3ReleaseNotesPrefix, cliName),
		S3ObjectKey:          objectKey,
		Logger:               logger,
	})
	return update.NewClient(&update.ClientParams{
		Repository:    repo,
		DisableCheck:  disableUpdateCheck,
		CheckFile:     fmt.Sprintf(CheckFileFmt, cliName),
		CheckInterval: CheckInterval,
		Logger:        logger,
		Out:           os.Stdout,
	}), nil
}

type command struct {
	Command *cobra.Command
	cliName string
	config  *v3.Config
	version *cliVersion.Version
	logger  *log.Logger
	client  update.Client
	// for testing
	prompt          pcmd.Prompt
	analyticsClient analytics.Client
}

// New returns the command for the built-in updater.
func New(cliName string, logger *log.Logger, version *cliVersion.Version, prompt pcmd.Prompt,
	client update.Client, analytics analytics.Client) *cobra.Command {
	cmd := &command{
		cliName:         cliName,
		version:         version,
		logger:          logger,
		prompt:          prompt,
		client:          client,
		analyticsClient: analytics,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update the %s CLI.", c.cliName),
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	c.Command.Flags().Bool("yes", false, "Update without prompting.")
	c.Command.Flags().SortFlags = false
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	updateYes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, "error reading --yes as bool")
	}
	pcmd.ErrPrintln(cmd, "Checking for updates...")
	updateAvailable, latestVersion, err := c.client.CheckForUpdates(c.cliName, c.version.Version, true)
	if err != nil {
		c.Command.SilenceUsage = true
		return errors.Wrap(err, "error checking for updates")
	}

	if !updateAvailable {
		pcmd.Println(cmd, "Already up to date.")
		return nil
	}

	releaseNotes := c.getReleaseNotes(latestVersion)

	// HACK: our packaging doesn't include the "v" in the version, so we add it back so  that the prompt is consistent
	//   example S3 path: ccloud-cli/binaries/0.50.0/ccloud_0.50.0_darwin_amd64
	// Without this hack, the prompt looks like
	//   Current Version: v0.0.0
	//   Latest Version:  0.50.0
	// Unfortunately the "UpdateBinary" output will still show 0.50.0, and we can't hack that since it must match S3
	doUpdate := c.client.PromptToDownload(c.cliName, c.version.Version, "v"+latestVersion, releaseNotes, !updateYes)
	if !doUpdate {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := c.client.UpdateBinary(c.cliName, latestVersion, oldBin); err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, "Update your autocomplete scripts as instructed by: %s help completion\n", c.config.CLIName)

	return nil
}

func (c *command) getReleaseNotes(latestBinaryVersion string) string {
	latestReleaseNotesVersion, releaseNotes, err := c.client.GetLatestReleaseNotes()
	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf("error obtaining release notes: %s", err)
	} else {
		isSameVersion, err := sameVersionCheck(latestBinaryVersion, latestReleaseNotesVersion)
		if err != nil {
			errMsg = fmt.Sprintf("unable to perform release notes and binary version check: %s", err)
		}
		if !isSameVersion {
			errMsg = fmt.Sprintf("binary version (v%s) and latest release notes version (v%s) mismatch", latestBinaryVersion, latestReleaseNotesVersion)
		}
	}
	if errMsg != "" {
		c.logger.Debugf(errMsg)
		c.analyticsClient.SetSpecialProperty(analytics.ReleaseNotesErrorPropertiesKeys, errMsg)
		return ""
	}
	return releaseNotes
}

func sameVersionCheck(v1 string, v2 string) (bool, error) {
	version1, err := version.NewVersion(v1)
	if err != nil {
		return false, err
	}
	version2, err := version.NewVersion(v2)
	if err != nil {
		return false, err
	}
	return version1.Compare(version2) == 0, nil
}
