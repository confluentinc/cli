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
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type manifest struct {
	Name    string    `json:"name"`
	Title   string    `json:"title"`
	Version string    `json:"version"`
	Owner   owner     `json:"owner"`
	Archive archive   `json:"archive"`
	License []license `json:"license"`
}

type owner struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type archive struct {
	Url  string `json:"url"`
	Md5  string `json:"md5"`
	Sha1 string `json:"sha1"`
}

type license struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

const (
	invalidDirectoryErrorMsg       = `plugin directory "%s" does not exist`
	unexpectedInstallationErrorMsg = "unexpected installation type: %s"
)

func (c *pluginCommand) newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install a Connect plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.install,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Install the latest release of the Datagen connector onto your local Confluent Platform environment.",
				Code: "confluent connect plugin install confluentinc/kafka-connect-datagen:latest",
			},
			examples.Example{
				Text: "Install the latest release of the Datagen connector in a user specified directory and update one worker-config file.",
				Code: "confluent connect plugin install confluentinc/kafka-connect-datagen:latest --plugin-directory $CONFLUENT_HOME/plugins --worker-configs $CONFLUENT_HOME/etc/kafka/connect-distributed.properties",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().String("plugin-directory", "", "The plugin installation directory. If not specified, a default will be selected based on your Confluent Platform installation.")
	cmd.Flags().StringSlice("worker-configs", []string{}, "A comma-separated list of paths to one or more Kafka Connect worker configuration files. Each worker file will be updated to load plugins from the plugin directory in addition to any preexisting directories.")
	pcmd.AddDryRunFlag(cmd)
	cmd.Flags().Bool("force", false, "Proceed without user input.")

	cobra.CheckErr(cmd.MarkFlagDirname("plugin-directory"))

	return cmd
}

func (c *pluginCommand) install(cmd *cobra.Command, args []string) error {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	if dryRun {
		output.Println("[DRY RUN] Performing a dry run of this command.")
	}

	pluginManifest, err := getManifest(args[0])
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Select plugin-directory
	pluginDir, err := getPluginDirFromFlag(cmd)
	if err != nil {
		return err
	}

	var installation *platformInstallation
	if pluginDir == "" {
		installation, err = getConfluentPlatformInstallation(cmd, force)
		if err != nil {
			return err
		}
		pluginDir, err = choosePluginDir(installation, force)
		if err != nil {
			return err
		}
	}

	// Check for, and possibly remove, existing installation
	if previousInstallations, err := existingPluginInstallation(pluginDir, pluginManifest); err != nil {
		return err
	} else if len(previousInstallations) > 0 {
		output.Println("\nA version of this plugin is already installed and must be removed to continue.")
		for _, previousInstallation := range previousInstallations {
			if err := replacePluginInstallation(previousInstallation, dryRun, force); err != nil {
				return err
			}
		}
	}

	// Install
	if err := checkLicenseAcceptance(pluginManifest, force); err != nil {
		return err
	}

	output.Printf("\nInstalling %s %s, provided by %s\n", pluginManifest.Title, pluginManifest.Version, pluginManifest.Owner.Name)
	if dryRun {
		output.Println("[DRY RUN] Skipping plugin installation.")
	} else {
		if utils.DoesPathExist(args[0]) {
			if err := installFromLocal(pluginManifest, args[0], pluginDir); err != nil {
				return err
			}
		} else {
			if err := installFromRemote(pluginManifest, pluginDir); err != nil {
				return err
			}
		}
	}

	// Select and update worker-configs
	workerConfigs, err := cmd.Flags().GetStringSlice("worker-configs")
	if err != nil {
		return err
	}

	if len(workerConfigs) == 0 {
		if installation == nil {
			installation, err = getConfluentPlatformInstallation(cmd, force)
			if err != nil {
				return err
			}
		}
		workerConfigs, err = chooseWorkerConfigs(cmd, installation, force)
		if err != nil {
			return err
		}
	}

	if len(workerConfigs) > 0 {
		updateWorkerMsg := "\nAdding plugin installation directory to the plugin path in the following files:"
		for _, workerConfig := range workerConfigs {
			updateWorkerMsg = fmt.Sprintf("%s\n\t* %v", updateWorkerMsg, workerConfig)
		}
		output.Println(updateWorkerMsg)
	}

	for _, workerConfig := range workerConfigs {
		if err := updateWorkerConfig(pluginDir, workerConfig, dryRun); err != nil {
			return err
		}
	}

	if dryRun {
		output.Println("\n[DRY RUN] Success. All requirements are met and you may proceed with the installation.")
	}

	return nil
}

func parsePluginId(plugin string) (string, string, string, error) {
	parsePluginErrorMsg := "plugin not found"
	parsePluginSuggestions := "Provide a path to a local file or provide a plugin ID from Confluent Hub with the format: `<owner>/<name>:<version>`."

	ownerNameSplit := strings.Split(plugin, "/")
	if len(ownerNameSplit) != 2 || ownerNameSplit[0] == "" {
		return "", "", "", errors.NewErrorWithSuggestions(parsePluginErrorMsg, parsePluginSuggestions)
	}
	nameVersionSplit := strings.Split(ownerNameSplit[1], ":")
	if len(nameVersionSplit) != 2 || nameVersionSplit[0] == "" || nameVersionSplit[1] == "" {
		return "", "", "", errors.NewErrorWithSuggestions(parsePluginErrorMsg, parsePluginSuggestions)
	}

	return ownerNameSplit[0], nameVersionSplit[0], nameVersionSplit[1], nil
}

