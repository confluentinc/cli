package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/client9/gospell"

	"github.com/confluentinc/cli/v3/internal"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/linter"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
)

var commandRules = []linter.CommandRule{
	// Hard Requirements
	linter.RequireLowerCase("Use"),
	linter.RequireRealWords("Use", '-'),
	linter.Filter(linter.RequireSingular("Name"),
		linter.ExcludeCommandContains("local services"),
		linter.ExcludeCommand("kafka client-config create nodejs")),

	linter.Filter(linter.RequireCapitalizeProperNouns("Short", properNouns), linter.ExcludeCommand("local current")),
	linter.RequireEndWithPunctuation("Short", false),
	linter.Filter(linter.RequireNotTitleCase("Short", properNouns)),
	linter.RequireStartWithCapital("Short"),

	linter.Filter(linter.RequireEndWithPunctuation("Long", true), linter.ExcludeCommand("prompt")),
	linter.Filter(linter.RequireCapitalizeProperNouns("Long", properNouns),
		linter.ExcludeCommand("plugin"),
		linter.ExcludeCommand("completion"),
		linter.ExcludeCommandContains("kafka client-config create"),
		linter.ExcludeCommandContains("local services kafka start"),
		linter.ExcludeCommand("local current")),
	linter.RequireStartWithCapital("Long"),

	linter.RequireListRequiredFlagsFirst(),
	linter.RequireValidExamples(),

	// Soft Requirements
	linter.Filter(linter.RequireLengthBetween("Short", 10, 60),
		linter.ExcludeCommand("audit-log config edit"),
		linter.ExcludeCommand("audit-log config update")),
}

var flagRules = []linter.FlagRule{
	// Hard Requirements
	linter.FlagFilter(linter.RequireFlagKebabCase, linter.ExcludeFlag("producer.config", "consumer.config")),
	linter.RequireFlagRealWords('-'),
	linter.FlagFilter(linter.RequireFlagCharacters('-'), linter.ExcludeFlag("consumer.config", "producer.config")),
	linter.FlagFilter(linter.RequireStringSlicePrefix, linter.ExcludeFlag("property")),

	linter.FlagFilter(linter.RequireFlagUsageMessage, linter.ExcludeFlag("key-deserializer", "value-deserializer")),
	linter.RequireFlagUsageRealWords(properNouns),
	linter.RequireFlagUsageCapitalized(properNouns),
	linter.FlagFilter(
		linter.RequireFlagUsageEndWithPunctuation,
		linter.ExcludeFlag(
			"batch-size",
			"enable-systest-events",
			"formatter",
			"isolation-level",
			"key-deserializer",
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
			"value-deserializer",
			"value-format",
		),
	),

	// Soft Requirements
	linter.FlagFilter(
		linter.RequireFlagNameLength(2, 20),
		linter.ExcludeFlag(
			"azure-subscription-id",
			"destination-api-key",
			"destination-api-secret",
			"destination-bootstrap-server",
			"destination-cluster",
			"remote-api-key",
			"remote-api-secret",
			"remote-bootstrap-server",
			"remote-cluster",
			"enable-systest-events",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"request-required-acks",
			"schema-registry-api-key",
			"schema-registry-api-secret",
			"schema-registry-cluster",
			"schema-registry-context",
			"schema-registry-endpoint",
			"schema-registry-subjects",
			"skip-message-on-error",
			"source-bootstrap-server",
			"update-schema-registry",
			"worker-configurations",
		),
	),
	linter.FlagFilter(
		linter.RequireFlagDelimiter('-', 1),
		linter.ExcludeFlag(
			"aws-account-id",
			"azure-subscription-id",
			"ca-cert-path",
			"client-cert-path",
			"client-key-path",
			"destination-api-key",
			"destination-api-secret",
			"destination-bootstrap-server",
			"remote-api-key",
			"remote-api-secret",
			"local-api-key",
			"local-api-secret",
			"remote-bootstrap-server",
			"enable-systest-events",
			"gcp-project-id",
			"if-not-exists",
			"kafka-api-key",
			"kafka-rest-port",
			"local-secrets-file",
			"log-exclude-rows",
			"max-block-ms",
			"max-memory-bytes",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"metadata-expiry-ms",
			"remote-secrets-file",
			"request-required-acks",
			"request-timeout-ms",
			"retry-backoff-ms",
			"schema-registry-api-key",
			"schema-registry-api-secret",
			"schema-registry-cluster",
			"schema-registry-context",
			"schema-registry-endpoint",
			"schema-registry-subjects",
			"skip-message-on-error",
			"socket-buffer-size",
			"source-api-key",
			"source-api-secret",
			"source-bootstrap-server",
			"update-schema-registry",
			"use-schema-registry",
		),
	),
}

