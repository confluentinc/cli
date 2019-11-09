package test_integ

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	set                   = "set"
	count                 = "count"
	atomic                = "atomic"
	tmpArgsFilePrefix     = "integ_args"
	tmpCoverageFilePrefix = "temp_coverage"
)

type CoverageCollector struct {
	T                      *testing.T
	MergedCoverageFilename string
	testNum                int
	tmpArgsFile            *os.File
	coverMode              string
	tmpCoverageFilenames   []string
}

func (c *CoverageCollector) Setup() {
	var err error
	c.tmpArgsFile, err = ioutil.TempFile("", tmpArgsFilePrefix)
	require.NoError(c.T, errors.Wrap(err, "could not create temporary args file"))
	require.NotEmpty(c.T, c.MergedCoverageFilename, "must specify merged coverage profile filename")
}

func (c *CoverageCollector) TearDown() {
	if c.testNum == 0 {
		return
	}
	// Merge coverage profiles.
	header := fmt.Sprintf("mode: %s", c.coverMode)
	mergedProfile := header
	for _, filename := range c.tmpCoverageFilenames {
		buf, err := ioutil.ReadFile(filename)
		require.NoError(c.T, errors.Wrap(err, "error reading temp coverage profiles"))
		profile := string(buf)
		pattern := fmt.Sprintf("^%s", header)
		re := regexp.MustCompile(pattern)
		loc := re.FindStringIndex(profile)
		if loc == nil {
			c.T.Fatal("coverage mode is missing from coverage profiles")
		}
		mergedProfile += profile[loc[1]+1:]
	}
	err := ioutil.WriteFile(c.MergedCoverageFilename, []byte(mergedProfile), 0666)
	require.NoError(c.T, errors.Wrap(err, "error merging coverage profiles"))
}

func (c *CoverageCollector) RunCommand(t *testing.T, binPath string, env []string, args string, wantErrCode int, cover bool) string {
	c.writeArgs(t, args)
	// TODO: Make "TestRunMain" dynamic.
	if cover {
		f, err := ioutil.TempFile("", tmpCoverageFilePrefix)
		require.NoError(t, err)
		c.tmpCoverageFilenames = append(c.tmpCoverageFilenames, f.Name())
		args = fmt.Sprintf("-test.run=TestRunMain -test.coverprofile=%s -args-file=%s", f.Name(), c.tmpArgsFile.Name())
	} else {
		args = fmt.Sprintf("-test.run=TestRunMain -args-file=%s", c.tmpArgsFile.Name())
	}
	_, _ = fmt.Println(binPath, args)
	cmd := exec.Command(binPath, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This exit code testing requires 1.12 - https://stackoverflow.com/a/55055100/337735.
		if exitError, ok := err.(*exec.ExitError); ok {
			if wantErrCode == 0 {
				require.Failf(t, "unexpected error",
					"exit %d: %s\n%s", exitError.ExitCode(), args, string(output))
			} else {
				require.Equal(t, wantErrCode, exitError.ExitCode())
			}
		} else {
			require.Failf(t, "unexpected error", "command returned err: %s", err)
		}
	} else {
		require.Equal(t, wantErrCode, 0)
	}
	cmdOutput, coverMode := parseCommandOutput(string(output))
	if cover {
		if c.coverMode == "" {
			c.coverMode = coverMode
		}
		require.NotEmpty(t, c.coverMode)
		// https://github.com/wadey/gocovmerge/blob/b5bfa59ec0adc420475f97f89b58045c721d761c/gocovmerge.go#L18	
		require.Equal(t, c.coverMode, coverMode, "cannot merge profiles with different modes")
		if c.coverMode != set && c.coverMode != count && c.coverMode != atomic {
			require.FailNow(c.T, "cover mode must be set, count, or atomic")
		}
		c.testNum++
	}
	return cmdOutput
}

func (c *CoverageCollector) writeArgs(t *testing.T, args string) {
	err := c.tmpArgsFile.Truncate(0)
	require.NoError(t, err)
	_, err = c.tmpArgsFile.Seek(0, 0)
	require.NoError(t, err)
	args = strings.ReplaceAll(args, " ", "\n")
	_, err = c.tmpArgsFile.WriteAt([]byte(args), 0)
	require.NoError(t, err)
}

func parseCommandOutput(output string) (cmdOutput string, coverMode string) {
	divIndex := strings.Index(output, endOfInputDivider)
	if divIndex == -1 {
		panic("Integration test divider is missing")
	}
	cmdOutput = output[:divIndex]
	tail := output[divIndex+len(endOfInputDivider):]
	// Trim extra newline after cmd output.
	tail = strings.TrimPrefix(tail, "\n")
	coverMode = tail[:strings.Index(tail, "\n")]
	return cmdOutput, coverMode
}
