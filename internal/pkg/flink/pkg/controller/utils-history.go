package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type History struct {
	Data []string `json:"data"`
}

func loadHistory() []string {

	// Opening data
	jsonFile, err := os.Open(".cache/history.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var history History
	json.Unmarshal(byteValue, &history)

	return history.Data
}

func SaveHistory(history []string) {
	// Limit history to 500 entries
	length := len(history)
	if length > 500 {
		history = history[length-500:]
	}

	// Convert struct to JSON
	b, err := json.Marshal(History{history})
	if err != nil {
		panic(err)
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

func appendToHistory(history []string, statements []string) []string {
	formattedStatements := formatStatements(statements)

	history = append(history, formattedStatements...)
	return history
}

func formatStatements(statements []string) []string {
	var formattedStatements []string

	for _, statement := range statements {
		if statement != "" {
			formattedStatements = append(formattedStatements, statement+";")
		}
	}
	return formattedStatements
}
