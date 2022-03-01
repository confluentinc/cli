package chub

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/properties"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/examples"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

const (
	archiveInstallation = "ARCHIVE"
	packageInstallation = "PACKAGE"

	standardDistributedConfigLocation       = "/etc/kafka/connect-distributed.properties"
	standardStandaloneConfigLocation        = "/etc/kafka/connect-standalone.properties"
	schemaRegistryDistributedConfigLocation = "/etc/schema-registry/connect-avro-distributed.properties"
	schemaRegistryStandaloneConfigLocation  = "/etc/schema-registry/connect-avro-standalone.properties"
	confluentCurrentConfigLocation          = "confluent.current"
	connectCurrentConfigLocation            = "/connect/connect.properties"

	componentDirArchiveDefault = "share/confluent-hub-components"
	componentDirPackageDefault = "/usr/share/confluent-hub-components"

	confluentCurrentEnvVar  = "CONFLUENT_CURRENT"
	confluentHomeEnvVar     = "CONFLUENT_HOME"
	tmpDirEnvVar            = "TMPDIR"
	confluentCommonLocation = "share/java/confluent-common"
	connectDistributedPath  = "/usr/bin/connect-distributed"

	confluentHubBackEnd = "https://api.hub.confluent.io"

	pluginPathProperty = "plugin.path"

	daemonOption = "-daemon"

	connectProcessRegex = `org\.apache\.kafka\.connect\.cli\.Connect(Distributed|Standalone)`
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hub",
		Short: "Confluent Hub.",
		Args:  cobra.NoArgs,
	}
	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
	}
	c.init()

	return c.Command
}

func (c *command) init() {
	installCmd := &cobra.Command{
		Use:   "install <id>",
		Short: "Install a Confluent Hub component.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.install),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Install the latest release of the Datagen connector onto your local Confluent Platform environment, deducing the installation directory from your $CONFLUENT_HOME environment variable",
				Code: fmt.Sprintf("confluent hub install confluentinc/kafka-connect-datagen:latest"),
			},
		),
	}
	installCmd.Flags().String("component-dir", "", "The local directory into which components are installed. Defaults to $share/confluent-hub-components when running the client in an archive installation of Confluent Platform, and to /usr/share/confluent-hub-components when running the client in a deb/rpm installation.\n\nThis options value must be a path that must exist on the file system. The provided path must be readable and writable.")
	installCmd.Flags().String("worker-configs", "", "Path(s) to your Kafka Connect worker config file(s). Multiple paths can be delimited using the colon (':') character. If provided, each worker file will be updated to load plugins from the component directory (in addition to any other directories it is already configured to read)")
	installCmd.Flags().Bool("dry-run", false, "Dry run mode; no files will be installed onto your Confluent Platform environment")
	installCmd.Flags().Bool("no-prompt", false, "Force installation without interactively confirming details such as license and component directory")
	// Already supported by top-level CLI logic. We may want to preserve this flag here just to ease migration from the existing confluent-hub tool
	// installCmd.Flags().Boolean("verbose", false, "Log extra information about the installation as it takes place")
	c.AddCommand(installCmd)

	// TODO:
	//   - Add search command (show components available on Confluent Hub)
	//   - Add list command (show components already present on worker)

}

