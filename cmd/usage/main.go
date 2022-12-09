package main

import (
	"fmt"
	"sort"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	pcmd "github.com/confluentinc/cli/internal/cmd"
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

	whitelist.Add("__complete")

	// Compile a whitelist for all three subsets of commands: no context, cloud, and on-prem
	configs := []*v1.Config{
		{CurrentContext: "No Context"},
		{CurrentContext: "Cloud", Contexts: map[string]*v1.Context{"Cloud": {PlatformName: "https://confluent.cloud", State: &v1.ContextState{Auth: &v1.AuthConfig{Organization: &ccloudv1.Organization{}}}}}},
		{CurrentContext: "On-Prem", Contexts: map[string]*v1.Context{"On-Prem": {PlatformName: "https://example.com"}}},
	}
	for _, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		cmd := pcmd.NewConfluentCommand(cfg)
		usage.WhitelistCommandsAndFlags(cmd, whitelist)
	}

	return whitelist.Slice()
}
