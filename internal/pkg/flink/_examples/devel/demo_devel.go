package main

import (
	"github.com/confluentinc/flink-sql-client/pkg/app"
	"github.com/confluentinc/flink-sql-client/pkg/types"
)

func main() {
	integration := "LHLM6WR52IF5EPFR:1PztlitmLzoogth7i7iYopfAxvMykU14+CYHr/ViEvSgPHCDPPRfS1kpkLIcAK7R"
	integrationPlayback := "PCWFUBHIAZ3YYIUP:VzTDEnhFAsKH9MzZrJRdSACWp3yrYTPUF39kflXpTW/euRNQcB8UTopCMNMLfM6v"
	integrationWebsite := "3HENPABCYN7WKU4A:x/cOHMjBO7BdMeFVIfzO/9n8w1ik4ryA49GkisAjQBr0EZ4uCjIAzQ211/Pom+c3"
	metrics := "ZC632C3FY2WCLQ6Z:kY/v0TAw1qBPrtpkYsaqFlAe5fdBGh3M5hYldCzvkUK065bqwiPEU8a7BsaTjkbv"
	metricsQos := "XPLV5KXN42IA6OWJ:ZXuXs6osL4uSayWh5i8SeU9Js8wNBpzQu6cMbiixdu6KJQEOciRKOFkox0m8frei"
	engagement := "VOUPSZQWODCAU7X6:b3P8sfTFNo54IeZCLPznNjmpnFxzoyjJ0VqTkGsDUKIwJWqK7H14QmDMhr0FMPyE"
	engagementCoreMetrics := "KBIHCYSML52E4DME:AiNdyjhWs9oL2kIc0XoUusHPygeW01gJMkUiKCxrhNpZ78ow2j8EBP8EZ+jUr0hD"
	engagementPersonalization := "KG3K2LCJEVD6T7YH:jBaU6oiBEAV25fhOzQb/UEnVW6qmEqLdJpfiTjdSHSDUdsJUWlnxk/v7RoKgldHA"
	app.StartApp(
		"env-okyxpp",
		"d07e2cb7-52df-451b-bf4b-b8bd81e0d20c",
		"playback",
		"lfcp-10vygz",
		func() string { return "authToken" },
		func() error { return nil },
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
				"catalog":                "integration",
				"confluent.kafka.keys": "playback:" + integrationPlayback + ";lkc-j0m02p:" + integrationPlayback + ";" +
					"website:" + integrationWebsite + ";lkc-rkmkgp:" + integrationWebsite + ";" +
					"qos:" + metricsQos + ";lkc-zn7n9y:" + metricsQos + ";" +
					"core_metrics:" + engagementCoreMetrics + ";lkc-30p0m0:" + engagementCoreMetrics + ";" +
					"personalization:" + engagementPersonalization + ";lkc-n0m096:" + engagementPersonalization,
				"confluent.schema_registry.keys": "integration:" + integration + ";" +
					"metrics:" + metrics + ";" +
					"engagement:" + engagement,
			},
		})
}
