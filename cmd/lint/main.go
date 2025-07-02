package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/client9/gospell"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/linter"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
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
	linter.Filter(linter.RequireValidExamples(),
		linter.ExcludeCommand("connect custom-plugin version create"),
		linter.ExcludeCommand("connect custom-plugin version update"),
		linter.ExcludeCommand("pipeline update"),
		linter.ExcludeCommand("flink statement update")),

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
			"config",
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
			"accepted-environments",
			"add-operation-groups",
			"azure-subscription",
			"certificate-authority-path",
			"certificate-chain-filename",
			"compute-pool-defaults",
			"destination-api-key",
			"destination-api-secret",
			"destination-bootstrap-server",
			"destination-cluster",
			"enable-systest-events",
			"encrypted-key-material",
			"include-parent-scopes",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"private-link-access-point",
			"record-failure-strategy",
			"remote-api-key",
			"remote-api-secret",
			"remote-bootstrap-server",
			"remote-cluster",
			"remove-operation-groups",
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
		linter.RequireFlagDelimiter('-', 2),
		linter.ExcludeFlag(
			"aws-ram-share-arn",
			"max-partition-memory-bytes",
			"message-send-max-retries",
			"private-link-access-point",
			"schema-registry-api-key",
			"schema-registry-api-secret",
			"skip-message-on-error",
		),
	),
}

// properNouns are words that don't obey normal capitalization rules
var properNouns = []string{
	"ACLs",
	"AI",
	"Apache",
	"Async",
	"AsyncAPI",
	"Avro",
	"C#",
	"C/C++",
	"CIDR",
	"CFU",
	"Clients",
	"Clojure",
	"Confluent Cloud",
	"Confluent Local",
	"Confluent Platform",
	"Confluent",
	"Connect",
	"Control Center",
	"CRL",
	"Data Encryption Key",
	"DEK",
	"Flink",
	"Go",
	"Groovy",
	"Java",
	"Kafka",
	"Key Encryption Key",
	"KEK",
	"Kotlin",
	"KRaft Controller",
	"Ktor",
	"Kubernetes",
	"Node.js",
	"Python",
	"Ruby",
	"Rust",
	"Scala",
	"Schema Registry",
	"Spring Boot",
	"Stream Designer",
	"Tableflow",
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
	"ai",
	"apac",
	"api",
	"apis",
	"arn",
	"asyncapi",
	"auth",
	"avro",
	"aws",
	"azureml",
	"azureopenai",
	"backoff",
	"base64",
	"bedrock",
	"byok",
	"byos",
	"ccpm",
	"cel",
	"cfu",
	"cidr",
	"cku",
	"cli",
	"clojure",
	"cmf",
	"codec",
	"config",
	"configs",
	"consumer.config",
	"confluent_jdbc",
	"couchbase",
	"cpp",
	"crl",
	"crn",
	"csharp",
	"csu",
	"decrypt",
	"dek",
	"deserializer",
	"deserializers",
	"dns",
	"elastic",
	"env",
	"eu",
	"failover",
	"filepath",
	"flink",
	"formatter",
	"gcm",
	"gcp",
	"geo",
	"googleai",
	"gzip",
	"hostname",
	"https",
	"html",
	"iam",
	"io",
	"ip",
	"ips",
	"jdbc",
	"json",
	"jit",
	"jsonschema",
	"jwks",
	"JWT",
	"kafka",
	"kek",
	"keychain",
	"kms",
	"kotlin",
	"kraft",
	"ksql",
	"ksqldb",
	"ktor",
	"kubernetes",
	"librdkafka",
	"lifecycle",
	"lkc",
	"lz4",
	"md",
	"mds",
	"mongodb",
	"namespace",
	"nodejs",
	"openai",
	"pem",
	"pinecone",
	"plaintext",
	"prem",
	"privatelink",
	"producer.config",
	"protobuf",
	"rbac",
	"readonly",
	"readwrite",
	"recv",
	"rescale",
	"restapi",
	"ruleset",
	"s3",
	"sagemaker",
	"sasl",
	"scala",
	"schemas",
	"signup",
	"siv",
	"springboot",
	"sql",
	"ssl",
	"sso",
	"subresource",
	"stdin",
	"systest",
	"tableflow",
	"tcp",
	"transactional",
	"transitgateway",
	"txt",
	"ui",
	"undelete",
	"undeleted",
	"unregister",
	"uri",
	"url",
	"us",
	"v2",
	"vertexai",
	"vv",
	"vvv",
	"vvvv",
	"vnet",
	"vpc",
	"whitelist",
	"yaml",
	"yml",
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
