package test_integ

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

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
	MergedCoverageFilename string
	CollectCoverage        bool
	testNum                int
	tmpArgsFile            *os.File
	coverMode              string
	tmpCoverageFilenames   []string
}

// NewCoverageCollector initializes a CoverageCollector with the specified
// merged coverage filename. collectCoverage can be set to true to collect coverage,
// or set to false to skip coverage collection. This is provided in order to enable reuse of CoverageCollector
// for tests where coverage measurement is not needed.
func NewCoverageCollector(mergedCoverageFilename string, collectCoverage bool) *CoverageCollector {
	collector := &CoverageCollector{
		MergedCoverageFilename: mergedCoverageFilename,
		CollectCoverage:        collectCoverage,
	}
	if collector.MergedCoverageFilename == "" {
		log.Fatal("merged coverage profile filename cannot be empty")
	}
	var err error
	collector.tmpArgsFile, err = ioutil.TempFile("", tmpArgsFilePrefix)
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not create temporary args file"))
	}
	return collector
}

// MergeCoverageProfiles merges the coverage profiles collecting from repeated runs of RunBinary.
// It must be called at the teardown stage of the test suite, otherwise no merged coverage profile will be created. 
func (c *CoverageCollector) MergeCoverageProfiles() {
	if c.testNum == 0 {
		return
	}
	header := fmt.Sprintf("mode: %s", c.coverMode)
	var parsedProfiles []string
	for _, filename := range c.tmpCoverageFilenames {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error reading temp coverage profiles"))
		}
		profile := string(buf)
		loc := strings.Index(profile, header)
		if loc == -1 {
			log.Fatal("coverage mode is missing from coverage profiles")
		}
		parsedProfile := strings.TrimSpace(profile[loc+len(header):])
		parsedProfiles = append(parsedProfiles, parsedProfile)
	}
	mergedProfile := fmt.Sprintf("%s\n%s", header, strings.Join(parsedProfiles, "\n"))
	err := ioutil.WriteFile(c.MergedCoverageFilename, []byte(mergedProfile), 0600)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error merging coverage profiles"))
	}
}

// RunBinary runs the instrumented binary at binPath with env environment variables, executing only the test with mainTestName with the specified args.  
func (c *CoverageCollector) RunBinary(binPath string, mainTestName string, env []string, args []string) (output string, exitCode int, err error) {
	err = c.writeArgs(args)
	if err != nil {
		log.Fatal(err)
	}
	var binArgs string
	if c.CollectCoverage {
		f, err := ioutil.TempFile("", tmpCoverageFilePrefix)
		if err != nil {
			log.Fatal(err)
		}
		c.tmpCoverageFilenames = append(c.tmpCoverageFilenames, f.Name())
		binArgs = fmt.Sprintf("-test.run=%s -test.coverprofile=%s -args-file=%s", mainTestName, f.Name(), c.tmpArgsFile.Name())
	} else {
		binArgs = fmt.Sprintf("-test.run=%s -args-file=%s", mainTestName, c.tmpArgsFile.Name())
	}
	_, _ = fmt.Println(binPath, args)
	cmd := exec.Command(binPath, strings.Split(binArgs, " ")...)
	cmd.Env = append(os.Environ(), env...)
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		// This exit code testing requires 1.12 - https://stackoverflow.com/a/55055100/337735.
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			log.Fatal("error retrieving command exit code")
		}
	} else {
		exitCode = 0
	}
	cmdOutput, coverMode := parseCommandOutput(string(combinedOutput))
	if c.CollectCoverage {
		if c.coverMode == "" {
			c.coverMode = coverMode
		}
		if c.coverMode == "" {
			log.Fatal("coverage mode cannot be empty")
		}
		// https://github.com/wadey/gocovmerge/blob/b5bfa59ec0adc420475f97f89b58045c721d761c/gocovmerge.go#L18
		if c.coverMode != coverMode {
			log.Fatal("cannot merge profiles with different modes")
		}
		if c.coverMode != set && c.coverMode != count && c.coverMode != atomic {
			log.Fatalf("coverage mode cannot be \"%s\". Cover mode must be set, count, or atomic", c.coverMode)
		}
		c.testNum++
	}
	return cmdOutput, exitCode, err
}

func (c *CoverageCollector) writeArgs(args []string) error {
	err := c.tmpArgsFile.Truncate(0)
	if err != nil {
		return err
	}
	_, err = c.tmpArgsFile.Seek(0, 0)
	if err != nil {
		return err
	}
	argStr := strings.Join(args, "\n")
	_, err = c.tmpArgsFile.WriteAt([]byte(argStr), 0)
	if err != nil {
		return err
	}
	return err
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
