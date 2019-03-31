package update

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/terminal"
	"github.com/confluentinc/cli/internal/pkg/updater"
	"github.com/confluentinc/cli/internal/pkg/updater/s3"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	S3BinBucket   = "confluent.cloud"
	S3BinRegion   = "us-west-2"
	S3BinPrefix   = "ccloud-cli/binaries"
	LastCheckFile = "~/.ccloud_update"
)

var (
	// in priority order to check for credentials
	AWSProfiles = []string{"confluent-dev", "confluent", "default"}
)

type command struct {
	Command *cobra.Command
	cliName string
	config  *config.Config
	version *cliVersion.Version
	logger  *log.Logger
	// for testing
	prompt terminal.Prompt
}

// New returns the command for the built-in updater.
func New(cliName string, config *config.Config, version *cliVersion.Version, prompt terminal.Prompt) *cobra.Command {
	cmd := &command{
		cliName: cliName,
		config:  config,
		version: version,
		logger:  config.Logger,
		prompt:  prompt,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "update",
		Short: "Update ccloud",
		Long:  "Update ccloud",
		RunE:  c.update,
	}
	c.Command.Flags().Bool("yes", false, "Update without prompting.")
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	updateYes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return errors.Wrap(err, "error reading --yes as bool")
	}
	//repo, err := s3.NewPrivateRepo(&s3.PrivateRepoParams{
	//	S3BinBucket: S3BinBucket,
	//	S3BinRegion: S3BinRegion,
	//	S3BinPrefix: S3BinPrefix,
	//	AWSProfiles: AWSProfiles,
	//	logger:      c.logger,
	//})
	//if err != nil {
	//	return err
	//}
	repo := &s3.PublicRepo{
		S3BinRegion: S3BinRegion,
		S3BinBucket: S3BinBucket,
		S3BinPrefix: S3BinPrefix,
		Logger:      c.logger,
	}
	updateClient := updater.NewUpdateClient(repo, LastCheckFile, c.logger)

	c.prompt.Println("Checking for updates...")
	updateAvailable, latestVersion, err := updateClient.CheckForUpdates(c.cliName, c.version.Version)
	if err != nil {
		c.logger.Errorf("error checking for updates: %s", err)
	}

	if !updateAvailable {
		c.prompt.Println("Already up to date")
		return nil
	}

	doUpdate := updateClient.PromptToDownload(c.cliName, c.version.Version, latestVersion, !updateYes)
	if !doUpdate {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := updateClient.UpdateBinary(c.cliName, latestVersion, oldBin); err != nil {
		return err
	}

	if err := updateClient.TouchUpdateCheckFile(); err != nil {
		c.logger.Errorf("error touching last check file: %s", err)
		return err
	}
	return nil
}
