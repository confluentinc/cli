package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type MockData struct {
	Data []string `json:"data"`
}

func test() {
	jsonFile, err := os.Open("mock-data.json")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened mock-data.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var mockData MockData
	json.Unmarshal(byteValue, &mockData)

	for i := 0; i < len(mockData.Data); i++ {
		fmt.Printf("%d %s \n", i, mockData.Data[i])
	}
}
