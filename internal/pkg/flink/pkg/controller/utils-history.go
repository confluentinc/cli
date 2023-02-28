package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type History struct {
	Data []string `json:"data"`
}

func LoadHistory() (history History) {
	// Opening data
	jsonFile, err := os.ReadFile(".cache/history.json")
	if errors.Is(err, os.ErrNotExist) {
		return History{}
	}

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = json.Unmarshal(jsonFile, &history)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return history
}

func (history *History) Save() {
	// Limit history to 500 entries
	if len(history.Data) > 500 {
		history.Data = history.Data[len(history.Data)-500:]
	}

	// Convert struct to JSON
	b, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}

	if err := os.Mkdir(".cache", os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			panic(err)
		}
	}

	// Write JSON to file

	f, err := os.Create(".cache/history.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}

func (history *History) Append(statements []string) {
	formattedStatements := formatStatements(statements)
	history.Data = append(history.Data, formattedStatements...)
}

func formatStatements(statements []string) []string {
	var formattedStatements []string

	for _, statement := range statements {
		if statement != "" {
			formattedStatements = append(formattedStatements, statement)
		}
	}
	return formattedStatements
}