func (c *command) install(cmd *cobra.Command, args []string) error {
	componentDirRaw, err := cmd.Flags().GetString("component-dir")
	if err != nil {
		return err
	}
	noPrompt, err := cmd.Flags().GetBool("no-prompt")
	if err != nil {
		return err
	}
	prompt := !noPrompt
	input := bufio.NewReader(os.Stdin)
	var selectedInstallation *Installation
	var componentDir string
	if componentDirRaw == "" {
		selectedInstallation, err = pickInstallation(!noPrompt, input)
		if err != nil {
			return err
		}
		componentDir, err = chooseComponentDir(selectedInstallation, prompt, input)
		if err != nil {
			return err
		}
	} else {
		if componentDir, err = filepath.Abs(componentDirRaw); err != nil {
			return fmt.Errorf(
				"failed to determine absolute path to component directory %s: %w",
				componentDirRaw,
				err,
			)
		}
	}
	workerConfigsRaw, err := cmd.Flags().GetString("worker-configs")
	if err != nil {
		return err
	}
	var workerConfigs []string
	if workerConfigsRaw == "" {
		if selectedInstallation == nil {
			selectedInstallation, err = pickInstallation(!noPrompt, input)
			if err != nil {
				return err
			}
		}
		workerConfigs, err = chooseWorkerConfigs(selectedInstallation, prompt, input)
		if err != nil {
			return err
		}
	} else {
		workerConfigs = strings.Split(workerConfigsRaw, ":")
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	componentName := args[0]
	var component *Component
	var verifyChecksum bool
	if exists, err := pathExists(componentName); err == nil {
		if exists {
			if component, err = openLocalArchive(componentName); err != nil {
				return err
			}
			// We only verify checksums for files downloaded from Confluent Hub
			verifyChecksum = false
		} else {
			id, err := parseComponentId(componentName)
			if err != nil {
				return err
			}
			if component, err = downloadArchive(id); err != nil {
				return err
			}
			verifyChecksum = true
		}
	} else {
		return err
	}
	defer component.ArchiveContent.Close()
	if err := checkIfAlreadyInstalled(component.Id, componentDir, prompt, input); err != nil {
		return err
	}
	if err := installArchive(component, componentDir, verifyChecksum, prompt, input, dryRun); err != nil {
		return err
	}
	for _, workerConfig := range workerConfigs {
		if err := updateWorkerConfig(componentDir, workerConfig, dryRun); err != nil {
			return err
		}
	}
	return nil
}

func pickInstallation(prompt bool, input *bufio.Reader) (*Installation, error) {
	installations, err := inferInstallationDirectories()
	if err != nil {
		return nil, err
	}
	if len(installations) == 0 {
		return nil, errors.New("unable to detect Confluent Platform installation. Specify --component-dir and --worker-configs explicitly")
	}
	if len(installations) == 1 || !prompt {
		if len(installations) == 1 {
			fmt.Printf("Using the only available installation: %s\n", installations[0])
		}
		return &installations[0], nil
	}
	// There's more than one installation and we're running interactively;
	// make the user choose which installation they'd like to work with
	fmt.Println("The component can be installed in any of the following Confluent Platform installations:")
	for i, installation := range installations {
		fmt.Printf("  %d. %s (%s)\n", i+1, installation.Path, installation.Use)
	}
	var choice int
	for {
		fmt.Printf("Choose one of these to continue the installation (%d-%d): ", 1, len(installations))
		line, err := readLine(input)
		if err != nil {
			return nil, fmt.Errorf("could not read response from terminal: %w", err)
		}
		choice, err = strconv.Atoi(line)
		if err != nil {
			// Ignore, and just ask again
			continue
		} else if choice < 1 || choice > len(installations) {
			fmt.Printf("Number must be in the range %d to %d, inclusive\n", 1, len(installations))
		} else {
			break
		}
	}

	return &installations[choice-1], nil
}

// Return value maps Installation -> list of uses for that installation
func inferInstallationDirectories() ([]Installation, error) {
	// Check in descending order of precedence:
	//   - $CONFLUENT_HOME
	//   - current directory
	//   - standard rpm/deb
	//   - based on the client

	var result []Installation
	confluentHome := os.Getenv(confluentHomeEnvVar)
	if confluentHome != "" {
		if exists, err := hasArchiveInstallation(confluentHome); exists && err == nil {
			installation := Installation{
				Type: archiveInstallation,
				Path: confluentHome,
				Use:  "based on $CONFLUENT_HOME",
			}
			result = append(result, installation)
		} else if err != nil {
			return nil, fmt.Errorf(
				"failed while checking $CONLFUENT_HOME directory for archive-style installation layout: %w",
				err,
			)
		}
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to determine current working directory: %w", err)
	}
	if exists, err := hasArchiveInstallation(currentDirectory); exists && err == nil {
		installation := Installation{
			Type: archiveInstallation,
			Path: currentDirectory,
			Use:  "found in the current directory",
		}
		result = append(result, installation)
	} else if err != nil {
		return nil, fmt.Errorf(
			"failed while checking current working directory for archive-style installation layout: %w",
			err,
		)
	}

	if exists, err := hasPackageInstallation(); exists && err == nil {
		installation := Installation{
			Type: packageInstallation,
			Path: filepath.FromSlash("/"),
			Use:  "installed rpm/deb package",
		}
		result = append(result, installation)
	} else if err != nil {
		return nil, fmt.Errorf(
			"failed while checking root directory for package-style installation layout: %w",
			err,
		)
	}

	cliPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to determine path to CLI: %w", err)
	}
	cliDirectory := filepath.Dir(cliPath)
	cliUse := "where this tool is installed"
	if exists, err := hasArchiveInstallation(cliDirectory); exists && err == nil {
		installation := Installation{
			Type: archiveInstallation,
			Path: cliDirectory,
			Use:  cliUse,
		}
		result = append(result, installation)
	} else if err != nil {
		return nil, fmt.Errorf(
			"failed while checking CLI directory for archive-style installation layout: %w",
			err,
		)
	} else if filepath.ToSlash(cliDirectory) == "/usr/bin" {
		if exists, err := hasPackageInstallation(); exists && err == nil {
			installation := Installation{
				Type: packageInstallation,
				Path: filepath.FromSlash("/"),
				Use:  cliUse,
			}
			result = append(result, installation)
		} else if err != nil {
			return nil, fmt.Errorf(
				"failed while checking root directory for package-style installation layout: %w",
				err,
			)
		}
	}

	return result, nil
}

