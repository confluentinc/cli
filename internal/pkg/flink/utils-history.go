package main

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

func saveHistory(history []string) {
	// Convert struct to JSON
	b, err := json.Marshal(History{history})
	if err != nil {
		panic(err)
	}

	// Write JSON to file
	f, err := os.Create("test.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}
