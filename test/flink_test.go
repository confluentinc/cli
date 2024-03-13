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
	flinkShellFixtureOutputFolder = "test/fixtures/output/flink/shell"
	timezoneEnvVar                = "TZ"
)

type FlinkShellTest struct {
	name       string
	commands   []string
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
		{
			name:       "TestUseCatalog",
			goldenFile: "use-catalog.golden",
			commands: []string{
				"use catalog default;",
				"set;",
			},
		},
		{
			name:       "TestUseDatabase",
			goldenFile: "use-database.golden",
			commands: []string{
				"use db1;",
				"set;",
			},
		},
		{
			name:       "TestSetSingleKey",
			goldenFile: "set-key.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set;",
			},
		},
		{
			name:       "TestResetSingleKey",
			goldenFile: "reset-key.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"reset 'cli.a-key';",
				"set;",
			},
		},
		{
			name:       "TestResetAllKeys",
			goldenFile: "reset-all.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set 'cli.another-key'='another value';",
				"reset;", "set;",
			},
		},
	}

	for _, test := range tests {
		s.runFlinkShellTest(test)
	}
}

func (s *CLITestSuite) runFlinkShellTest(flinkShellTest FlinkShellTest) {
	s.T().Run(flinkShellTest.name, func(t *testing.T) {
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
		output.WriteString(waitForLine(stdoutScanner, "[Ctrl-Q] Quit [Ctrl-S] Toggle Smart Completion"))

		// Execute commands
		require.NoError(t, err)
		outputFromCommands, err := executeCommands(stdin, flinkShellTest.commands, stdoutScanner)
		require.NoError(t, err)
		output.WriteString(outputFromCommands)

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

func executeCommands(stdin *os.File, commands []string, stdoutScanner *bufio.Scanner) (string, error) {
	// add exit command to ensure we always close the flink shell
	commands = append(commands, "exit")
	output := strings.Builder{}
	for _, command := range commands {
		// Simulate the user entering a command and add a new line to flush the output buffer
		_, err := stdin.WriteString(command + "\n")
		if err != nil {
			return "", err
		}

		output.WriteString(waitForLine(stdoutScanner, fmt.Sprintf("> %s", command)))

		// submit the statement
		_, err = stdin.WriteString("\n")
		if err != nil {
			return "", err
		}

		output.WriteString(waitForLine(stdoutScanner, "Statement successfully submitted."))
	}
	return output.String(), nil
}

func waitForLine(stdoutScanner *bufio.Scanner, lineToWaitFor string) string {
	output := strings.Builder{}
	for stdoutScanner.Scan() {
		// Strip all terminal control sequences and skip empty lines
		line := removeAnsiEscapeSequences(stdoutScanner.Text())
		if line == "" {
			continue
		}

		// Record the output
		output.WriteString(line + "\n")

		// Once we've seen the line we wanted to wait for, we break.
		if strings.HasPrefix(line, lineToWaitFor) {
			break
		}
	}
	return output.String()
}

func removeAnsiEscapeSequences(input string) string {
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