func hasArchiveInstallation(dir string) (bool, error) {
	slashed := filepath.ToSlash(dir)
	if slashed == "/usr" {
		return false, nil
	}
	confluentCommonPath := filepath.Join(dir, filepath.FromSlash(confluentCommonLocation))
	return pathExists(confluentCommonPath)
}

func hasPackageInstallation() (bool, error) {
	return pathExists(filepath.FromSlash(connectDistributedPath))
}

func readLine(reader *bufio.Reader) (string, error) {
	result, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(result, "\n"), nil
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf(
			"failed to check for existence of file/directory %s: %w",
			path,
			err,
		)
	}
}

func chooseComponentDir(installation *Installation, prompt bool, input *bufio.Reader) (string, error) {
	defaultComponentDir, err := defaultComponentDir(installation)
	if err != nil {
		return "", err
	}
	if prompt {
		for /*ever*/ {
			fmt.Printf("Do you want to install this into %s? (yn) ", defaultComponentDir)
			line, err := readLine(input)
			if err != nil {
				return "", fmt.Errorf("could not read response from terminal: %w", err)
			} else if strings.EqualFold("y", line) {
				return defaultComponentDir, nil
			} else if strings.EqualFold("n", line) {
				break
			}
		}
		fmt.Print("Specify installation directory: ")
		line, err := readLine(input)
		if err != nil {
			return "", fmt.Errorf("failed to read response from terminal: %w", err)
		}
		result, err := filepath.Abs(line)
		if err != nil {
			return "", fmt.Errorf("failed to determine absolute path to component directory: %w", err)
		}
		return result, nil
	} else {
		return defaultComponentDir, nil
	}
}

func defaultComponentDir(installation *Installation) (string, error) {
	switch installation.Type {
	case archiveInstallation:
		{
			return filepath.Join(installation.Path, componentDirArchiveDefault), nil
		}
	case packageInstallation:
		{
			return componentDirPackageDefault, nil
		}
	default:
		{
			return "", errors.New(fmt.Sprintf("unexpected installation type: %s", installation.Type))
		}
	}
}

