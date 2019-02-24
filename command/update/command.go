package update

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cliCommand "github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/update/s3"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
	cliVersion "github.com/confluentinc/cli/version"
)

const (
	S3BinBucket   = "cloud-confluent-bin"
	S3BinRegion   = "us-west-2"
	S3BinPrefix   = "cpd"
	LastCheckFile = "~/.ccloud_update"
)

var (
	// in priority order to check for credentials
	AWSProfiles = []string{"confluent-dev", "confluent", "default"}
)

type command struct {
	Command *cobra.Command
	config  *shared.Config
	logger  *log.Logger
	cliName string
	version *cliVersion.Version
	// for testing
	prompt cliCommand.Prompt
}

// New returns the command for the built-in updater.
func New(cliName string, config *shared.Config, version *cliVersion.Version) *cobra.Command {
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
	creds, err := s3.GetCredentials(AWSProfiles)
	if err != nil {
		c.logger.Fatalf("error getting update client %s", err)
		return err
	}
	repo, err := s3.NewPrivateRepo(&s3.PrivateRepoParams{
		S3BinBucket: S3BinBucket,
		S3BinRegion: S3BinRegion,
		S3BinPrefix: S3BinPrefix,
		Credentials: creds,
		Logger:      c.logger,
	})
	if err != nil {
		return err
	}
	updateClient, err := NewUpdateClient(&Params{
		Repository: repo,
		Logger:     c.logger,
	})
	if err != nil {
		c.logger.Fatalf("error getting update client %s", err)
	}

	c.logger.Print("Checking for updates...")
	updateAvailable, latestVersion, err := updateClient.CheckForUpdates("cpd", c.version.Version)
	if err != nil {
		c.logger.Fatalf("error checking for updates: %s", err)
	}

	if !updateAvailable {
		c.logger.Printf("Already up to date")
		return nil
	}

	fmt.Println(updateYes, latestVersion)
	doUpdate := updateClient.PromptToDownload(c.cliName, c.version.Version, latestVersion, !updateYes)
	if !doUpdate {
		return nil
	}

	oldBin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := updateClient.UpdateBinary("cpd", latestVersion, oldBin); err != nil {
		return err
	}

	if err := updateClient.TouchUpdateCheckFile(); err != nil {
		c.logger.Fatalf("error checking for updates: %s", err)
	}
	return nil
}
