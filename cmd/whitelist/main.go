package main

import (
	"fmt"
	"sort"

	"github.com/confluentinc/cli/v3/internal"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/types"
	"github.com/confluentinc/cli/v3/pkg/usage"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
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
	whitelist := types.NewSet[string]()

	for _, os := range []string{"darwin", "linux", "windows"} {
		whitelist.Add(os)
	}

	for _, arch := range []string{"amd64", "arm64"} {
		whitelist.Add(arch)
	}

	whitelist.Add("__complete")

	// Compile a whitelist for all three subsets of commands: no context, cloud, and on-prem
	configs := []*config.Config{
		{CurrentContext: "No Context"},
		{CurrentContext: "Cloud", Contexts: map[string]*config.Context{"Cloud": {PlatformName: "https://confluent.cloud"}}},
		{CurrentContext: "On-Prem", Contexts: map[string]*config.Context{"On-Prem": {PlatformName: "https://example.com"}}},
	}
	for _, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		cmd := internal.NewConfluentCommand(cfg)
		usage.WhitelistCommandsAndFlags(cmd, whitelist)
	}

	return whitelist.Slice()
}