func chooseWorkerConfigs(installation *Installation, prompt bool, input *bufio.Reader) ([]string, error) {
	var workerConfigs []WorkerConfig

	standardWorkerConfigs, err := standardWorkerConfigLocations(installation)
	if err != nil {
		return nil, fmt.Errorf("could not infer possible worker config locations from standard candidates: %w", err)
	}
	for _, workerConfig := range standardWorkerConfigs {
		if exists, err := pathExists(workerConfig.Path); exists && err == nil {
			workerConfigs = append(workerConfigs, workerConfig)
		} else if err != nil {
			return nil, fmt.Errorf("could not infer worker config locations from standard candidates: %w", err)
		}
	}

	runningWorkerConfigs, err := runningWorkerConfigLocations()
	if err != nil {
		return nil, fmt.Errorf("could not infer possible worker config locations from running processes: %w", err)
	}
	for _, workerConfig := range runningWorkerConfigs {
		if exists, err := pathExists(workerConfig.Path); exists && err == nil {
			workerConfigs = append(workerConfigs, workerConfig)
		} else if err != nil {
			return nil, fmt.Errorf("could not infer worker config locations from running processes: %w", err)
		}
	}

	if len(workerConfigs) == 0 {
		// No need to choose from literally zero options
		return []string{}, nil
	}

	var filteredWorkerConfigs []WorkerConfig
	if prompt {
		fmt.Println("Detected Worker's configs:")
		for i, workerConfig := range workerConfigs {
			fmt.Printf("  %d. %s (%s)\n", i+1, workerConfig.Path, workerConfig.Use)
		}
		for /*ever*/ {
			fmt.Print("Do you want to update all detected configs? (yn) ")
			line, err := readLine(input)
			if err != nil {
				return nil, fmt.Errorf("could not read response from terminal: %w", err)
			} else if strings.EqualFold("y", line) {
				filteredWorkerConfigs = workerConfigs
				break
			} else if strings.EqualFold("n", line) {
				filteredWorkerConfigs, err = chooseIndividualWorkerConfigs(workerConfigs, input)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	} else {
		filteredWorkerConfigs = workerConfigs
	}

	var result []string
	for _, workerConfig := range filteredWorkerConfigs {
		result = append(result, workerConfig.Path)
	}
	return result, nil
}

func chooseIndividualWorkerConfigs(workerConfigs []WorkerConfig, input *bufio.Reader) ([]WorkerConfig, error) {
	var result []WorkerConfig
	for i, workerConfig := range workerConfigs {
		for /*ever*/ {
			fmt.Printf("Do you want to update %d? (yn) ", i+1)
			line, err := readLine(input)
			if err != nil {
				return nil, fmt.Errorf("could not read response from terminal: %w", err)
			} else if strings.EqualFold("y", line) {
				result = append(result, workerConfig)
				break
			} else if strings.EqualFold("n", line) {
				break
			}

		}
	}
	return result, nil
}

func standardWorkerConfigLocations(installation *Installation) ([]WorkerConfig, error) {
	workerConfigLocations := []string{
		standardDistributedConfigLocation, standardStandaloneConfigLocation,
		schemaRegistryDistributedConfigLocation, schemaRegistryStandaloneConfigLocation,
	}
	switch installation.Type {
	case archiveInstallation:
		{
			var result []WorkerConfig
			for _, workerConfigLocation := range workerConfigLocations {
				workerConfigPath := filepath.Join(installation.Path, filepath.FromSlash(workerConfigLocation))
				workerConfig := WorkerConfig{Path: workerConfigPath, Use: "Standard"}
				result = append(result, workerConfig)
			}
			confluentCurrentDir := os.Getenv(confluentCurrentEnvVar)
			if confluentCurrentDir == "" {
				confluentCurrentDir = os.Getenv(tmpDirEnvVar)
			}
			confluentCurrentFile := filepath.Join(confluentCurrentDir, confluentCurrentConfigLocation)
			if exists, err := pathExists(confluentCurrentFile); exists && err == nil {
				confluentCurrentContent, err := os.ReadFile(confluentCurrentFile)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to read possible CONFLUENT_CURRENT file %s: %w",
						confluentCurrentFile,
						err,
					)
				}
				confluentCurrentLines := strings.SplitN(string(confluentCurrentContent), "\n", 3)
				if len(confluentCurrentLines) == 1 {
					connectCurrentConfigFile := filepath.Join(confluentCurrentLines[0], connectCurrentConfigLocation)
					workerConfig := WorkerConfig{Path: connectCurrentConfigFile, Use: "Based on " + confluentCurrentEnvVar}
					result = append(result, workerConfig)
				}
			} else if err != nil {
				return nil, fmt.Errorf(
					"failed to check for existence of possible CONFLUENT_CURRENT file %s: %w",
					confluentCurrentFile,
					err,
				)
			}
			return result, nil
		}
	case packageInstallation:
		{
			var result []WorkerConfig
			for _, workerConfigLocation := range workerConfigLocations {
				workerConfig := WorkerConfig{Path: filepath.FromSlash(workerConfigLocation), Use: "Standard"}
				result = append(result, workerConfig)
			}
			return result, nil
		}
	default:
		{
			return nil, errors.New(fmt.Sprintf("unexpected installation type: %s", installation.Type))
		}
	}
}

