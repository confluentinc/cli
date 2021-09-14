package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/update/s3"
	"github.com/confluentinc/cli/internal/pkg/utils"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	S3BinBucket             = "confluent.cloud"
	S3BinRegion             = "us-west-2"
	S3BinPrefixFmt          = "%s-cli/binaries"
	S3ReleaseNotesPrefixFmt = "%s-cli/release-notes"
	CheckFileFmt            = "%s/.%s/update_check"
	CheckInterval           = 24 * time.Hour
)

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

type command struct {
	Command *cobra.Command
	cliName string
	version *cliVersion.Version
	logger  *log.Logger
	client  update.Client
	// for testing
	analyticsClient analytics.Client
}

// New returns the command for the built-in updater.
func New(cliName string, logger *log.Logger, version *cliVersion.Version,
	client update.Client, analytics analytics.Client) *cobra.Command {
	cmd := &command{
		cliName:         cliName,
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
		Short: fmt.Sprintf("Update the %s.", cliVersion.GetFullCLIName(c.cliName)),
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.update),
	}
	c.Command.Flags().BoolP("yes", "y", false, "Update without prompting.")
	c.Command.Flags().Bool("major", false, "Allow major version updates.")
	c.Command.Flags().SortFlags = false
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
	latestMajorVersion, latestMinorVersion, err := c.client.CheckForUpdates(c.cliName, c.version.Version, true)
	if err != nil {
		return errors.NewUpdateClientWrapError(err, errors.CheckingForUpdateErrorMsg, c.cliName)
	}

	if latestMajorVersion == "" && latestMinorVersion == "" {
		utils.Println(cmd, errors.UpToDateMsg)
		return nil
	}

	if latestMajorVersion != "" && latestMinorVersion == "" && !major {
		utils.Printf(cmd, errors.MajorVersionUpdateMsg, c.cliName)
		return nil
	}

	if latestMajorVersion == "" && major {
		utils.Print(cmd, errors.NoMajorVersionUpdateMsg)
		return nil
	}

	isMajorVersionUpdate := major && latestMajorVersion != ""

	updateName := c.cliName
	updateVersion := latestMinorVersion
	if isMajorVersionUpdate {
		updateName = "confluent"
		updateVersion = latestMajorVersion
	}

	releaseNotes := c.getReleaseNotes(updateName, updateVersion)

	// HACK: our packaging doesn't include the "v" in the version, so we add it back so that the prompt is consistent
	//   example S3 path: ccloud-cli/binaries/0.50.0/ccloud_0.50.0_darwin_amd64
	// Without this hack, the prompt looks like
	//   Current Version: v0.0.0
	//   Latest Version:  0.50.0
	// Unfortunately the "UpdateBinary" output will still show 0.50.0, and we can't hack that since it must match S3
	if !c.client.PromptToDownload(c.cliName, c.version.Version, "v"+updateVersion, releaseNotes, !updateYes) {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}

	if err := c.client.UpdateBinary(updateName, updateVersion, oldBin); err != nil {
		return errors.NewUpdateClientWrapError(err, errors.UpdateBinaryErrorMsg, c.cliName)
	}

	if isMajorVersionUpdate {
		var current, other *v3.Config

		for _, cliName := range []string{"confluent", "ccloud"} {
			cfg, err := getConfig(cliName)
			if err != nil {
				return err
			}

			if cliName == c.cliName {
				current = cfg
			} else {
				other = cfg
			}
		}

		if other != nil {
			current = mergeConfigs(current, other)
		}

		current.CLIName = "confluent"

		if err := backupConfig("confluent"); err != nil {
			return err
		}

		if err := current.Save(); err != nil {
			return err
		}
	}

	utils.ErrPrintf(cmd, errors.UpdateAutocompleteMsg, updateName)
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

func getConfig(cliName string) (*v3.Config, error) {
	path, err := getConfigPath(cliName)
	if err != nil {
		return nil, err
	}

	if !utils.DoesPathExist(path) {
		return nil, nil
	}

	cfg := new(v3.Config)
	cfg.Filename = path
	err = cfg.Load()
	return cfg, err
}

// mergeConfigs merges the current CLI config with another config, if it exists.
func mergeConfigs(current, other *v3.Config) *v3.Config {
	current.DisableUpdates = current.DisableUpdates || other.DisableUpdates
	current.DisableUpdateCheck = current.DisableUpdateCheck || other.DisableUpdateCheck
	current.NoBrowser = current.NoBrowser || other.NoBrowser

	if current.CurrentContext == "" {
		current.CurrentContext = other.CurrentContext
	}

	for name, ctx := range other.Contexts {
		current.Contexts[name] = ctx
	}
	for name, platform := range other.Platforms {
		current.Platforms[name] = platform
	}
	for name, credential := range other.Credentials {
		current.Credentials[name] = credential
	}
	for name, state := range current.ContextStates {
		current.ContextStates[name] = state
	}

	return current
}

func backupConfig(cliName string) error {
	path, err := getConfigPath(cliName)
	if err != nil {
		return err
	}

	return os.Rename(path, path+".old")
}

func getConfigPath(cliName string) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "."+cliName, "config.json"), nil
}
