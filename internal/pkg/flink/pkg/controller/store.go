package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type MockData struct {
	Data []string `json:"data"`
}

type Store struct {
	fetchData func() string
}

func NewStore() Store {
	i := -1

	// Opening mock data
	jsonFile, err := os.Open("mock-data.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var mockData MockData
	json.Unmarshal(byteValue, &mockData)

	// Actions
	fetchData := func() string {
		i++
		return mockData.Data[i%len(mockData.Data)]
	}

	return Store{fetchData}
}
