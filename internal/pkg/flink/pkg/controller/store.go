package controller

import (
	"encoding/json"
	"fmt"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
	"os"
)

type Store struct {
	Data  []string `json:"data"`
	index int
}

func (s *Store) FetchData(query string) string {
	s.index++
	return s.Data[s.index%len(s.Data)]
}

func NewStore(client *v2.APIClient) Store {
	store := Store{}
	// Opening mock data
	jsonFile, err := os.ReadFile("mock-data.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(jsonFile, &store)

	return store
}
