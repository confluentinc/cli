package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atrox/homedir"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const longDescription = `Use these commands to try out Confluent Platform by running a single-node
instance locally on your machine. This set of commands are NOT intended for production use.

You must download and install Confluent Platform from https://www.confluent.io/download on your
machine. These commands require the path to the installation directory via the --path flag or
the CONFLUENT_HOME environment variable.

You can use these commands to explore, test, experiment, and otherwise familiarize yourself
with Confluent Platform.

DO NOT use these commands to setup or manage Confluent Platform in production.
`

var (
	commonInstallDirs = []string{
		"./confluent*",
		"/opt/confluent*",
		"/usr/local/confluent*",
		"~/confluent*",
		"~/Downloads/confluent*",
	}

	validCPInstallBinCanaries = []string{
		"connect-distributed",
		"kafka-server-start",
		"ksql-server-start",
		"zookeeper-server-start",
	}
	validCPInstallEtcCanary = filepath.Join("etc", "schema-registry", "connect-avro-distributed.properties")
)

type command struct {
	*cobra.Command
	shell ShellRunner
	log   *log.Logger
	fs    io.FileSystem
}

// New returns the Cobra command for `local`.
func New(rootCmd *cobra.Command, prerunner pcmd.PreRunner, shell ShellRunner, log *log.Logger, fs io.FileSystem) *cobra.Command {
	localCmd := &command{
		Command: &cobra.Command{
			Use:               "local",
			Short:             "Manage a local Confluent Platform development environment.",
			Long:              longDescription,
			Args:              cobra.ArbitraryArgs,
			PersistentPreRunE: prerunner.Anonymous(),
		},
		shell: shell,
		log:   log,
		fs:    fs,
	}
	localCmd.Command.RunE = localCmd.run
	localCmd.Flags().String("path", "", "Path to Confluent Platform install directory.")
	localCmd.Flags().SortFlags = false
	// This is used for "confluent help local foo" and "confluent local foo --help"
	localCmd.Command.SetHelpFunc(localCmd.help)

	// Explicit suggestions since we can't use cobra's "SuggestFor" for bash commands
	for _, cmd := range []string{"start", "stop"} {
		rootCmd.AddCommand(localCommandError(cmd))
	}

	return localCmd.Command
}

func (c *command) parsePath(cmd *cobra.Command, args []string) (string, error) {
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		return "", errors.HandleCommon(err, cmd)
	}
	if path == "" {
		home, found := os.LookupEnv("CONFLUENT_HOME")
		if found {
			path = home
		} else {
			// try to determine the confluent install dir heuristically
			if home, found, err := determineConfluentInstallDir(c.fs); err != nil {
				return "", err
			} else if found {
				path, err = filepath.Abs(home)
				if err != nil {
					return "", err
				}
			} else if len(args) != 0 { // don't error if no args specified, we'll just show usage
				return "", fmt.Errorf("Pass --path /path/to/confluent flag or set environment variable CONFLUENT_HOME")
			}
		}
	}
	return path, nil
}

func (c *command) run(cmd *cobra.Command, args []string) error {
	path, err := c.parsePath(cmd, args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	c.log.Warnf("Using Confluent installation dir: %s", path)
	c.log.Warnf("To override Confluent installation dir, pass --path /path/to/confluent flag or set environment variable CONFLUENT_HOME")
	err = c.runBashCommand(path, "main", args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

// versionedDirectory is a type that implements the sort.Interface interface
// so that versions can be sorted and the original directory path returned.
type versionedDirectory struct {
	dir string
	ver *version.Version
}

func (v *versionedDirectory) String() string {
	return v.dir
}

type byVersion []*versionedDirectory

func (b byVersion) Len() int {
	return len(b)
}
func (b byVersion) Less(i, j int) bool {
	return b[i].ver.LessThan(b[j].ver)
}
func (b byVersion) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Heuristically determine the Confluent installation directory.
//
// Algorithm:
//   1. Search for a dir matching confluent* (glob) in the common places in order: (., /opt, /usr/local, ~, ~/Downloads)
//      This list is ordered by priority (always prefer /opt to ~/Downloads for example).
//      But each directory may contain multiple matches (e.g., /opt/confluent-5.2.2, /opt/confluent-4.1.0, etc).
//   2. For each match, look for multiple well-known files as canaries to ensure it's a valid CP install dir.
//   3. If it's a valid install dir, try to extract a version from the format "confluent-<version>" and collect all versions
//   4. If there were any versioned dirs, sort them by version and return the dir with the latest version
//   5. If there were no versioned dirs but there was a match, return it (should be a dir just named "confluent" like /opt/confluent)
func determineConfluentInstallDir(fs io.FileSystem) (string, bool, error) {
	for _, dir := range commonInstallDirs {
		dir, err := homedir.Expand(dir)
		if err != nil {
			return "", false, err
		}
		dir = filepath.Clean(dir)
		if matches, err := fs.Glob(dir); err != nil {
			return "", false, err
		} else if len(matches) > 0 {
			// We have at least one match in this directory.
			// Let's validate each to see if it's a real CP install dir.
			// If there's more than one, then we'll choose the newest version.
			foundValid := false
			var versions []*versionedDirectory
			for _, dir := range matches {
				// MacOS replaces homedir with ~ under fs.Glob, so we have to
				// call homedir.Expand again under each globbed match
				dir, err := homedir.Expand(dir)
				if err != nil {
					return "", false, err
				}
				if valid, err := validateConfluentPlatformInstallDir(fs, dir); err != nil {
					return "", false, err
				} else if !valid {
					// Skip this match because it doesn't look like a real confluent install dir
					continue
				}
				foundValid = true
				i := strings.LastIndex(dir, "confluent-")
				if i >= 0 {
					v, err := version.NewSemver(dir[i+len("confluent-"):])
					if err != nil {
						return "", false, err
					}
					versions = append(versions, &versionedDirectory{dir: dir, ver: v})
				}
			}
			// we foundValid at least one versioned directory
			if len(versions) > 0 {
				sort.Sort(byVersion(versions))
				return versions[len(versions)-1].dir, true, nil
			} else if foundValid {
				// no versioned directories so the match might just be a dir named "confluent"
				return matches[0], true, nil
			}
		}
	}
	return "", false, nil
}

func (c *command) help(cmd *cobra.Command, args []string) {
	// if "confluent help local foo bar" is called, args is empty, so we just show usage :(
	// if "confluent local foo bar --help" is called, args is [local, foo, bar, --help]
	// transform args: drop first "local" and any "--help" flag. [local, foo, bar, --help] -> [help, foo, bar]
	if len(args) > 0 && args[0] == "local" {
		args = args[1:]
	}
	var a []string
	for _, arg := range args {
		if arg != "--help" {
			a = append(a, arg)
		}
	}
	// Ignore error and attempt to print help anyway
	path, _ := c.parsePath(cmd, args)
	_ = c.runBashCommand(path, "help", a)
}

func (c *command) runBashCommand(path string, command string, args []string) error {
	c.shell.Init(os.Stdout, os.Stderr)
	c.shell.Export("CONFLUENT_HOME", path)
	c.shell.Export("CONFLUENT_CURRENT", os.Getenv("CONFLUENT_CURRENT"))
	c.shell.Export("TMPDIR", os.Getenv("TMPDIR"))
	c.shell.Export("JAVA_HOME", os.Getenv("JAVA_HOME"))
	c.shell.Export("PATH", os.Getenv("PATH"))
	c.shell.Export("HOME", os.Getenv("HOME"))
	err := c.shell.Source("cp_cli/confluent.sh", Asset)
	if err != nil {
		return err
	}

	_, err = c.shell.Run(command, args)
	if err != nil {
		return err
	}
	return nil
}

func validateConfluentPlatformInstallDir(fs io.FileSystem, dir string) (bool, error) {
	// Validate home directory exists and is in fact a directory
	f, err := fs.Stat(dir)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	case !f.IsDir():
		return false, nil
	}

	// Validate bin directory contents
	filesToCheck := make(map[string]bool, len(validCPInstallBinCanaries))
	for _, name := range validCPInstallBinCanaries {
		filesToCheck[filepath.Join(dir, "bin", name)] = false
	}

	files, err := fs.ReadDir(filepath.Join(dir, "bin"))
	if err != nil {
		return false, err
	}
	for _, f := range files {
		fullPath := filepath.Join(dir, "bin", f.Name())
		if _, ok := filesToCheck[fullPath]; ok {
			filesToCheck[fullPath] = true
		}
	}
	for _, v := range filesToCheck {
		if !v {
			return false, nil
		}
	}

	// Validate etc directory contents/location
	f, err = fs.Stat(filepath.Join(dir, validCPInstallEtcCanary))
	switch {
	case os.IsNotExist(err):
		// workaround for the cases when 'etc' is not under the same directory as 'bin'
		f, err = fs.Stat(filepath.Join(dir, "..", validCPInstallEtcCanary))
		switch {
		case os.IsNotExist(err):
			return false, nil
		case err != nil:
			return false, err
		}
	case err != nil:
		return false, err
	}

	// If we make it here, then its a real CP install dir. Hurray!
	return true, nil
}

func localCommandError(command string) *cobra.Command {
	err := fmt.Errorf(`unknown command "%s" for "confluent"

Did you mean this?
        local %s

Run 'confluent --help' for usage.`, command, command)

	runE := func(cmd *cobra.Command, args []string) error {
		return err
	}
	run := func(cmd *cobra.Command, args []string) {
		// We explicitly prepend "Error: " and append a newline to match the standard error printing format
		pcmd.ErrPrintf(cmd, "Error: %s\n", err.Error())
		// We exit 0 though because this means that the user explicitly requested help
	}

	cmd := &cobra.Command{Use: command, Hidden: true, SilenceUsage: true, RunE: runE}
	cmd.SetHelpFunc(run)
	return cmd
}
