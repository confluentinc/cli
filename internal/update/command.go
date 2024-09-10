package update

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	pupdate "github.com/confluentinc/cli/v3/pkg/update"
)

const homebrewTap = "confluentinc/tap/cli"

type command struct {
	*pcmd.CLICommand
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "update",
		Short:  "Update the Confluent CLI.",
		Args:   cobra.NoArgs,
		Hidden: cfg.DisableUpdates,
	}

	cmd.Flags().BoolP("yes", "y", false, "Update without prompting.")
	cmd.Flags().Bool("major", false, "Allow major version updates.")
	cmd.Flags().Bool("no-verify", false, "Skip checksum verification of new binary.")

	c := &command{pcmd.NewAnonymousCLICommand(cmd, prerunner)}
	cmd.RunE = c.update

	return cmd
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	if c.Config.DisableUpdates {
		message := "updates are disabled for this binary"
		if isHomebrew() {
			return errors.NewErrorWithSuggestions(
				message,
				fmt.Sprintf("If installed with Homebrew, run `brew upgrade %s`.", homebrewTap),
			)
		}

		return errors.NewErrorWithSuggestions(
			message,
			"Use a package manager to update the binary, if applicable. Otherwise, consider deleting this binary and re-installing a newer version.",
		)
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	major, err := cmd.Flags().GetBool("major")
	if err != nil {
		return err
	}

	noVerify, err := cmd.Flags().GetBool("no-verify")
	if err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, "Checking for updates...")

	current, err := version.NewVersion(c.Config.Version.Version)
	if err != nil {
		return err
	}

	client := pupdate.NewClient(c.Config.IsTest)

	binaries, err := client.GetBinaries()
	if err != nil {
		return err
	}

	minorVersions, majorVersions := pupdate.FilterUpdates(binaries, current, major)

	if len(minorVersions) == 0 && len(majorVersions) == 0 {
		output.Println(c.Config.EnableColor, "Already up to date.")
		return nil
	} else if !major && len(minorVersions) == 0 && len(majorVersions) > 0 {
		output.Println(c.Config.EnableColor, "The only available update is a major version update. Use `confluent update --major` to accept the update.")
	}

	versions := minorVersions
	if major {
		versions = majorVersions
	}

	output.Printf(c.Config.EnableColor, "New version of confluent is available\n")
	output.Printf(c.Config.EnableColor, "Current Version: %s\n", c.Config.Version.Version)
	output.Printf(c.Config.EnableColor, "Latest Version:  v%s\n", versions[len(versions)-1])
	output.Printf(c.Config.EnableColor, "\n")

	const maxReleaseNotes = 5

	for _, version := range versions[max(len(versions)-maxReleaseNotes, 0):] {
		releaseNotes, err := client.GetReleaseNotes(version.String())
		if err != nil {
			log.CliLogger.Warnf(`Failed to fetch release notes for version "%s": %v`, version, err)
			continue
		}

		output.Print(false, releaseNotes)
	}

	if len(versions) > maxReleaseNotes {
		output.Println(c.Config.EnableColor, "For all release notes, see: https://docs.confluent.io/confluent-cli/current/release-notes.html")
		output.Println(c.Config.EnableColor, "")
	}

	if !yes {
		f := form.New(form.Field{
			ID:        "download",
			Prompt:    "Do you want to download and install this update?",
			IsYesOrNo: true,
		})
		if err := f.Prompt(form.NewPrompt()); err != nil {
			return err
		}
		if !f.Responses["download"].(bool) {
			return nil
		}
	}

	version := versions[len(versions)-1]

	output.Printf(c.Config.EnableColor, "Downloading confluent version %s...\n", version)

	filename := fmt.Sprintf("confluent_%s_%s", getOs(), runtime.GOARCH)
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	binary, err := client.GetBinary(version.String(), filename)
	if err != nil {
		return fmt.Errorf("unable to download confluent version %s for %s/%s: %w", version, runtime.GOOS, runtime.GOARCH, err)
	}
	defer binary.Close()

	opts := update.Options{}

	if !noVerify {
		checksums, err := client.GetBinaryChecksums(version.String())
		if err != nil {
			return fmt.Errorf(`unable to fetch checksum for version "%s" of "%s": %w`, version, filename, err)
		}

		for _, line := range strings.Split(checksums, "\n") {
			if strings.HasSuffix(line, filename) {
				checksum := strings.Split(line, " ")[0]
				log.CliLogger.Debugf(`Checksum for version "%s" of "%s": %s`, version, filename, checksum)

				opts.Checksum = make([]byte, len(checksum)/2)
				if _, err := hex.Decode(opts.Checksum, []byte(checksum)); err != nil {
					return err
				}

				break
			}
		}
	}

	if !c.Config.IsTest {
		if err := update.Apply(binary, opts); err != nil {
			return err
		}
	}

	output.Println(c.Config.EnableColor, "Done.")

	return nil
}

func isHomebrew() bool {
	out, err := exec.Command("brew", "ls", homebrewTap).Output()
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

func getOs() string {
	if runtime.GOOS == "linux" {
		stderr := new(bytes.Buffer)

		cmd := exec.Command("ldd", "--version")
		cmd.Stderr = stderr
		_ = cmd.Run()

		if strings.Contains(stderr.String(), "musl") {
			return "alpine"
		}
	}

	return runtime.GOOS
}
