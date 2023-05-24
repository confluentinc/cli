package history

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
)

const HISTORY_FILE_NAME = "flink_statements_history.json"

type History struct {
	Data          []string `json:"data"`
	confluentPath string
	historyPath   string
}

func LoadHistory() (history *History) {
	historyPath, history := initPath(history)

	return loadFromPath(historyPath, history)
}

func loadFromPath(historyPath string, history *History) *History {
	jsonFile, err := os.ReadFile(historyPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Couldn't load past statements: file doesn't exist: %s. Error: %v", historyPath, err)
		return history
	}

	if err != nil {
		log.Printf("Couldn't load past statements history: unable to read file Error: " + err.Error())
	}

	err = json.Unmarshal(jsonFile, &history)
	if err != nil {
		log.Printf("Couldn't load past statements history. Error: " + err.Error())
	}
	return history
}

func initPath(history *History) (string, *History) {
	if history == nil {
		history = &History{}
	}

	home, osHomedirErr := os.UserHomeDir()
	if osHomedirErr != nil {
		log.Printf("Couldn't get homedir with os.UserHomeDir(). Error: " + osHomedirErr.Error())
	}

	confluentDir := os.Getenv(config.CONFLUENT_HOME_DIR_ENV_VAR)
	if confluentDir == "" {
		confluentDir = config.CONFLUENT_HOME_DIR_DEFAULT
	}
	confluentPath := filepath.Join(home, confluentDir)
	historyPath := filepath.Join(confluentPath, HISTORY_FILE_NAME)

	history.confluentPath = confluentPath
	history.historyPath = historyPath

	return historyPath, history
}

func (history *History) Save() {
	// Limit history to 500 entries
	if len(history.Data) > 500 {
		history.Data = history.Data[len(history.Data)-500:]
	}

	// Convert struct to JSON
	b, err := json.Marshal(history)
	if err != nil {
		log.Printf("Couldn't save past statements history: couldn't marhsal history. Error: " + err.Error())
	}

	if err := os.Mkdir(history.confluentPath, os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Printf("Couldn't save past statements history: couldn't create directory. Error: " + err.Error())
		}
	}

	// Write JSON to file
	f, err := os.Create(history.historyPath)
	if err != nil {
		log.Printf("Couldn't save past statements history: couldn't create file. Error: " + err.Error())
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		log.Printf("Couldn't save past statements history: couldn't write to file. Error: " + err.Error())
	}
}

func (history *History) Append(statements []string) {
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
