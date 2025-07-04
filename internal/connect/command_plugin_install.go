package connect

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/cpstructs"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/hub"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	invalidDirectoryErrorMsg       = `plugin directory "%s" does not exist`
	unexpectedInstallationErrorMsg = "unexpected installation type: %s"
	workerProcessRegexStr          = `org\.apache\.kafka\.connect\.cli\.Connect(Distributed|Standalone)`
)

type pluginInstallCommand struct {
	*pcmd.CLICommand
}

func newInstallCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install a Connect plugin.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Install the latest version of the Datagen connector into your local Confluent Platform environment.",
				Code: "confluent connect plugin install confluentinc/kafka-connect-datagen:latest",
			},
			examples.Example{
				Text: "Install the latest version of the Datagen connector in a user-specified directory and update a worker configuration file.",
				Code: "confluent connect plugin install confluentinc/kafka-connect-datagen:latest --plugin-directory $CONFLUENT_HOME/plugins --worker-configurations $CONFLUENT_HOME/etc/kafka/connect-distributed.properties",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.Flags().String("plugin-directory", "", "The plugin installation directory. If not specified, a default will be selected based on your Confluent Platform installation.")
	cmd.Flags().StringSlice("worker-configurations", []string{}, "A comma-separated list of paths to one or more Kafka Connect worker configuration files. Each worker file will be updated to load plugins from the plugin directory in addition to any prior directories.")
	cmd.Flags().String("confluent-platform", "", "The path to a Confluent Platform archive installation. By default, this command will search for Confluent Platform installations in common locations.")
	pcmd.AddDryRunFlag(cmd)
	cmd.Flags().Bool("force", false, "Proceed without user input.")

	cobra.CheckErr(cmd.MarkFlagDirname("plugin-directory"))

	c := &pluginInstallCommand{pcmd.NewAnonymousCLICommand(cmd, prerunner)}
	cmd.RunE = c.install

	return cmd
}

func (c *pluginInstallCommand) install(cmd *cobra.Command, args []string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("this command is not available on Windows")
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("plugin-directory") && cmd.Flags().Changed("worker-configurations") && cmd.Flags().Changed("confluent-platform") {
		return fmt.Errorf("at most two of `--plugin-directory`, `--worker-configurations`, and `--confluent-platform` may be set")
	}

	client, err := c.GetHubClient()
	if err != nil {
		return err
	}

	pluginManifest, err := getManifest(client, args[0])
	if err != nil {
		return err
	}

	pluginDir, err := getPluginDirFromFlag(cmd)
	if err != nil {
		return err
	}

	workerConfigs, err := getWorkerConfigsFromFlag(cmd)
	if err != nil {
		return err
	}

	var installation *platformInstallation
	prompt := form.NewPrompt()
	if pluginDir == "" {
		installation, err = getConfluentPlatformInstallation(cmd, prompt, force)
		if err != nil {
			return err
		}
		pluginDir, err = choosePluginDir(installation, prompt, force)
		if err != nil {
			return err
		}
	}

	// Check for, and possibly remove, existing installation
	previousInstallations, err := existingPluginInstallation(pluginDir, pluginManifest)
	if err != nil {
		return err
	}

	if err := removePluginInstallations(previousInstallations, prompt, dryRun, force); err != nil {
		return err
	}

	// Install
	if err := checkLicenseAcceptance(pluginManifest, prompt, force); err != nil {
		return err
	}

	if err := c.installPlugin(client, pluginManifest, args[0], pluginDir, dryRun); err != nil {
		return err
	}

	// Select and update worker-configurations
	if len(workerConfigs) == 0 {
		if installation == nil {
			installation, err = getConfluentPlatformInstallation(cmd, prompt, force)
			if err != nil {
				return err
			}
		}
		workerConfigs, err = chooseWorkerConfigs(cmd, installation, prompt, force)
		if err != nil {
			return err
		}
	}

	if err := updateWorkerConfigs(pluginDir, workerConfigs, dryRun); err != nil {
		return err
	}

	successStr := fmt.Sprintf("Installed %s %s.\n", pluginManifest.Title, pluginManifest.Version)
	if dryRun {
		successStr = utils.AddDryRunPrefix(successStr)
	}

	output.Println(c.Config.EnableColor, "")
	output.Print(c.Config.EnableColor, successStr)

	return nil
}

func parsePluginId(plugin string) (string, string, string, error) {
	err := errors.NewErrorWithSuggestions(
		"plugin not found",
		"Provide a path to a local file or provide a plugin ID from Confluent Hub with the format: `<owner>/<name>:<version>`.",
	)

	ownerNameSplit := strings.Split(plugin, "/")
	if len(ownerNameSplit) != 2 || ownerNameSplit[0] == "" {
		return "", "", "", err
	}
	nameVersionSplit := strings.Split(ownerNameSplit[1], ":")
	if len(nameVersionSplit) != 2 || nameVersionSplit[0] == "" || nameVersionSplit[1] == "" {
		return "", "", "", err
	}

	return ownerNameSplit[0], nameVersionSplit[0], nameVersionSplit[1], nil
}