func runningWorkerConfigLocations() ([]WorkerConfig, error) {
	command := `ps ax |
					grep -E '` + connectProcessRegex + `'|
					grep -v grep;
				test ${PIPESTATUS[0]} -eq 0`
	out, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run shell command to locate running Connect worker processes: %w", err)
	}
	var result []WorkerConfig
	for _, line := range strings.Split(string(out), "\n") {
		firstWord := true
		reachedArgs := false
		var pid string
		for _, word := range strings.Split(line, " ") {
			if firstWord {
				firstWord = false
				pid = word
			} else if matched, err := regexp.MatchString(connectProcessRegex, word); matched && err == nil {
				reachedArgs = true
			} else if err != nil {
				return nil, fmt.Errorf(
					"failed to match output of `ps` command (%s) against regular expression for Connect worker process: %w",
					word,
					err,
				)
			} else if reachedArgs && word != daemonOption {
				// TODO: This doesn't work on workers that were started with relative paths to their config files
				//		 unless the CLI is run in the same directory that the Connect worker was started in
				workerConfig := WorkerConfig{Path: word, Use: "Used by Connect process with PID " + pid}
				result = append(result, workerConfig)
				break
			}
		}
	}
	return result, nil
}

func parseComponentId(connector string) (*ComponentId, error) {
	ownerSplit := strings.SplitN(connector, "/", 2)
	if len(ownerSplit) != 2 {
		return nil, errors.New("unable to parse plugin id, make sure it has format: <owner>/<name>:<version|latest>")
	}
	nameVersionSplit := strings.SplitN(ownerSplit[1], ":", 2)
	if len(nameVersionSplit) != 2 {
		return nil, errors.New("unable to parse plugin id, make sure it has format: <owner>/<name>:<version|latest>")
	}
	return &ComponentId{
		Owner:   ownerSplit[0],
		Name:    nameVersionSplit[0],
		Version: nameVersionSplit[1],
	}, nil
}

func downloadArchive(id *ComponentId) (*Component, error) {
	var versionSuffix string
	if id.Version == "latest" {
		versionSuffix = ""
	} else {
		versionSuffix = "/versions/" + id.Version
	}
	manifestUrl := fmt.Sprintf(
		"%s/api/plugins/%s/%s%s",
		confluentHubBackEnd,
		id.Owner,
		id.Name,
		versionSuffix,
	)
	manifestBody, err := queryComponentUrl(manifestUrl, id, "read manifest")
	if manifestBody != nil {
		defer manifestBody.Close()
	}
	if err != nil {
		return nil, err
	}
	manifest := Manifest{}
	if err := json.NewDecoder(manifestBody).Decode(&manifest); err != nil {
		return nil, fmt.Errorf(
			"failed to parse manifest for component %s from Confluent Hub: %w",
			id,
			err,
		)
	}
	archiveBody, err := queryComponentUrl(manifest.Archive.Url, id, "download archive")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download archive for component %s from Confluent Hub: %w",
			id,
			err,
		)
	}
	if id.Version == "latest" {
		id.Version = manifest.Version
	}
	return &Component{
		Id:             id,
		Manifest:       &manifest,
		ArchiveContent: archiveBody,
	}, nil
}

