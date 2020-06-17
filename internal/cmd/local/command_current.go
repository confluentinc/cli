package local

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewCurrentCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	currentCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "current",
			Short: "Get the path of the data and logs for the current Confluent run.",
			Args:  cobra.NoArgs,
			RunE:  runCurrentCommand,
		},
		cfg, prerunner)

	return currentCommand.Command
}

func runCurrentCommand(command *cobra.Command, _ []string) error {
	current, err := getConfluentCurrent()
	if err != nil {
		return err
	}

	command.Println(current)
	return nil
}

func getConfluentCurrent() (string, error) {
	root := os.Getenv("CONFLUENT_CURRENT")
	if root == "" {
		root = os.TempDir()
	}

	var confluentCurrent string

	trackingFile := filepath.Join(root, "confluent.current")
	if _, err := os.Stat(trackingFile); os.IsNotExist(err) {
		confluentCurrent = createCurrentDirectory(root)
		if err := os.Mkdir(confluentCurrent, 0777); err != nil {
			return "", err
		}
		if err := ioutil.WriteFile(trackingFile, []byte(confluentCurrent), 0644); err != nil {
			return "", err
		}
	} else {
		data, err := ioutil.ReadFile(trackingFile)
		if err != nil {
			return "", err
		}
		confluentCurrent = strings.TrimSuffix(string(data), "\n")
	}

	return confluentCurrent, nil
}

func createCurrentDirectory(parentDir string) string {
	rand.Seed(time.Now().Unix())

	for {
		childDir := fmt.Sprintf("confluent.%06d", rand.Intn(1000000))
		path := filepath.Join(parentDir, childDir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}
