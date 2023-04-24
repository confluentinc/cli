package main

import (
	"flag"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"os"
	"strings"

	"github.com/client9/gospell"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/linter"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var commandRules = []linter.CommandRule{
	// Hard Requirements
	linter.RequireLowerCase("Use"),
	linter.RequireRealWords("Use", '-'),
	linter.Filter(linter.RequireSingular("Use"),
		linter.ExcludeCommandContains("local services"),
		linter.ExcludeCommand("kafka client-config create nodejs")),

	linter.RequireCapitalizeProperNouns("Short", properNouns),
	linter.RequireEndWithPunctuation("Short", false),
	linter.Filter(linter.RequireNotTitleCase("Short", properNouns), linter.ExcludeCommandContains("ksql app")),
	linter.RequireStartWithCapital("Short"),

	linter.Filter(linter.RequireEndWithPunctuation("Long", true), linter.ExcludeCommand("prompt")),
	linter.Filter(linter.RequireCapitalizeProperNouns("Long", properNouns),
		linter.ExcludeCommand("completion"),
		linter.ExcludeCommandContains("kafka client-config create")),
	linter.RequireStartWithCapital("Long"),

	linter.RequireListRequiredFlagsFirst(),
	linter.RequireValidExamples(),

	// Soft Requirements
	linter.RequireLengthBetween("Short", 10, 60),
}

var flagRules = []linter.FlagRule{
	// Hard Requirements
	linter.FlagFilter(
		linter.RequireFlagKebabCase,
		linter.ExcludeFlag(
			"producer.config",
			"consumer.config",
		),
	),
	linter.RequireFlagRealWords('-'),
	linter.FlagFilter(linter.RequireFlagCharacters('-'), linter.ExcludeFlag("consumer.config", "producer.config")),

	linter.FlagFilter(linter.RequireFlagUsageMessage, linter.ExcludeFlag("key-deserializer", "value-deserializer")),
	linter.RequireFlagUsageRealWords,
	linter.RequireFlagUsageStartWithCapital(properNouns),
	linter.FlagFilter(
		linter.RequireFlagUsageEndWithPunctuation,
		linter.ExcludeFlag(
			"batch-size",
			"enable-systest-events",
			"formatter",
			"isolation-level",
			"line-reader",
			"max-block-ms",
			"max-memory-bytes",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"metadata-expiry-ms",
			"offset",
			"property",
			"request-required-acks",
			"request-timeout-ms",
			"retry-backoff-ms",
			"socket-buffer-size",
			"timeout",
			"value-format",
		),
	),

	// Soft Requirements
	linter.FlagFilter(
		linter.RequireFlagNameLength(2, 20),
		linter.ExcludeFlag(
			"destination-bootstrap-server",
			"destination-cluster-id",
			"destination-api-key",
			"destination-api-secret",
			"enable-systest-events",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"request-required-acks",
			"schema-registry-cluster-id",
			"skip-message-on-error",
			"source-bootstrap-server",
		),
	),
	linter.FlagFilter(
		linter.RequireFlagDelimiter('-', 1),
		linter.ExcludeFlag(
			"ca-cert-path",
			"client-cert-path",
			"client-key-path",
			"connect-cluster-id",
			"destination-bootstrap-server",
			"destination-api-key",
			"destination-api-secret",
			"destination-cluster-id",
			"enable-systest-events",
			"if-not-exists",
			"kafka-cluster-id",
			"ksql-cluster-id",
			"local-secrets-file",
			"max-block-ms",
			"max-memory-bytes",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"metadata-expiry-ms",
			"remote-secrets-file",
			"request-required-acks",
			"request-timeout-ms",
			"retry-backoff-ms",
			"schema-registry-cluster-id",
			"skip-message-on-error",
			"socket-buffer-size",
			"source-api-key",
			"source-api-secret",
			"source-bootstrap-server",
			"source-cluster-id",
			"sr-api-key",
			"sr-api-secret",
		),
	),
}