func getManifest(client *hub.Client, id string) (*cpstructs.Manifest, error) {
	if utils.DoesPathExist(id) {
		// if installing plugin from local archive
		return getLocalManifest(id)
	}

	// if installing plugin from Confluent Hub
	owner, name, version, err := parsePluginId(id)
	if err != nil {
		return nil, err
	}

	return client.GetRemoteManifest(owner, name, version)
}

func getLocalManifest(archivePath string) (*cpstructs.Manifest, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local archive file %s: %w", archivePath, err)
	}
	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		isManifest, err := filepath.Match("*/manifest.json", filepath.ToSlash(zipFile.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to examine file %s inside local archive file %s: %w", zipFile.Name, archivePath, err)
		}

		if isManifest {
			manifestFile, err := zipFile.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open manifest file %s inside local archive file %s: %w", zipFile.Name, archivePath, err)
			}
			defer manifestFile.Close()

			pluginManifest := new(cpstructs.Manifest)
			if err := json.NewDecoder(manifestFile).Decode(&pluginManifest); err != nil {
				return nil, err
			}

			return pluginManifest, nil
		}
	}

	return nil, fmt.Errorf(`failed to find manifest file inside local archive file "%s"`, archivePath)
}

func getPluginDirFromFlag(cmd *cobra.Command) (string, error) {
	if !cmd.Flags().Changed("plugin-directory") {
		return "", nil
	}

	pluginDir, err := cmd.Flags().GetString("plugin-directory")
	if err != nil {
		return "", err
	}

	if pluginDir, err = filepath.Abs(pluginDir); err != nil {
		return "", err
	}

	if !utils.DoesPathExist(pluginDir) {
		return "", fmt.Errorf(invalidDirectoryErrorMsg, pluginDir)
	}

	return pluginDir, nil
}

func getWorkerConfigsFromFlag(cmd *cobra.Command) ([]string, error) {
	workerConfigs, err := cmd.Flags().GetStringSlice("worker-configurations")
	if err != nil {
		return nil, err
	}

	var errs *multierror.Error
	for _, workerConfig := range workerConfigs {
		if workerConfig, err = filepath.Abs(workerConfig); err != nil {
			errs = multierror.Append(errs, err)
		}

		if !utils.DoesPathExist(workerConfig) {
			errs = multierror.Append(errs, fmt.Errorf(`worker config file "%s" does not exist`, workerConfig))
		}
	}

	return workerConfigs, errs.ErrorOrNil()
}

func existingPluginInstallation(pluginDir string, pluginManifest *cpstructs.Manifest) ([]string, error) {
	// Bundled installations
	if utils.DoesPathExist(filepath.Join(pluginDir, pluginManifest.Name)) {
		return nil, fmt.Errorf("unable to install plugin because it is already bundled")
	}

	// Other previous installations
	immediateDirectory := filepath.Join(pluginDir, fmt.Sprintf("%s-%s", pluginManifest.Owner.Username, pluginManifest.Name))
	uberJar := filepath.Join(pluginDir, fmt.Sprintf("%s-%s.jar", pluginManifest.Name, pluginManifest.Version))

	var installations []string
	if utils.DoesPathExist(immediateDirectory) {
		installations = append(installations, immediateDirectory)
	}
	if utils.DoesPathExist(uberJar) {
		installations = append(installations, uberJar)
	}

	return installations, nil
}

func removePluginInstallations(previousInstallations []string, prompt form.Prompt, dryRun, force bool) error {
	if len(previousInstallations) > 0 {
		output.Println(false, "A version of this plugin is already installed and must be removed to continue.")
	}

	for _, previousInstallation := range previousInstallations {
		if force {
			output.Printf(false, "Uninstalling the existing version of the plugin located at \"%s\".\n", previousInstallation)
		} else {
			f := form.New(form.Field{
				ID:        "confirm",
				Prompt:    fmt.Sprintf("Do you want to uninstall an existing version of this plugin located at %s?", previousInstallation),
				IsYesOrNo: true,
			})
			if err := f.Prompt(prompt); err != nil {
				return err
			}
			if !f.Responses["confirm"].(bool) {
				return fmt.Errorf("previous versions must be uninstalled to continue")
			}
		}

		uninstallStr := "Successfully removed existing version.\n"
		if dryRun {
			output.Println(false, utils.AddDryRunPrefix(uninstallStr))
			return nil
		}

		if err := os.RemoveAll(previousInstallation); err != nil {
			return err
		}

		output.Println(false, uninstallStr)
	}

	if len(previousInstallations) > 0 {
		output.Println(false, "")
	}
	return nil
}