// properNouns are words that don't obey normal capitalization rules
var properNouns = []string{
	"ACLs",
	"Apache",
	"Async",
	"AsyncAPI",
	"Avro",
	"C#",
	"C/C++",
	"CFU",
	"Clients",
	"Clojure",
	"Confluent Cloud",
	"Confluent Platform",
	"Confluent",
	"Connect",
	"Control Center",
	"Flink",
	"Go",
	"Groovy",
	"Java",
	"Kafka",
	"Kotlin",
	"Ktor",
	"Node.js",
	"Python",
	"Ruby",
	"Rust",
	"Scala",
	"Schema Registry",
	"Spring Boot",
	"Stream Designer",
	"ZooKeeper™",
	"ksqlDB Server",
	"ksqlDB",
}

// vocabWords are words that don't appear in the US dictionary, but are Confluent-related words.
var vocabWords = []string{
	"ack",
	"acks",
	"acl",
	"acls",
	"apac",
	"api",
	"asyncapi",
	"auth",
	"avro",
	"aws",
	"backoff",
	"byok",
	"cel",
	"cfu",
	"cku",
	"cli",
	"clojure",
	"codec",
	"config",
	"configs",
	"consumer.config",
	"cpp",
	"crn",
	"csharp",
	"csu",
	"decrypt",
	"deserializer",
	"deserializers",
	"env",
	"eu",
	"failover",
	"filepath",
	"flink",
	"formatter",
	"gcp",
	"geo",
	"gzip",
	"hostname",
	"iam",
	"json",
	"jsonschema",
	"jwks",
	"kafka",
	"keychain",
	"kotlin",
	"ksql",
	"ksqldb",
	"ktor",
	"lifecycle",
	"lkc",
	"lz4",
	"mds",
	"netrc",
	"nodejs",
	"pem",
	"plaintext",
	"prem",
	"producer.config",
	"protobuf",
	"rbac",
	"readonly",
	"readwrite",
	"recv",
	"restapi",
	"ruleset",
	"sasl",
	"scala",
	"schemas",
	"signup",
	"springboot",
	"sql",
	"ssl",
	"sso",
	"stdin",
	"systest",
	"tcp",
	"transactional",
	"txt",
	"unregister",
	"uri",
	"url",
	"us",
	"v2",
	"vv",
	"vvv",
	"vvvv",
	"vpc",
	"whitelist",
	"yaml",
	"zstd",
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
	configs := []*config.Config{
		{CurrentContext: "No Context"},
		{CurrentContext: "Cloud", Contexts: map[string]*config.Context{"Cloud": {PlatformName: "https://confluent.cloud"}}},
		{CurrentContext: "On-Prem", Contexts: map[string]*config.Context{"On-Prem": {PlatformName: "https://example.com"}}},
	}

	code := 0
	for _, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		cmd := internal.NewConfluentCommand(cfg)
		if err := l.Lint(cmd); err != nil {
			fmt.Printf(`For context "%s", %v`, cfg.CurrentContext, err)
			code = 1
		}
	}
	os.Exit(code)
}