// properNouns are words that don't obey normal capitalization rules
var properNouns = []string{
	"ACL",
	"API",
	"Apache",
	"CLI",
	"Confluent Cloud",
	"Confluent Platform",
	"Confluent",
	"Connect",
	"Control Center",
	"DEPRECATED",
	"IAM",
	"ID",
	"Kafka",
	"RBAC",
	"REST",
	"Schema Registry",
	"ZooKeeper™",
	"ksqlDB Server",
	"ksqlDB",
	"Clients",
	"Clojure",
	"C/C++",
	"C#",
	"Go",
	"Groovy",
	"Java",
	"Kotlin",
	"Ktor",
	"Node.js",
	"Python",
	"Ruby",
	"Rust",
	"Scala",
	"Spring Boot",
}

// vocabWords are words that don't appear in the US dictionary, but are Confluent-related words.
var vocabWords = []string{
	"ack",
	"acks",
	"acl",
	"acls",
	"apac",
	"api",
	"apikey",
	"apisecret",
	"auth",
	"avro",
	"aws",
	"backoff",
	"cku",
	"cli",
	"codec",
	"config",
	"configs",
	"consumer.config",
	"crn",
	"csu",
	"decrypt",
	"deserializer",
	"deserializers",
	"env",
	"eu",
	"failover",
	"formatter",
	"gcp",
	"geo",
	"gzip",
	"hostname",
	"iam",
	"json",
	"jsonschema",
	"kafka",
	"keychain",
	"ksql",
	"lifecycle",
	"lz4",
	"mds",
	"netrc",
	"pem",
	"plaintext",
	"prem",
	"producer.config",
	"protobuf",
	"rbac",
	"readonly",
	"readwrite",
	"recv",
	"sasl",
	"signup",
	"sr",
	"ssl",
	"sso",
	"stdin",
	"systest",
	"tcp",
	"transactional",
	"txt",
	"unregister",
	"url",
	"us",
	"v2",
	"vpc",
	"whitelist",
	"yaml",
	"zstd",
	"clojure",
	"cpp",
	"csharp",
	"kotlin",
	"ktor",
	"nodejs",
	"restapi",
	"scala",
	"springboot",
}

var (
	affFile string
	dicFile string
	debug   bool
)

func init() {
	flag.StringVar(&affFile, "aff-file", "", "hunspell .aff file")
	flag.StringVar(&dicFile, "dic-file", "", "hunspell .dic file")
	flag.BoolVar(&debug, "debug", false, "print debug output")
}

func main() {
	flag.Parse()

	vocab, err := gospell.NewGoSpell(affFile, dicFile)
	if err != nil {
		panic(err)
	}
	for _, word := range vocabWords {
		vocab.AddWordRaw(strings.ToLower(word))
		vocab.AddWordRaw(strings.ToUpper(word))
	}
	linter.SetVocab(vocab)

	l := linter.Linter{
		Rules:     commandRules,
		FlagRules: flagRules,
		Debug:     debug,
	}

	// Lint all three subsets of commands: no context, cloud, and on-prem
	configs := []*v1.Config{
		{
			CurrentContext: "no context",
		},
		{
			Contexts:       map[string]*v1.Context{"cloud": {PlatformName: ccloudv2.Hostnames[0]}},
			CurrentContext: "cloud",
		},
		{
			Contexts:       map[string]*v1.Context{"on-prem": {PlatformName: "https://example.com"}},
			CurrentContext: "on-prem",
		},
	}

	code := 0
	for _, cfg := range configs {
		cmd := pcmd.NewConfluentCommand(cfg, true, new(version.Version))
		if err := l.Lint(cmd.Command); err != nil {
			fmt.Printf(`For context "%s", %v`, cfg.CurrentContext, err)
			code = 1
		}
	}
	os.Exit(code)
}