func queryComponentUrl(url string, id *ComponentId, action string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to %s for component %s from Confluent Hub: %w",
			action,
			id,
			err,
		)
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New(fmt.Sprintf(
			"component %s not found in Confluent Hub",
			id,
		))
	} else if resp.StatusCode == http.StatusInternalServerError {
		// TODO: Log full response here? (Perhaps only when verbose mode is enabled?)
		return nil, errors.New(fmt.Sprintf(
			"encountered unexpected server-side error from Confluent Hub while attempting to %s for component %s",
			action,
			id,
		))
	} else if resp.StatusCode != http.StatusOK {
		// TODO: Does this fail on redirects?
		return nil, errors.New(fmt.Sprintf(
			"encountered unexpected error with HTTP status %d from Confluent Hub while attempting to %s for component %s",
			resp.StatusCode,
			action,
			id,
		))
	} else {
		return resp.Body, nil
	}
}

func openLocalArchive(archivePath string) (*Component, error) {
	archive, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open local archive file %s: %w",
			archivePath,
			err,
		)
	}
	defer archive.Close()
	zipContent, err := io.ReadAll(archive)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read local archive file %s: %w",
			archivePath,
			err,
		)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read local archive file %s: %w",
			archivePath,
			err,
		)
	}
	var manifest Manifest
	for _, zipFile := range zipReader.File {
		isManifest, err := filepath.Match("*/manifest.json", filepath.ToSlash(zipFile.Name))
		if err != nil {
			return nil, fmt.Errorf(
				"failed to examine file %s inside local archive file %s: %w",
				zipFile.Name,
				archivePath,
				err,
			)
		}
		if isManifest {
			manifest = Manifest{}
			zipFileReader, err := zipFile.Open()
			if err != nil {
				return nil, fmt.Errorf(
					"failed to open manifest file %s inside local archive file %s: %w",
					zipFile.Name,
					archivePath,
					err,
				)
			}
			defer zipFileReader.Close()
			if err = json.NewDecoder(zipFileReader).Decode(&manifest); err != nil {
				return nil, fmt.Errorf(
					"failed to parse manifest file %s inside local archive file %s: %w",
					zipFile.Name,
					archivePath,
					err,
				)
			}
			break
		}
	}
	archiveContent, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to reopen local archive file %s after parsing manifest contents: %w",
			archivePath,
			err,
		)
	}
	return &Component{
		Id:             manifest.id(),
		Manifest:       &manifest,
		ArchiveContent: archiveContent,
	}, nil
}

func checkIfAlreadyInstalled(id *ComponentId, componentDir string, prompt bool, input *bufio.Reader) error {
	if exists, err := pathExists(componentDir); err != nil {
		return err
	} else if !exists {
		return nil
	}

	bundledPath := filepath.Join(componentDir, id.Name)
	if exists, err := pathExists(bundledPath); err != nil {
		return err
	} else if exists {
		return errors.New("unable to install component because it's already bundled")
	}

	previousInstallationPath := filepath.Join(componentDir, fmt.Sprintf("%s-%s", id.Owner, id.Name))
	if exists, err := pathExists(previousInstallationPath); err != nil {
		return err
	} else if exists {
		if err := maybeRemoveInstallation(previousInstallationPath, prompt, input); err != nil {
			return err
		}
	}

	uberJarPath := filepath.Join(componentDir, fmt.Sprintf("%s-%s.jar", id.Name, id.Version))
	if exists, err := pathExists(uberJarPath); err != nil {
		return err
	} else if exists {
		if err := maybeRemoveInstallation(uberJarPath, prompt, input); err != nil {
			return err
		}
	}

	return nil
}

func maybeRemoveInstallation(pathToComponent string, prompt bool, input *bufio.Reader) error {
	if !prompt {
		fmt.Printf("Automatically uninstalling existing version of the component located at %s", pathToComponent)
	} else {
		for /*ever*/ {
			fmt.Printf("Do you want to uninstall an existing version of this component located at %s? (yn) ", pathToComponent)
			line, err := readLine(input)
			if err != nil {
				return fmt.Errorf("could not read response from terminal: %w", err)
			} else if strings.EqualFold("y", line) {
				break
			} else if strings.EqualFold("n", line) {
				return errors.New(
					"A version of this component is already installed. " +
						"If you want to uninstall existing version, " +
						"confirm or run the command with the \"--no-prompt\" option",
				)
			}
		}
	}
	return os.RemoveAll(pathToComponent)
}

