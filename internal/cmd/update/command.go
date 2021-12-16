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
	S3BinBucket             = "confluent.cloud"
	S3BinRegion             = "us-west-2"
	S3BinPrefixFmt          = "%s-cli/binaries"
	S3ReleaseNotesPrefixFmt = "%s-cli/release-notes"
	CheckFileFmt            = "%s/.%s/update_check"
	CheckInterval           = 24 * time.Hour
)

type command struct {
	*pcmd.CLICommand
	version *pversion.Version
	logger  *log.Logger
	client  update.Client
	// for testing
	analyticsClient analytics.Client
}

func New(prerunner pcmd.PreRunner, logger *log.Logger, version *pversion.Version, client update.Client, analytics analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       fmt.Sprintf("Update the %s.", pversion.FullCLIName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireUpdatesEnabled},
	}

	cmd.Flags().BoolP("yes", "y", false, "Update without prompting.")
	cmd.Flags().Bool("major", false, "Allow major version updates.")

	c := &command{
		CLICommand:      pcmd.NewAnonymousCLICommand(cmd, prerunner),
		version:         version,
		logger:          logger,
		client:          client,
		analyticsClient: analytics,
	}

	c.RunE = pcmd.NewCLIRunE(c.update)

	return c.Command
}

// NewClient returns a new update.Client configured for the CLI
func NewClient(cliName string, disableUpdateCheck bool, logger *log.Logger) update.Client {
	repo := s3.NewPublicRepo(&s3.PublicRepoParams{
		S3BinRegion:             S3BinRegion,
		S3BinBucket:             S3BinBucket,
		S3BinPrefixFmt:          S3BinPrefixFmt,
		S3ReleaseNotesPrefixFmt: S3ReleaseNotesPrefixFmt,
		Logger:                  logger,
	})
	homedir, _ := os.UserHomeDir()
	return update.NewClient(&update.ClientParams{
		Repository:    repo,
		DisableCheck:  disableUpdateCheck,
		CheckFile:     fmt.Sprintf(CheckFileFmt, homedir, cliName),
		CheckInterval: CheckInterval,
		Logger:        logger,
		Out:           os.Stdout,
	})
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	updateYes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, errors.ReadingYesFlagErrorMsg)
	}

	major, err := cmd.Flags().GetBool("major")
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.CheckingForUpdatesMsg)
	latestMajorVersion, latestMinorVersion, err := c.client.CheckForUpdates(pversion.CLIName, c.version.Version, true)
	if err != nil {
		return errors.NewUpdateClientWrapError(err, errors.CheckingForUpdateErrorMsg)
	}

	if latestMajorVersion == "" && latestMinorVersion == "" {
		utils.Println(cmd, errors.UpToDateMsg)
		return nil
	}

	if latestMajorVersion != "" && latestMinorVersion == "" && !major {
		utils.Printf(cmd, errors.MajorVersionUpdateMsg, pversion.CLIName)
		return nil
	}

	if latestMajorVersion == "" && major {
		utils.Print(cmd, errors.NoMajorVersionUpdateMsg)
		return nil
	}

	isMajorVersionUpdate := major && latestMajorVersion != ""

	updateVersion := latestMinorVersion
	if isMajorVersionUpdate {
		updateVersion = latestMajorVersion
	}

	releaseNotes := c.getReleaseNotes(pversion.CLIName, updateVersion)

	// HACK: our packaging doesn't include the "v" in the version, so we add it back so that the prompt is consistent
	//   example S3 path: ccloud-cli/binaries/0.50.0/ccloud_0.50.0_darwin_amd64
	// Without this hack, the prompt looks like
	//   Current Version: v0.0.0
	//   Latest Version:  0.50.0
	// Unfortunately the "UpdateBinary" output will still show 0.50.0, and we can't hack that since it must match S3
	if !c.client.PromptToDownload(pversion.CLIName, c.version.Version, "v"+updateVersion, releaseNotes, !updateYes) {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := c.client.UpdateBinary(pversion.CLIName, updateVersion, oldBin); err != nil {
		return errors.NewUpdateClientWrapError(err, errors.UpdateBinaryErrorMsg)
	}

	utils.ErrPrintf(cmd, errors.UpdateAutocompleteMsg, pversion.CLIName)
	return nil
}

func (c *command) getReleaseNotes(cliName, latestBinaryVersion string) string {
	latestReleaseNotesVersion, allReleaseNotes, err := c.client.GetLatestReleaseNotes(cliName, c.version.Version)

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
