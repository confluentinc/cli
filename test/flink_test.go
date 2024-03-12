package test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/go-prompt"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
)

const (
	flinkShellFixtureInputFolder  = "test/fixtures/input/flink/shell"
	flinkShellFixtureOutputFolder = "test/fixtures/output/flink/shell"
	timezoneEnvVar                = "TZ"
)

type FlinkShellTest struct {
	inputFile  string
	goldenFile string
	timeout    time.Duration
}

func (s *CLITestSuite) TestFlinkComputePool() {
	tests := []CLITest{
		{args: "flink compute-pool create my-compute-pool --cloud aws --region us-west-2", fixture: "flink/compute-pool/create.golden"},
		{args: "flink compute-pool describe lfcp-123456", fixture: "flink/compute-pool/describe.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list.golden"},
		{args: "flink compute-pool list --region us-west-2", fixture: "flink/compute-pool/list-region.golden"},
		{args: "flink compute-pool update lfcp-123456 --max-cfu 5", fixture: "flink/compute-pool/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolDelete() {
	tests := []CLITest{
		{args: "flink compute-pool delete lfcp-123456 --force", fixture: "flink/compute-pool/delete.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "n\n", fixture: "flink/compute-pool/delete-multiple-refuse.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "y\n", fixture: "flink/compute-pool/delete-multiple-success.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-654321", fixture: "flink/compute-pool/delete-multiple-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUnset() {
	tests := []CLITest{
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset-before-use.golden"},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use-before-unset.golden"},
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUse() {
	tests := []CLITest{
		{args: "flink compute-pool use lfcp-999999", login: "cloud", fixture: "flink/compute-pool/use-fail.golden", exitCode: 1},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use.golden"},
		{args: "flink compute-pool describe", fixture: "flink/compute-pool/describe-after-use.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-after-use.golden"},
		{args: "flink compute-pool update --max-cfu 5", fixture: "flink/compute-pool/update-after-use.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkRegion() {
	tests := []CLITest{
		{args: "flink region use --cloud aws --region eu-west-1", fixture: "flink/region/use.golden"},
		{args: "flink region list", fixture: "flink/region/list.golden"},
		{args: "flink region list -o json", fixture: "flink/region/list-json.golden"},
		{args: "flink region list --cloud aws", fixture: "flink/region/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatement() {
	tests := []CLITest{
		{args: "flink statement delete my-statement --force --cloud aws --region eu-west-1", fixture: "flink/statement/delete.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1", fixture: "flink/statement/list.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/list-yaml.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status completed", fixture: "flink/statement/list-completed.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status pending", fixture: "flink/statement/list-pending.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/describe.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/describe-yaml.golden"},
		{args: "flink statement stop my-statement --region eu-west-1 --cloud aws", fixture: "flink/statement/stop.golden"},
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/exception/list.golden"},
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/exception/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatementCreate() {
	tests := []CLITest{
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "flink/statement/create.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456`, fixture: "flink/statement/create-service-account-warning.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 --wait`, fixture: "flink/statement/create-wait.golden"},
		{args: `flink statement create --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 -o yaml`, fixture: "flink/statement/create-no-name-yaml.golden", regex: true},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlink_Autocomplete() {
	tests := []CLITest{
		{args: `__complete flink compute-pool create my-compute-pool --cloud ""`, fixture: "flink/compute-pool/create-cloud-autocomplete.golden"},
		{args: `__complete flink compute-pool create my-compute-pool --cloud aws --region ""`, fixture: "flink/compute-pool/create-region-autocomplete.golden"},
		{args: `__complete flink compute-pool delete ""`, fixture: "flink/compute-pool/delete-autocomplete.golden"},
		{args: `__complete flink compute-pool list --region ""`, fixture: "flink/compute-pool/list-region-autocomplete.golden"},
		{args: `__complete flink statement create my-statement --database ""`, fixture: "flink/statement/create-database-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkShell() {
	tests := []FlinkShellTest{
		{inputFile: "use-catalog", goldenFile: "use-catalog.golden"},
		{inputFile: "use-database", goldenFile: "use-database.golden"},
		{inputFile: "set-key", goldenFile: "set-key.golden"},
		{inputFile: "reset-key", goldenFile: "reset-key.golden"},
		{inputFile: "reset-all", goldenFile: "reset-all.golden"},
	}

	for _, test := range tests {
		s.runFlinkShellTest(test)
	}
}

func (s *CLITestSuite) runFlinkShellTest(flinkShellTest FlinkShellTest) {
	s.T().Run(flinkShellTest.inputFile, func(t *testing.T) {
		s.login(t)

		// Create a file for go-prompt to use as the input stream
		stdin, err := os.Create("flink_shell_input_stream.txt")
		require.NoError(t, err, "error creating file")
		defer cleanupInputFile(stdin)

		envVars, err := setEnvVars(stdin.Name())
		require.NoError(t, err, "error setting env vars")
		defer cleanupEnvVars(envVars)

		// Start flink shell
		dir, err := os.Getwd()
		require.NoError(t, err)
		cmd := exec.Command(filepath.Join(dir, testBin), "flink", "shell", "--compute-pool", "lfcp-123456")

		// Register stdout scanner
		pipe, err := cmd.StdoutPipe()
		require.NoError(t, err)
		defer pipe.Close()
		stdoutScanner := bufio.NewScanner(pipe)

		// Start command
		err = cmd.Start()
		require.NoError(t, err)
		go killCommandIfTimeoutExceeded(flinkShellTest.timeout, cmd)

		output := &strings.Builder{}
		waitForFlinkShellToBeReady(stdoutScanner, output)

		// Execute commands
		commands, err := getCommandsFromFixture(filepath.Join(dir, flinkShellFixtureInputFolder, flinkShellTest.inputFile))
		require.NoError(t, err)
		err = executeCommands(stdin, commands, stdoutScanner, output)
		require.NoError(t, err)

		// Wait for flink shell to exit
		err = cmd.Wait()
		require.NoError(t, err)

		// Compare to golden file
		snapshotConfig := cupaloy.New(
			cupaloy.SnapshotSubdirectory(filepath.Join(dir, flinkShellFixtureOutputFolder)),
			// Update snapshot if update flag was set
			cupaloy.ShouldUpdate(func() bool {
				return *update
			}),
		)
		assert.NoError(t, snapshotConfig.SnapshotWithName(flinkShellTest.goldenFile, output.String()),
			fmt.Sprintf("full output was %s", output.String()))
	})
}

func (s *CLITestSuite) login(t *testing.T) {
	loginString := fmt.Sprintf("login --url %s", s.TestBackend.GetCloudUrl())
	env := []string{pauth.ConfluentCloudEmail + "=fake@user.com", pauth.ConfluentCloudPassword + "=pass1"}
	for _, e := range env {
		keyVal := strings.Split(e, "=")
		os.Setenv(keyVal[0], keyVal[1])
	}

	defer func() {
		for _, e := range env {
			keyVal := strings.Split(e, "=")
			os.Unsetenv(keyVal[0])
		}
	}()

	if output := runCommand(t, testBin, env, loginString, 0, ""); *debug {
		fmt.Println(output)
	}
}

func cleanupInputFile(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Printf("failed to close file: %v\n", err)
	}
	err = os.Remove(file.Name())
	if err != nil {
		fmt.Printf("failed to remove file: %v\n", err)
	}
}

func setEnvVars(inputFileName string) ([]string, error) {
	var envVars []string

	// Set the go-prompt file input env var, so go-prompt uses this file as the input stream
	if err := os.Setenv(prompt.EnvVarInputFile, inputFileName); err != nil {
		return envVars, err
	}
	envVars = append(envVars, prompt.EnvVarInputFile)

	// Fake the timezone, to ensure CI and local run with the same default timezone
	if err := os.Setenv(timezoneEnvVar, "Europe/London"); err != nil {
		return envVars, err
	}
	envVars = append(envVars, timezoneEnvVar)

	return envVars, nil
}

func cleanupEnvVars(envVars []string) {
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func killCommandIfTimeoutExceeded(timeout time.Duration, cmd *exec.Cmd) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	time.Sleep(timeout)
	if err := cmd.Process.Kill(); err != nil {
		fmt.Printf("failed to kill command %v/n", err)
	}
}

func waitForFlinkShellToBeReady(scanner *bufio.Scanner, output *strings.Builder) {
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line + "\n")
		if strings.Contains(line, "[Ctrl-Q] Quit [Ctrl-S] Toggle Smart Completion") {
			break
		}
	}
}

func getCommandsFromFixture(inputFilePath string) ([]string, error) {
	file, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(file), "\n"), nil
}

func executeCommands(stdin *os.File, commands []string, stdoutScanner *bufio.Scanner, output *strings.Builder) error {
	for _, command := range commands {
		// Simulate the user entering a command and add a new line to flush the output buffer
		_, err := stdin.WriteString(command + "\n")
		if err != nil {
			return err
		}

		for stdoutScanner.Scan() {
			// Strip all terminal control sequences and skip empty lines
			line := removeANSIEscapeSequences(stdoutScanner.Text())
			if line == "" {
				continue
			}

			// Record the output
			output.WriteString(line + "\n")

			// Once we've seen our command be printed, we know we can continue to submit the statement.
			// This effectively means we wait for the previous command to complete, since the new command will only
			// be printed, if the previous one is finished.
			if strings.Contains(line, fmt.Sprintf("> %s", command)) {
				err := submitStatement(stdin)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func removeANSIEscapeSequences(input string) string {
	stripColors := regexp.MustCompile(`\x1b\[[0-9;]*[JKmsu]`)
	stripBellCharacter := regexp.MustCompile(`\a`)
	stripTitle := regexp.MustCompile(`\x1B]2;`)
	stripSqlPrompt := regexp.MustCompile(`sql-prompt`)
	stripErasePrefix := regexp.MustCompile(`> \x1b\[2D`)
	stripCursorUp := regexp.MustCompile(`\x1b\[1A.*`)

	input = stripColors.ReplaceAllString(input, "")
	input = stripBellCharacter.ReplaceAllString(input, "")
	input = stripTitle.ReplaceAllString(input, "")
	input = stripSqlPrompt.ReplaceAllString(input, "")
	input = stripErasePrefix.ReplaceAllString(input, "")
	input = stripCursorUp.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

func submitStatement(stdin *os.File) error {
	// We need to wait before and after writing a new line, so go-prompt receives the ENTER as an individual key press
	time.Sleep(50 * time.Millisecond)
	_, err := stdin.WriteString("\n")
	time.Sleep(50 * time.Millisecond)
	return err
}