func installArchive(component *Component, componentDir string, integrityCheck bool, prompt bool, input *bufio.Reader, dryRun bool) error {
	id := component.Id
	relativeInstallationDir := filepath.Join(componentDir, fmt.Sprintf("%s-%s", id.Owner, id.Name))
	installationDir, err := filepath.Abs(relativeInstallationDir)
	if err != nil {
		return fmt.Errorf(
			"failed to resolve absolute path for directory %s: %w",
			relativeInstallationDir,
			err,
		)
	}
	zipContent, err := io.ReadAll(component.ArchiveContent)
	if err != nil {
		return fmt.Errorf(
			"failed to read zipped archive for component %s: %w",
			id,
			err,
		)
	}
	if integrityCheck {
		calculatedMd5Checksum := md5.Sum(zipContent)
		if err = verifyChecksum(component.Id, component.md5(), calculatedMd5Checksum[:], "md5"); err != nil {
			return err
		}
		calculatedSha1Checksum := sha1.Sum(zipContent)
		if err = verifyChecksum(component.Id, component.sha1(), calculatedSha1Checksum[:], "sha1"); err != nil {
			return err
		}
	}
	if err := checkLicenseAcceptance(component, prompt, input); err != nil {
		return err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return fmt.Errorf(
			"failed to read zipped archive for component %s: %w",
			id,
			err,
		)
	}

	if dryRun {
		fmt.Println("Skipping installation of connector onto worker as part of dry run mode")
		return nil
	}

	for _, zipFile := range zipReader.File {
		var isDir bool
		var destFileDir string
		if zipFile.FileInfo().IsDir() {
			destFileDir = filepath.Join(installationDir, zipFile.Name)
			isDir = true
		} else {
			destFileDir = filepath.Dir(filepath.Join(installationDir, zipFile.Name))
			isDir = false
		}
		if err = os.MkdirAll(destFileDir, 0755); err != nil {
			return fmt.Errorf(
				"failed to create directory %s on local storage: %w",
				destFileDir,
				err,
			)
		}
		if isDir {
			continue
		}

		zipFileReader, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf(
				"failed to read file %s from archive: %w",
				zipFile.Name,
				err,
			)
		}
		destFilePath := filepath.Join(installationDir, zipFile.Name)
		destFile, _ := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR, 0644)
		_, err = io.Copy(destFile, zipFileReader)
		zipFileReader.Close()
		destFile.Close()
		if err != nil {
			return fmt.Errorf(
				"failed to copy file %s from archive to local file %s: %w",
				zipFile.Name,
				destFilePath,
				err,
			)
		}
	}
	return nil
}

func verifyChecksum(id *ComponentId, expectedChecksum string, actualChecksumRaw []byte, checksumAlgorithm string) error {
	if expectedChecksum != "" {
		actualChecksum := hex.EncodeToString(actualChecksumRaw)
		if actualChecksum != expectedChecksum {
			return errors.New(fmt.Sprintf(
				"%s checksum for downloaded archive (%s) does not match checksum in manifest (%s) for component %s",
				checksumAlgorithm,
				actualChecksum,
				expectedChecksum,
				id,
			))
		}
	} else {
		fmt.Printf(
			"Warning: no %s checksum found for component %s; the integrity of this download cannot be verified\n",
			checksumAlgorithm,
			id,
		)
	}
	return nil
}