func getManifest(id string) (*manifest, error) {
	if utils.DoesPathExist(id) {
		// if installing plugin from local archive
		return getLocalManifest(id)
	} else {
		// if installing plugin from Confluent Hub
		owner, name, version, err := parsePluginId(id)
		if err != nil {
			return nil, err
		}

		remoteManifest, err := getRemoteManifest(owner, name, version)
		if err != nil {
			return nil, err
		}
		return remoteManifest, nil
	}
}

func getLocalManifest(archivePath string) (*manifest, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open local archive file %s", archivePath)
	}
	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		isManifest, err := filepath.Match("*/manifest.json", filepath.ToSlash(zipFile.Name))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to examine file %s inside local archive file %s", zipFile.Name, archivePath)
		}

		if isManifest {
			manifestFile, err := zipFile.Open()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to open manifest file %s inside local archive file %s", zipFile.Name, archivePath)
			}
			defer manifestFile.Close()

			jsonByteArr, err := io.ReadAll(manifestFile)
			if err != nil {
				return nil, err
			}

			pluginManifest := new(manifest)
			if err := json.Unmarshal(jsonByteArr, &pluginManifest); err != nil {
				return nil, err
			}

			return pluginManifest, nil
		}
	}

	return nil, errors.Errorf(`failed to find manifest file inside local archive file "%s"`, archivePath)
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
		return "", errors.Errorf(invalidDirectoryErrorMsg, pluginDir)
	}

	return pluginDir, nil
}

func existingPluginInstallation(pluginDir string, pluginManifest *manifest) ([]string, error) {
	// Bundled installations
	if utils.DoesPathExist(filepath.Join(pluginDir, pluginManifest.Name)) {
		return nil, errors.New("unable to install plugin because it is already bundled")
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

func replacePluginInstallation(pathToPlugin string, dryRun, force bool) error {
	if force {
		output.Printf("Uninstalling the existing version of the plugin located at \"%s\".\n", pathToPlugin)
	} else {
		f := form.New(form.Field{
			ID:        "confirm",
			Prompt:    fmt.Sprintf("Do you want to uninstall an existing version of this plugin located at %s?", pathToPlugin),
			IsYesOrNo: true,
		})
		if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
			return err
		}
		if !f.Responses["confirm"].(bool) {
			return errors.New("previous versions must be uninstalled to continue")
		}
	}

	err := os.RemoveAll(pathToPlugin)
	if err != nil {
		return err
	}

	if dryRun {
		output.Println("[DRY RUN] Success.")
		return nil
	}

	output.Println("Success.")
	return nil
}

func installFromLocal(pluginManifest *manifest, archivePath, pluginDir string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open local archive file %s", archivePath)
	}
	defer zipReader.Close()

	return unzipPlugin(pluginManifest, zipReader.File, pluginDir)
}

func installFromRemote(pluginManifest *manifest, pluginDir string) error {
	archive, err := getRemoteArchive(pluginManifest)
	if err != nil {
		return err
	}

	checksumErrorMsg := "%s checksum for downloaded archive (%s) does not match checksum in manifest (%s) for plugin %s"
	calculatedMd5Checksum := fmt.Sprintf("%x", md5.Sum(archive))
	if calculatedMd5Checksum != pluginManifest.Archive.Md5 {
		return errors.Errorf(checksumErrorMsg, "md5", calculatedMd5Checksum, pluginManifest.Archive.Md5, pluginManifest.Name)
	}
	calculatedSha1Checksum := fmt.Sprintf("%x", sha1.Sum(archive))
	if calculatedSha1Checksum != pluginManifest.Archive.Sha1 {
		return errors.Errorf(checksumErrorMsg, "sha1", calculatedSha1Checksum, pluginManifest.Archive.Sha1, pluginManifest.Name)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return errors.Wrapf(err, "failed to open remote archive file %s", archive)
	}

	return unzipPlugin(pluginManifest, zipReader.File, pluginDir)
}

func unzipPlugin(pluginManifest *manifest, zipFiles []*zip.File, pluginDir string) error {
	relativeInstallationDir := filepath.Join(pluginDir, fmt.Sprintf("%s-%s", pluginManifest.Owner.Username, pluginManifest.Name))
	installationDir, err := filepath.Abs(relativeInstallationDir)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve absolute path for directory %s", relativeInstallationDir)
	}

	for _, zipFile := range zipFiles {
		versionPrefix := fmt.Sprintf("%s-%s-%s", pluginManifest.Owner.Username, pluginManifest.Name, pluginManifest.Version)
		destFilePath := filepath.Join(installationDir, strings.TrimPrefix(zipFile.Name, versionPrefix))

		createDirectoryErrorMsg := "failed to create directory %s on local storage"
		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return errors.Wrapf(err, createDirectoryErrorMsg, destFilePath)
			}
			continue
		} else {
			if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
				return errors.Wrapf(err, createDirectoryErrorMsg, filepath.Dir(destFilePath))
			}
		}

		zipFileReader, err := zipFile.Open()
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s from archive", zipFile.Name)
		}
		defer zipFileReader.Close()

		destFile, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, zipFileReader); err != nil {
			return errors.Wrapf(err, "failed to copy file %s from archive to local file %s", zipFile.Name, destFilePath)
		}
	}

	return nil
}

func checkLicenseAcceptance(pluginManifest *manifest, force bool) error {
	for _, license := range pluginManifest.License {
		if force {
			output.Printf("\nImplicitly agreeing to the following license:\n%s\n%s\n", license.Name, license.Url)
		} else {
			f := form.New(form.Field{
				ID:        "confirm",
				Prompt:    fmt.Sprintf("\nLicense:\n%s\n%s\nI agree to this software license agreement.", license.Name, license.Url),
				IsYesOrNo: true,
			})
			if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
				return err
			}
			if !f.Responses["confirm"].(bool) {
				return errors.New("you must accept all license agreements to install this plugin")
			}
		}
	}

	return nil
}
