package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/client9/gospell"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/linter"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var commandRules = []linter.CommandRule{
	// Hard Requirements
	linter.RequireLowerCase("Use"),
	linter.RequireRealWords("Use", '-'),
	linter.Filter(linter.RequireSingular("Use"),
		linter.ExcludeCommandContains("local services"),
		linter.ExcludeCommand("kafka client-config create nodejs")),

	linter.Filter(linter.RequireCapitalizeProperNouns("Short", properNouns), linter.ExcludeCommand("local current")),
	linter.RequireEndWithPunctuation("Short", false),
	linter.Filter(linter.RequireNotTitleCase("Short", properNouns), linter.ExcludeCommandContains("ksql app")),
	linter.RequireStartWithCapital("Short"),

	linter.Filter(linter.RequireEndWithPunctuation("Long", true), linter.ExcludeCommand("prompt")),
	linter.Filter(linter.RequireCapitalizeProperNouns("Long", properNouns),
		linter.ExcludeCommand("plugin"),
		linter.ExcludeCommand("completion"),
		linter.ExcludeCommandContains("kafka client-config create"),
		linter.ExcludeCommand("local current")),
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
			"azure-subscription-id",
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
			"aws-account-id",
			"azure-subscription-id",
			"ca-cert-path",
			"client-cert-path",
			"client-key-path",
			"connect-cluster-id",
			"destination-bootstrap-server",
			"destination-api-key",
			"destination-api-secret",
			"destination-cluster-id",
			"enable-systest-events",
			"log-exclude-rows",
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
	"PATH",
	"Python",
	"Ruby",
	"Rust",
	"Scala",
	"Spring Boot",
	"Stream Designer",
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
	"jwks",
	"kafka",
	"keychain",
	"ksql",
	"ksqldb",
	"lifecycle",
	"lkc",
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
	"sql",
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
	"uri",
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
		{CurrentContext: "No Context"},
		{CurrentContext: "Cloud", Contexts: map[string]*v1.Context{"Cloud": {PlatformName: "https://confluent.cloud", State: &v1.ContextState{Auth: &v1.AuthConfig{Organization: &orgv1.Organization{}}}}}},
		{CurrentContext: "On-Prem", Contexts: map[string]*v1.Context{"On-Prem": {PlatformName: "https://example.com"}}},
	}

	code := 0
	for _, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		cmd := pcmd.NewConfluentCommand(cfg)
		if err := l.Lint(cmd); err != nil {
			fmt.Printf(`For context "%s", %v`, cfg.CurrentContext, err)
			code = 1
		}
	}
	os.Exit(code)
}
