package update

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/update/s3"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	S3BinBucket          = "confluent.cloud"
	S3BinRegion          = "us-west-2"
	S3BinPrefix          = "cli/binaries"
	S3ReleaseNotesPrefix = "cli/release-notes"
	CheckFileFmt         = "%s/.confluent/update_check"
	CheckInterval        = 24 * time.Hour
)

// NewClient returns a new update.Client configured for the CLI
func NewClient(disableUpdateCheck bool, logger *log.Logger) update.Client {
	// The following function will never err, since "_" is a valid separator.
	objectKey, _ := s3.NewPrefixedKey(S3BinPrefix, "_", true)

	repo := s3.NewPublicRepo(&s3.PublicRepoParams{
		S3BinRegion:          S3BinRegion,
		S3BinBucket:          S3BinBucket,
		S3BinPrefix:          S3BinPrefix,
		S3ReleaseNotesPrefix: S3ReleaseNotesPrefix,
		S3ObjectKey:          objectKey,
		Logger:               logger,
	})
	homedir, _ := os.UserHomeDir()
	return update.NewClient(&update.ClientParams{
		Repository:    repo,
		DisableCheck:  disableUpdateCheck,
		CheckFile:     fmt.Sprintf(CheckFileFmt, homedir),
		CheckInterval: CheckInterval,
		Logger:        logger,
		Out:           os.Stdout,
	})
}

type command struct {
	Command *cobra.Command
	version *pversion.Version
	logger  *log.Logger
	client  update.Client
	// for testing
	analyticsClient analytics.Client
}

// New returns the command for the built-in updater.
func New(logger *log.Logger, version *pversion.Version, client update.Client, analytics analytics.Client) *cobra.Command {
	cmd := &command{
		version:         version,
		logger:          logger,
		client:          client,
		analyticsClient: analytics,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update the %s.", pversion.FullCLIName),
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.update),
	}
	c.Command.Flags().BoolP("yes", "y", false, "Update without prompting.")
	c.Command.Flags().SortFlags = false
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	updateYes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, errors.ReadingYesFlagErrorMsg)
	}
	utils.ErrPrintln(cmd, errors.CheckingForUpdatesMsg)
	updateAvailable, latestVersion, err := c.client.CheckForUpdates(pversion.CLIName, c.version.Version, true)
	if err != nil {
		return errors.NewUpdateClientWrapError(err, errors.CheckingForUpdateErrorMsg)
	}

	if !updateAvailable {
		utils.Println(cmd, errors.UpToDateMsg)
		return nil
	}

	releaseNotes := c.getReleaseNotes(latestVersion)

	// HACK: our packaging doesn't include the "v" in the version, so we add it back so that the prompt is consistent
	//   example S3 path: ccloud-cli/binaries/0.50.0/ccloud_0.50.0_darwin_amd64
	// Without this hack, the prompt looks like
	//   Current Version: v0.0.0
	//   Latest Version:  0.50.0
	// Unfortunately the "UpdateBinary" output will still show 0.50.0, and we can't hack that since it must match S3
	if !c.client.PromptToDownload(pversion.CLIName, c.version.Version, "v"+latestVersion, releaseNotes, !updateYes) {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := c.client.UpdateBinary(pversion.CLIName, latestVersion, oldBin); err != nil {
		return errors.NewUpdateClientWrapError(err, errors.UpdateBinaryErrorMsg)
	}
	utils.ErrPrintf(cmd, errors.UpdateAutocompleteMsg, pversion.CLIName)

	return nil
}

func (c *command) getReleaseNotes(latestBinaryVersion string) string {
	latestReleaseNotesVersion, allReleaseNotes, err := c.client.GetLatestReleaseNotes(c.version.Version)

	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf(errors.ObtainingReleaseNotesErrorMsg, err)
	} else {
		isSameVersion, err := sameVersionCheck(latestBinaryVersion, latestReleaseNotesVersion)
		if err != nil {
			errMsg = fmt.Sprintf(errors.ReleaseNotesVersionCheckErrorMsg, err)
		}
		if !isSameVersion {
			errMsg = fmt.Sprintf(errors.ReleaseNotesVersionMismatchErrorMsg, latestBinaryVersion, latestReleaseNotesVersion)
		}
	}

	if errMsg != "" {
		c.logger.Debugf(errMsg)
		c.analyticsClient.SetSpecialProperty(analytics.ReleaseNotesErrorPropertiesKeys, errMsg)
		return ""
	}

	return strings.Join(allReleaseNotes, "\n")
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
