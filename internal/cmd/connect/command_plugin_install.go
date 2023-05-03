package connect

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Owner struct {
		Username string `json:"username"`
	} `json:"owner"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Archive struct {
		Url  string `json:"url"`
		Md5  string `json:"md5"`
		Sha1 string `json:"sha1"`
	} `json:"archive"`
	License []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"license"`
}

const (
	archiveInstallation = "ARCHIVE"
	packageInstallation = "PACKAGE"

	hubPluginsUrl = "https://api.hub.confluent.io/api/plugins"
)

func (c *pluginCommand) newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install a Confluent Hub plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.install,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Install the latest release of the Datagen connector onto your local Confluent Platform environment.",
				Code: "confluent connect plugin install confluentinc/kafka-connect-datagen:latest",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().String("component-dir", "", "The plugin installation directory. Defaults to $share/confluent-hub-components for archive deployment, and to /usr/share/confluent-hub-components for deb/rpm deployment.")
	cmd.Flags().StringSlice("worker-configs", []string{}, "Comma-delineated list of paths to one or more Kafka Connect worker configuration files. Each worker file will be updated to load plugins from the component directory in addition to any preexisting directories.")
	cmd.Flags().Bool("dry-run", false, "Simulate an operation without making any changes.")
	cmd.Flags().Bool("no-prompt", false, "Proceed without asking for user input.")

	return cmd
}

func (c *pluginCommand) install(cmd *cobra.Command, args []string) error {
	// Verify that the argument corresponds to a valid plugin
	var pluginManifest *manifest
	if utils.DoesPathExist(args[0]) {
		// if installing plugin from local archive
		localManifest, err := getLocalManifest(args[0])
		if err != nil {
			return err
		}
		pluginManifest = localManifest
	} else {
		// if installing plugin from Confluent Hub
		owner, name, version, err := parsePluginId(args[0])
		if err != nil {
			return err
		}

		remoteManifest, err := getRemoteManifest(owner, name, version)
		if err != nil {
			return err
		}
		pluginManifest = remoteManifest
	}

	// Select component-dir
	var ins *installation
	noPrompt, err := cmd.Flags().GetBool("no-prompt")
	if err != nil {
		return err
	}

	componentDir, err := getComponentDirFromFlag(cmd)
	if err != nil {
		return err
	}

	if componentDir == "" {
		ins, err = getInstallation(cmd, noPrompt)
		if err != nil {
			return err
		}
		componentDir, err = chooseComponentDir(ins, noPrompt)
		if err != nil {
			return err
		}
	}

	// Check for, and possibly remove, existing installation
	if previousInstallations, err := existingPluginInstallation(componentDir, pluginManifest); err != nil {
		return err
	} else if len(previousInstallations) > 0 {
		output.Println("A version of this component is already installed.")
		for _, previousInstallation := range previousInstallations {
			if err := uninstall(previousInstallation, noPrompt); err != nil {
				return err
			}
		}
	}

	// Install
	if err := checkLicenseAcceptance(pluginManifest, noPrompt); err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	if dryRun {
		output.Println("Skipping installation of connector onto worker as part of dry run mode.")
	} else {
		if utils.DoesPathExist(args[0]) {
			if err := installFromLocal(pluginManifest, args[0], componentDir); err != nil {
				return err
			}
		} else {
			if err := installFromRemote(pluginManifest, componentDir); err != nil {
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
		if ins == nil {
			ins, err = getInstallation(cmd, noPrompt)
			if err != nil {
				return err
			}
		}
		workerConfigs, err = chooseWorkerConfigs(cmd, ins, noPrompt)
		if err != nil {
			return err
		}
	}

	for _, workerConfig := range workerConfigs {
		output.Printf("Updating worker config at %s\n", workerConfig)
		updateWorkerConfig(componentDir, workerConfig, dryRun)
	}

	return nil
}

func parsePluginId(plugin string) (string, string, string, error) {
	parsePluginErrorMsg := "component not found"
	parsePluginSuggestions := "Provide either a path to a local file or a plugin id from Confluent Hub with the format: `<owner>/<name>:<version>`."

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

func getLocalManifest(archivePath string) (*manifest, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, errors.Errorf("failed to open local archive file %s: %v", archivePath, err)
	}
	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		isManifest, err := filepath.Match("*/manifest.json", filepath.ToSlash(zipFile.Name))
		if err != nil {
			return nil, errors.Errorf("failed to examine file %s inside local archive file %s: %v", zipFile.Name, archivePath, err)
		}

		if isManifest {
			manifestFile, err := zipFile.Open()
			if err != nil {
				return nil, errors.Errorf("failed to open manifest file %s inside local archive file %s: %v", zipFile.Name, archivePath, err)
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

	return nil, errors.Errorf("failed to find manifest file inside local archive file %s", archivePath)
}

func getRemoteManifest(owner, name, version string) (*manifest, error) {
	manifestUrl := fmt.Sprintf("%s/%s/%s", hubPluginsUrl, owner, name)
	if version != "latest" {
		manifestUrl = fmt.Sprintf("%s/versions/%s", manifestUrl, version)
	}

	r, err := http.Get(manifestUrl)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("failed to read manifest file from Confuent Hub")
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	pluginManifest := new(manifest)
	if err := json.Unmarshal(body, &pluginManifest); err != nil {
		return nil, err
	}

	return pluginManifest, nil
}

func getComponentDirFromFlag(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Changed("component-dir") {
		componentDir, err := cmd.Flags().GetString("component-dir")
		if err != nil {
			return "", err
		}

		if componentDir, err = filepath.Abs(componentDir); err != nil {
			return "", errors.Errorf(`failed to determine absolute path to component directory "%s": %v`, componentDir, err)
		}

		if !utils.DoesPathExist(componentDir) {
			return "", errors.Errorf(`component directory "%s" does not exist`, componentDir)
		}

		return componentDir, nil
	}

	return "", nil
}

func existingPluginInstallation(componentDir string, pluginManifest *manifest) ([]string, error) {
	// Bundled installations
	if utils.DoesPathExist(filepath.Join(componentDir, pluginManifest.Name)) {
		return nil, errors.New("unable to install component because it's already bundled")
	}

	// Other previous installations
	immediateDirectory := filepath.Join(componentDir, fmt.Sprintf("%s-%s", pluginManifest.Owner.Username, pluginManifest.Name))
	uberJar := filepath.Join(componentDir, fmt.Sprintf("%s-%s.jar", pluginManifest.Name, pluginManifest.Version))

	var installations []string
	if utils.DoesPathExist(immediateDirectory) {
		installations = append(installations, immediateDirectory)
	}
	if utils.DoesPathExist(uberJar) {
		installations = append(installations, uberJar)
	}

	return installations, nil
}

func uninstall(pathToComponent string, noPrompt bool) error {
	if noPrompt {
		fmt.Printf("Automatically uninstalling existing version of the component located at %s", pathToComponent)
	} else {
		f := form.New(form.Field{
			ID:        "confirm",
			Prompt:    fmt.Sprintf("Do you want to uninstall an existing version of this component located at %s?", pathToComponent),
			IsYesOrNo: true,
		})
		if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
			return err
		}
		if !f.Responses["confirm"].(bool) {
			return errors.New("previous versions must be uninstalled to continue")
		}
	}
	return os.RemoveAll(pathToComponent)
}

func installFromLocal(pluginManifest *manifest, archivePath, componentDir string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return errors.Errorf("failed to open local archive file %s: %v", archivePath, err)
	}
	defer zipReader.Close()

	return unzipPlugin(pluginManifest, zipReader.File, componentDir)
}

func installFromRemote(pluginManifest *manifest, componentDir string) error {
	r, err := http.Get(pluginManifest.Archive.Url)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.New("failed to retrieve archive from Confuent Hub")
	}
	defer r.Body.Close()

	archive, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	checksumErrorMsg := "%s checksum for downloaded archive (%s) does not match checksum in manifest (%s) for component %s"
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
		return errors.Errorf("failed to open remote archive file %s: %v", archive, err)
	}

	return unzipPlugin(pluginManifest, zipReader.File, componentDir)
}

