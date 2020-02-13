package update

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/update/s3"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	S3BinBucket   = "confluent.cloud"
	S3BinRegion   = "us-west-2"
	S3BinPrefix   = "%s-cli/binaries"
	CheckFileFmt  = "~/.%s/update_check"
	CheckInterval = 24 * time.Hour
)

// NewClient returns a new update.Client configured for the CLI
func NewClient(cliName string, disableUpdateCheck bool, logger *log.Logger) (update.Client, error) {
	objectKey, err := s3.NewPrefixedKey(fmt.Sprintf(S3BinPrefix, cliName), "_", true)
	if err != nil {
		return nil, err
	}
	repo := s3.NewPublicRepo(&s3.PublicRepoParams{
		S3BinRegion: S3BinRegion,
		S3BinBucket: S3BinBucket,
		S3BinPrefix: fmt.Sprintf(S3BinPrefix, cliName),
		S3ObjectKey: objectKey,
		Logger:      logger,
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
	config  *v2.Config
	version *cliVersion.Version
	logger  *log.Logger
	client  update.Client
	// for testing
	prompt pcmd.Prompt
}

// New returns the command for the built-in updater.
func New(cliName string, config *v2.Config, version *cliVersion.Version, prompt pcmd.Prompt,
	client update.Client) *cobra.Command {
	cmd := &command{
		cliName: cliName,
		config:  config,
		version: version,
		logger:  config.Logger,
		prompt:  prompt,
		client:  client,
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

func (c *command) update(cmd *cobra.Command, args []string) error {
	updateYes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, "error reading --yes as bool")
	}

	pcmd.Println(cmd, "Checking for updates...")
	updateAvailable, latestVersion, err := c.client.CheckForUpdates(c.cliName, c.version.Version, true)
	if err != nil {
		c.Command.SilenceUsage = true
		return errors.Wrap(err, "Error checking for updates.")
	}

	if !updateAvailable {
		pcmd.Println(cmd, "Already up to date.")
		return nil
	}

	// HACK: our packaging doesn't include the "v" in the version, so we add it back so  that the prompt is consistent
	//   example S3 path: ccloud-cli/binaries/0.50.0/ccloud_0.50.0_darwin_amd64
	// Without this hack, the prompt looks like
	//   Current Version: v0.0.0
	//   Latest Version:  0.50.0
	// Unfortunately the "UpdateBinary" output will still show 0.50.0, and we can't hack that since it must match S3
	doUpdate := c.client.PromptToDownload(c.cliName, c.version.Version, "v"+latestVersion, !updateYes)
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
	pcmd.Printf(cmd, "Update your autocomplete scripts as instructed by: %s help completion\n", c.config.CLIName)

	return nil
}