func (c *pluginInstallCommand) installPlugin(client *hub.Client, pluginManifest *cpstructs.Manifest, archivePath, pluginDir string, dryRun bool) error {
	installStr := fmt.Sprintf("Installing %s %s, provided by %s\n\n", pluginManifest.Title, pluginManifest.Version, pluginManifest.Owner.Name)
	if dryRun {
		output.Print(c.Config.EnableColor, utils.AddDryRunPrefix(installStr))
		return nil
	}
	output.Print(c.Config.EnableColor, installStr)

	if utils.DoesPathExist(archivePath) {
		return installFromLocal(pluginManifest, archivePath, pluginDir)
	}

	return installFromRemote(client, pluginManifest, pluginDir)
}

func installFromLocal(pluginManifest *cpstructs.Manifest, archivePath, pluginDir string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open local archive file %s: %w", archivePath, err)
	}
	defer zipReader.Close()

	return unzipPlugin(pluginManifest, zipReader.File, pluginDir)
}

func installFromRemote(client *hub.Client, pluginManifest *cpstructs.Manifest, pluginDir string) error {
	archive, err := client.GetRemoteArchive(pluginManifest)
	if err != nil {
		return err
	}

	checksumErrorMsg := `%s checksum for downloaded archive (%s) does not match checksum in manifest (%s) for plugin "%s"`
	calculatedMd5Checksum := fmt.Sprintf("%x", md5.Sum(archive))
	if calculatedMd5Checksum != pluginManifest.Archive.Md5 {
		return fmt.Errorf(checksumErrorMsg, "MD5", calculatedMd5Checksum, pluginManifest.Archive.Md5, pluginManifest.Name)
	}
	calculatedSha1Checksum := fmt.Sprintf("%x", sha1.Sum(archive))
	if calculatedSha1Checksum != pluginManifest.Archive.Sha1 {
		return fmt.Errorf(checksumErrorMsg, "SHA1", calculatedSha1Checksum, pluginManifest.Archive.Sha1, pluginManifest.Name)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("failed to open remote archive file %s: %w", archive, err)
	}

	return unzipPlugin(pluginManifest, zipReader.File, pluginDir)
}

func unzipPlugin(pluginManifest *cpstructs.Manifest, zipFiles []*zip.File, pluginDir string) error {
	relativeInstallationDir := filepath.Join(pluginDir, fmt.Sprintf("%s-%s", pluginManifest.Owner.Username, pluginManifest.Name))
	installationDir, err := filepath.Abs(relativeInstallationDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for directory %s: %w", relativeInstallationDir, err)
	}

	for _, zipFile := range zipFiles {
		versionPrefix := fmt.Sprintf("%s-%s-%s", pluginManifest.Owner.Username, pluginManifest.Name, pluginManifest.Version)
		destFilePath := filepath.Join(installationDir, strings.TrimPrefix(zipFile.Name, versionPrefix))

		createDirectoryErrorMsg := "failed to create directory %s on local storage: %w"
		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return fmt.Errorf(createDirectoryErrorMsg, destFilePath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
			return fmt.Errorf(createDirectoryErrorMsg, filepath.Dir(destFilePath), err)
		}

		zipFileReader, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf("failed to read file %s from archive: %w", zipFile.Name, err)
		}
		defer zipFileReader.Close()

		destFile, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, zipFileReader); err != nil {
			return fmt.Errorf("failed to copy file %s from archive to local file %s: %w", zipFile.Name, destFilePath, err)
		}
	}

	return nil
}

func checkLicenseAcceptance(pluginManifest *cpstructs.Manifest, prompt form.Prompt, force bool) error {
	for _, license := range pluginManifest.Licenses {
		if force {
			output.Printf(false, "Implicitly agreeing to the following license: %s (%s)\n", license.Name, license.Url)
		} else {
			f := form.New(form.Field{
				ID:        "confirm",
				Prompt:    fmt.Sprintf("License: %s (%s)\nI agree to this software license agreement.", license.Name, license.Url),
				IsYesOrNo: true,
			})
			if err := f.Prompt(prompt); err != nil {
				return err
			}
			if !f.Responses["confirm"].(bool) {
				return fmt.Errorf("you must accept all license agreements to install this plugin")
			}
		}
	}
	output.Println(false, "")

	return nil
}

func updateWorkerConfigs(pluginDir string, workerConfigs []string, dryRun bool) error {
	if len(workerConfigs) > 0 {
		updateWorkerMsg := "Adding plugin installation directory to the plugin path in the following files:"
		if dryRun {
			updateWorkerMsg = utils.AddDryRunPrefix(updateWorkerMsg)
		}
		for _, workerConfig := range workerConfigs {
			updateWorkerMsg = fmt.Sprintf("%s\n\t* %v", updateWorkerMsg, workerConfig)
		}
		output.Println(false, updateWorkerMsg)
	}

	for _, workerConfig := range workerConfigs {
		if err := updateWorkerConfig(pluginDir, workerConfig, dryRun); err != nil {
			return err
		}
	}
	return nil
}

func (c *pluginInstallCommand) GetHubClient() (*hub.Client, error) {
	unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, err
	}

	return hub.NewClient(c.Config.Version.UserAgent, c.Config.IsTest, unsafeTrace), nil
}
