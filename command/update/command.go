package update

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cliCommand "github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/update/s3"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
	cliVersion "github.com/confluentinc/cli/version"
)

const (
	S3BinBucket   = "confluent.cloud"
	S3BinRegion   = "us-west-2"
	S3BinPrefix   = "ccloud-cli"
	LastCheckFile = "~/.ccloud_update"
)

var (
	// in priority order to check for credentials
	AWSProfiles = []string{"confluent-dev", "confluent", "default"}
)

type command struct {
	Command *cobra.Command
	cliName string
	plugins []string
	config  *shared.Config
	version *cliVersion.Version
	logger  *log.Logger
	// for testing
	prompt cliCommand.Prompt
}

// New returns the command for the built-in updater.
func New(cliName string, plugins []string, config *shared.Config, version *cliVersion.Version) *cobra.Command {
	cmd := &command{
		config:  config,
		logger:  config.Logger,
		cliName: cliName,
		version: version,
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
	updateClient := NewUpdateClient(repo, LastCheckFile, c.logger)

	c.logger.Print("Checking for updates...")
	updateAvailable, latestVersion, err := updateClient.CheckForUpdates(c.cliName, c.version.Version)
	if err != nil {
		c.logger.Fatalf("error checking for updates: %s", err)
	}

	if !updateAvailable {
		c.logger.Printf("Already up to date")
		return nil
	}

	doUpdate := updateClient.PromptToDownload(c.cliName, c.version.Version, latestVersion, !updateYes)
	if !doUpdate {
		return nil
	}

	//binaries := append([]string{c.cliName}, c.plugins...)
	//oldBin, err := os.Executable()
	//if err != nil {
	//	return err
	//}
	if err := updateClient.UpdateBinary(c.cliName, latestVersion, "/tmp/ccloud"); err != nil {
		return err
	}

	if err := updateClient.TouchUpdateCheckFile(); err != nil {
		c.logger.Fatalf("error checking for updates: %s", err)
	}
	return nil
}
