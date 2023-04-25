package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type releaseNotes map[string][]string

func main() {
	r := releaseNotes{}
	if err := json.NewDecoder(os.Stdin).Decode(&r); err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "s3":
		fmt.Print(s3(r))
	case "github":
		fmt.Print(github(r))
	}
}

func s3(r releaseNotes) string {
	return r["New Features"][0]
}

func github(r releaseNotes) string {
	return r["New Features"][0]
}
