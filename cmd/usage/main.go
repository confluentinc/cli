package main

import (
	"fmt"
	"strings"

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
	whitelist := buildWhitelist()

	lines := []string{
		"BEGIN;",
		"",
		"INSERT INTO whitelist(version, keyword) VALUES",
	}

	for i, keyword := range whitelist {
		delimiter := ','
		if i == len(whitelist)-1 {
			delimiter = ';'
		}

		lines = append(lines, fmt.Sprintf("\t('%s', '%s')%c", version, keyword, delimiter))
	}

	lines = append(lines, "")
	lines = append(lines, "COMMIT;")

	fmt.Println(strings.Join(lines, "\n"))
}

func buildWhitelist() []string {
	whitelist := set.New()

	// Operating Systems
	whitelist.Add("darwin")
	whitelist.Add("linux")
	whitelist.Add("windows")

	// Architectures
	whitelist.Add("amd64")
	whitelist.Add("arm64")

	// Commands and Flags
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
