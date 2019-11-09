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
	tmpCoverageFmt = "temp_coverage%d.out"
)

type CoverageCollector struct {
	T                      *testing.T
	testNum                int
	tmpArgsFile            *os.File
	MergedCoverageFilename string
}

func (c *CoverageCollector) Setup() {
	var err error
	c.tmpArgsFile, err = ioutil.TempFile("", "integ_args")
	require.NoError(c.T, errors.Wrap(err, "could not create temporary args file"))
	require.NotEmpty(c.T, c.MergedCoverageFilename, "must specify merged coverage profile filename")
}

func (c *CoverageCollector) TearDown(header string) {
	if c.testNum == 0 {
		return
	}
	// Merge coverage profiles.
	mergedProfile := header
	cleanUp := func() {
		for i := 0; i < c.testNum; i++ {
			filename := fmt.Sprintf(tmpCoverageFmt, i)
			err := os.Remove(filename)
			// Log error but continue.
			if err != nil {
				c.T.Log(err)
			}
		}
	}
	defer cleanUp()
	for i := 0; i < c.testNum; i++ {
		filename := fmt.Sprintf(tmpCoverageFmt, i)
		buf, err := ioutil.ReadFile(filename)
		require.NoError(c.T, errors.Wrap(err, "error merging coverage profiles"))
		profile := string(buf)
		pattern := fmt.Sprintf("^%s", header)
		re := regexp.MustCompile(pattern)
		loc := re.FindStringIndex(profile)
		if loc == nil {
			c.T.Fatal("Coverage mode is missing from coverage profiles")
		}
		mergedProfile += profile[loc[1]+1:]
	}
	err := ioutil.WriteFile(c.MergedCoverageFilename, []byte(mergedProfile), 0666)
	require.NoError(c.T, errors.Wrap(err, "error merging coverage profiles"))
}

func (c *CoverageCollector) RunCommand(t *testing.T, binPath string, env []string, args string, wantErrCode int, cover bool) string {
	c.writeArgs(t, args)
	if cover {
		args = fmt.Sprintf("-test.run=TestRunMain -test.coverprofile="+tmpCoverageFmt+" -args-file=%s", c.testNum, c.tmpArgsFile.Name())
		c.testNum++
	} else {
		args = fmt.Sprintf("-test.run=TestRunMain -args-file=%s", c.tmpArgsFile.Name())
	}
	_, _ = fmt.Println(binPath, args)
	cmd := exec.Command(binPath, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This exit code testing requires 1.12 - https://stackoverflow.com/a/55055100/337735
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
	return parseCommandOutput(string(output))
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

func parseCommandOutput(output string) string {
	divIndex := strings.Index(output, endOfInputDivider)
	if divIndex == -1 {
		panic("Integration test divider is missing")
	}
	cmdOutput := output[:divIndex]
	return cmdOutput
}
