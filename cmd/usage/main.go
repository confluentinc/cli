package main

import (
	"fmt"
	"sort"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/usage"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var version = "v0.0.0"

/*
	Generate a DB migration for cc-cli-service that updates its whitelist. The output will look like the following:

	BEGIN;

	INSERT INTO whitelist(version, keyword) VALUES
		('v2.16.0', 'confluent'),
		('v2.16.0', 'login');

	COMMIT;
*/

func main() {
	fmt.Println("BEGIN;")
	fmt.Println("")
	fmt.Println("INSERT INTO whitelist(version, keyword) VALUES")

	whitelist := buildWhitelist()
	sort.Strings(whitelist)

	for i, keyword := range whitelist {
		delimiter := ','
		if i == len(whitelist)-1 {
			delimiter = ';'
		}

		fmt.Printf("\t('%s', '%s')%c\n", version, keyword, delimiter)
	}

	fmt.Println("")
	fmt.Println("COMMIT;")
}

func buildWhitelist() []string {
	whitelist := set.New()

	for _, os := range []string{"darwin", "linux", "windows"} {
		whitelist.Add(os)
	}

	for _, arch := range []string{"amd64", "arm64"} {
		whitelist.Add(arch)
	}

	// Certain commands and flags are only present in Confluent Cloud or Confluent Platform.
	// Consider all contexts when compiling the whitelist.
	configs := []*v1.Config{
		{CurrentContext: "A"},
		{CurrentContext: "B", Contexts: map[string]*v1.Context{"B": {PlatformName: ccloudv2.Hostnames[0]}}},
		{CurrentContext: "C", Contexts: map[string]*v1.Context{"C": {PlatformName: "https://example.com"}}},
	}
	for _, cfg := range configs {
		cmd := pcmd.NewConfluentCommand(cfg, new(pversion.Version), false)
		usage.WhitelistCommandsAndFlags(cmd, whitelist)
	}

	return whitelist.Slice()
}