func checkLicenseAcceptance(component *Component, prompt bool, input *bufio.Reader) error {
	for _, license := range component.Manifest.License {
		if prompt {
			fmt.Println()
			fmt.Println("Component's license:")
			fmt.Println(license.Name)
			fmt.Println(license.Url)
			for /*ever*/ {
				fmt.Print("I agree to the software license agreement (yn) ")
				line, err := readLine(input)
				if err != nil {
					return fmt.Errorf("could not read response from terminal: %w", err)
				} else if strings.EqualFold("y", line) {
					break
				} else if strings.EqualFold("n", line) {
					return errors.New("you must accept all license agreements for a component in order to install it")
				}
			}
		} else {
			fmt.Println("Implicitly agreeing to license:")
			fmt.Println(license.Name)
			fmt.Println(license.Url)
		}
	}
	return nil
}

func updateWorkerConfig(componentDir string, workerConfigPath string, dryRun bool) error {
	workerConfig, err := properties.LoadFile(workerConfigPath, properties.UTF8)
	if err != nil {
		return fmt.Errorf(
			"failed to parse worker config file %s: %w",
			workerConfigPath,
			err,
		)
	}
	pluginPath := workerConfig.GetString(pluginPathProperty, "")
	pluginPathElements := regexp.MustCompile(" *, *").Split(pluginPath, -1)
	for _, pluginPathElement := range pluginPathElements {
		if pluginPathElement == componentDir {
			// Component directory is already included in the worker's plugin.path
			// No further action required on our part
			return nil
		}
	}
	newPluginPath := strings.Join(append(pluginPathElements, componentDir), ", ")
	if _, _, err = workerConfig.Set(pluginPathProperty, newPluginPath); err != nil {
		return fmt.Errorf(
			"failed to update %s property to %s for worker config %s: %w",
			pluginPathProperty,
			newPluginPath,
			workerConfigPath,
			err,
		)
	}
	fileInfo, err := os.Stat(workerConfigPath)
	if err != nil {
		return fmt.Errorf(
			"failed to stat worker config file %s: %w",
			workerConfigPath,
			err,
		)
	}
	if dryRun {
		fmt.Printf("Skipping update of worker config %s as part of dry run mode\n", workerConfigPath)
		return nil
	}
	workerConfigFile, err := os.OpenFile(workerConfigPath, os.O_TRUNC|os.O_RDWR, fileInfo.Mode())
	if err != nil {
		return fmt.Errorf(
			"failed to open worker config file %s before updating with new %s value %s: %w",
			workerConfigPath,
			pluginPathProperty,
			newPluginPath,
			err,
		)
	}
	defer workerConfigFile.Close()
	// TODO: This doesn't preserve comment structure perfectly
	if _, err = workerConfig.WriteFormattedComment(workerConfigFile, properties.UTF8); err != nil {
		return fmt.Errorf(
			"failed to update worker config file %s with new %s value %s: %w",
			workerConfigPath,
			pluginPathProperty,
			newPluginPath,
			err,
		)
	}
	return nil
}

type Installation struct {
	Type string
	Path string
	Use  string
}

type WorkerConfig struct {
	Path string
	Use  string
}

// Intentionally ignore unknown fields in order to remain flexible for future changes in the
// REST API response format
type ignored map[string]interface{}

type ComponentId struct {
	Owner   string
	Name    string
	Version string
}

type Manifest struct {
	Owner struct {
		Username string  `json:"username"`
		_        ignored `json:"-"`
	} `json:"owner"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Archive struct {
		Url  string  `json:"url"`
		Md5  string  `json:"md5"`
		Sha1 string  `json:"sha1"`
		_    ignored `json:"-"`
	} `json:"archive"`
	License []struct {
		Name string  `json:"name"`
		Url  string  `json:"url"`
		_    ignored `json:"-"`
	} `json:"license"`
	_ ignored `json:"_"`
}

type Component struct {
	Id             *ComponentId
	Manifest       *Manifest
	ArchiveContent io.ReadCloser
}

func (id ComponentId) String() string {
	return fmt.Sprintf("%s/%s:%s", id.Owner, id.Name, id.Version)
}

func (manifest Manifest) id() *ComponentId {
	return &ComponentId{
		manifest.Owner.Username,
		manifest.Name,
		manifest.Version,
	}
}

func (component Component) md5() string {
	return component.Manifest.Archive.Md5
}

func (component Component) sha1() string {
	return component.Manifest.Archive.Sha1
}
