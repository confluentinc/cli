package update

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/exec"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/update"
	"github.com/confluentinc/cli/v3/pkg/update/s3"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
)

const (
	S3BinBucket             = "confluent.cloud"
	S3BinRegion             = "us-west-2"
	S3BinPrefixFmt          = "%s-cli/binaries"
	S3ReleaseNotesPrefixFmt = "%s-cli/release-notes"
	CheckFileFmt            = "%s/.%s/update_check"
	CheckInterval           = 24 * time.Hour
)

const homebrewFormula = "confluentinc/tap/cli"

type command struct {
	*pcmd.CLICommand
	version *pversion.Version
	client  update.Client
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, client update.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "update",
		Short:  fmt.Sprintf("Update the %s.", pversion.FullCLIName),
		Args:   cobra.NoArgs,
		Hidden: cfg.DisableUpdates,
	}

	cmd.Flags().BoolP("yes", "y", false, "Update without prompting.")
	cmd.Flags().Bool("major", false, "Allow major version updates.")
	cmd.Flags().Bool("no-verify", false, "Skip checksum verification of new binary.")

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		version:    cfg.Version,
		client:     client,
	}
	cmd.RunE = c.update

	return cmd
}

// NewClient returns a new update.Client configured for the CLI
func NewClient(cliName string, disableUpdateCheck bool) update.Client {
	repo := s3.NewPublicRepo(&s3.PublicRepoParams{
		S3BinRegion:             S3BinRegion,
		S3BinBucket:             S3BinBucket,
		S3BinPrefixFmt:          S3BinPrefixFmt,
		S3ReleaseNotesPrefixFmt: S3ReleaseNotesPrefixFmt,
	})
	homedir, _ := os.UserHomeDir()
	return update.NewClient(&update.ClientParams{
		Repository:    repo,
		DisableCheck:  disableUpdateCheck,
		CheckFile:     fmt.Sprintf(CheckFileFmt, homedir, cliName),
		CheckInterval: CheckInterval,
		Out:           os.Stdout,
	})
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	if c.Config.DisableUpdates {
		message := "updates are disabled for this binary"
		if IsHomebrew() {
			return errors.NewErrorWithSuggestions(
				message,
				fmt.Sprintf("If installed with Homebrew, run `brew upgrade %s`.", homebrewFormula),
			)
		}

		return errors.NewErrorWithSuggestions(
			message,
			"Use a package manager to update the binary, if applicable. Otherwise, consider deleting this binary and re-installing a newer version.",
		)
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, errors.ReadingYesFlagErrorMsg)
	}

	major, err := cmd.Flags().GetBool("major")
	if err != nil {
		return err
	}

	noVerify, err := cmd.Flags().GetBool("no-verify")
	if err != nil {
		return err
	}

	output.ErrPrintln("Checking for updates...")
	latestMajorVersion, latestMinorVersion, err := c.client.CheckForUpdates(pversion.CLIName, c.version.Version, true)
	if err != nil {
		return errors.NewUpdateClientWrapError(err, errors.CheckingForUpdateErrorMsg)
	}

	if latestMajorVersion == "" && latestMinorVersion == "" {
		output.Println("Already up to date.")
		return nil
	}

	if latestMajorVersion != "" && latestMinorVersion == "" && !major {
		output.Println("The only available update is a major version update. Use `confluent update --major` to accept the update.")
		return nil
	}

	if latestMajorVersion == "" && major {
		output.Println("No major version updates are available.")
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
	if !c.client.PromptToDownload(pversion.CLIName, c.version.Version, "v"+updateVersion, releaseNotes, !yes) {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := c.client.UpdateBinary(pversion.CLIName, updateVersion, oldBin, noVerify); err != nil {
		return errors.NewUpdateClientWrapError(err, errors.UpdateBinaryErrorMsg)
	}

	return nil
}

func IsHomebrew() bool {
	out, err := exec.NewCommand("brew", "ls", homebrewFormula).Output()
	if err != nil {
		return false
	}

	homebrewPaths := strings.Split(strings.TrimSpace(string(out)), "\n")
	log.CliLogger.Tracef("Detected Homebrew installation: %v", homebrewPaths)

	path, err := os.Executable()
	if err != nil {
		return false
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}

	log.CliLogger.Tracef("Executable path: %s", path)

	return slices.Contains(homebrewPaths, path)
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
		log.CliLogger.Debugf(errMsg)
		return ""
	}

	return strings.Join(allReleaseNotes, "\n")
}

func sameVersionCheck(v1, v2 string) (bool, error) {
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
