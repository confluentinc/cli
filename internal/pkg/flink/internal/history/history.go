package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const HISTORY_FILE_NAME = "flink_statements_history.json"

type History struct {
	Data          []string `json:"data"`
	confluentPath string
	historyPath   string
}

func LoadHistory() *History {
	history := initPath()
	return loadFromPath(history)
}

func loadFromPath(history *History) *History {
	if history == nil {
		return nil
	}
	jsonFile, err := os.ReadFile(history.historyPath)
	if errors.Is(err, os.ErrNotExist) {
		log.CliLogger.Warnf("Couldn't load past statements: file doesn't exist: %s. This is expected if that's the first time you're using the Flink SQL Client!", history.historyPath)
		return history
	}

	if err != nil {
		log.CliLogger.Warnf("Couldn't load past statements history: unable to read file: %v", err)
	}

	if err := json.Unmarshal(jsonFile, history); err != nil {
		log.CliLogger.Warnf("Couldn't load past statements history: %v", err)
	}
	return history
}

func initPath() *History {
	home, osHomedirErr := os.UserHomeDir()
	if osHomedirErr != nil {
		log.CliLogger.Warnf("Couldn't get homedir with os.UserHomeDir(): %v", osHomedirErr)
		return nil
	}

	confluentDir := os.Getenv(config.HomeConfluentPathEnvVar)
	if confluentDir == "" {
		confluentDir = config.HomeConfluentPathDefault
	}
	confluentPath := filepath.Join(home, confluentDir)
	historyPath := filepath.Join(confluentPath, HISTORY_FILE_NAME)

	return &History{
		Data:          nil,
		confluentPath: confluentPath,
		historyPath:   historyPath,
	}
}

func (history *History) Save() {
	// Limit history to 500 entries
	if len(history.Data) > 500 {
		history.Data = history.Data[len(history.Data)-500:]
	}

	// Convert struct to JSON
	b, err := json.Marshal(history)
	if err != nil {
		log.CliLogger.Warnf("Couldn't save past statements history: couldn't marshal history: %v", err)
	}

	if err := os.Mkdir(history.confluentPath, os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.CliLogger.Warnf("Couldn't save past statements history: couldn't create directory: %v", err)
		}
	}

	// Write JSON to file
	f, err := os.Create(history.historyPath)
	if err != nil {
		log.CliLogger.Warnf("Couldn't save past statements history: couldn't create file: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		log.CliLogger.Warnf("Couldn't save past statements history: couldn't write to file: %v", err)
	}
}

func (history *History) Append(statements ...string) {
	formattedStatements := formatStatements(statements)
	history.Data = append(history.Data, formattedStatements...)
}

// This removes empty statements
// and also repeated consecutive statetments
func formatStatements(statements []string) []string {
	formattedStatements := []string{}
	var previousStatement string

	for _, statement := range statements {
		if statement != "" && statement != previousStatement {
			formattedStatements = append(formattedStatements, statement)
			previousStatement = statement
		}
	}

	return formattedStatements
}