func unzipPlugin(pluginManifest *manifest, zipFiles []*zip.File, componentDir string) error {
	relativeInstallationDir := filepath.Join(componentDir, fmt.Sprintf("%s-%s", pluginManifest.Owner.Username, pluginManifest.Name))
	installationDir, err := filepath.Abs(relativeInstallationDir)
	if err != nil {
		return errors.Errorf("failed to resolve absolute path for directory %s: %v", relativeInstallationDir, err)
	}

	for _, zipFile := range zipFiles {
		versionPrefix := fmt.Sprintf("%s-%s-%s", pluginManifest.Owner.Username, pluginManifest.Name, pluginManifest.Version)
		destFilePath := filepath.Join(installationDir, strings.TrimPrefix(zipFile.Name, versionPrefix))

		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return errors.Errorf("failed to create directory %s on local storage: %v", destFilePath, err)
			}
			continue
		} else {
			if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
				return errors.Errorf("failed to create directory %s on local storage: %v", filepath.Dir(destFilePath), err)
			}
		}

		zipFileReader, err := zipFile.Open()
		if err != nil {
			return errors.Errorf("failed to read file %s from archive: %v", zipFile.Name, err)
		}
		defer zipFileReader.Close()

		destFile, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, zipFileReader); err != nil {
			return errors.Errorf("failed to copy file %s from archive to local file %s: %v", zipFile.Name, destFilePath, err)
		}
	}

	return nil
}

func checkLicenseAcceptance(pluginManifest *manifest, noPrompt bool) error {
	for _, license := range pluginManifest.License {
		if noPrompt {
			output.Printf("Implicitly agreeing to license:\n%s\n%s\n", license.Name, license.Url)
		} else {
			f := form.New(form.Field{
				ID:        "confirm",
				Prompt:    fmt.Sprintf("License:\n%s\n%s\nI agree to this software license agreement. ", license.Name, license.Url),
				IsYesOrNo: true,
			})
			if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
				return err
			}
			if !f.Responses["confirm"].(bool) {
				return errors.New("you must accept all license agreements for a component in order to install it")
			}
		}
	}

	return nil
}
