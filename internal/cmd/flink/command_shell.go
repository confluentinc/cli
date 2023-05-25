package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	client "github.com/confluentinc/cli/internal/pkg/flink/app"
	"github.com/confluentinc/cli/internal/pkg/flink/types"

	"github.com/spf13/cobra"
)

func (c *command) newShellCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd, args)
		},
	}
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	cmd.Flags().String("kafka-cluster", "", "Kafka cluster ID.")
	cmd.Flags().Bool("demo-devel", false, "Start client against the hardcoded demo devel pod.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command) func() error {
	return func() error {
		cfg, err := load.LoadAndMigrate(v1.New())
		if err != nil {
			return err
		}
		auth := cfg.Context().State.AuthToken
		authRefreshToken := cfg.Context().State.AuthRefreshToken
		err = c.Context.UpdateAuthTokens(auth, authRefreshToken)
		if err != nil {
			return err
		}
		return authenticated(cmd, nil)
	}
}

func (c *command) startFlinkSqlClient(prerunner pcmd.PreRunner, cmd *cobra.Command, args []string) error {

	demoDevel, err := cmd.Flags().GetBool("demo-devel")
	if err != nil {
		return err
	}

	if demoDevel {
		integration := "LHLM6WR52IF5EPFR:1PztlitmLzoogth7i7iYopfAxvMykU14+CYHr/ViEvSgPHCDPPRfS1kpkLIcAK7R"
		integrationPlayback := "PCWFUBHIAZ3YYIUP:VzTDEnhFAsKH9MzZrJRdSACWp3yrYTPUF39kflXpTW/euRNQcB8UTopCMNMLfM6v"
		integrationWebsite := "3HENPABCYN7WKU4A:x/cOHMjBO7BdMeFVIfzO/9n8w1ik4ryA49GkisAjQBr0EZ4uCjIAzQ211/Pom+c3"
		qos := "ZC632C3FY2WCLQ6Z:kY/v0TAw1qBPrtpkYsaqFlAe5fdBGh3M5hYldCzvkUK065bqwiPEU8a7BsaTjkbv"
		qosQosMetrics := "XPLV5KXN42IA6OWJ:ZXuXs6osL4uSayWh5i8SeU9Js8wNBpzQu6cMbiixdu6KJQEOciRKOFkox0m8frei"
		engagement := "VOUPSZQWODCAU7X6:b3P8sfTFNo54IeZCLPznNjmpnFxzoyjJ0VqTkGsDUKIwJWqK7H14QmDMhr0FMPyE"
		engagementCoreMetrics := "KBIHCYSML52E4DME:AiNdyjhWs9oL2kIc0XoUusHPygeW01gJMkUiKCxrhNpZ78ow2j8EBP8EZ+jUr0hD"
		engagementPersonalization := "KG3K2LCJEVD6T7YH:jBaU6oiBEAV25fhOzQb/UEnVW6qmEqLdJpfiTjdSHSDUdsJUWlnxk/v7RoKgldHA"
		generator := "ODZPNNZ7VG53GGSW:yyESkAz0NpuMWYqfMDnR+blFXqbssjrIszpem4EC+rFbyiGITSrPXsX6KTS0oYZi"
		generatorSpaTV := "LP3IM5JSUYMZWTRD:OtjtKb3LbktK+tuyupPwQqKK2IZqgckramG3Yi1+gKXF7ZAj2VueowHcXOIJXZdr"
		client.StartApp(
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
						"qos_metrics:" + qosQosMetrics + ";lkc-zn7n9y:" + qosQosMetrics + ";" +
						"core_metrics:" + engagementCoreMetrics + ";lkc-30p0m0:" + engagementCoreMetrics + ";" +
						"personalization:" + engagementPersonalization + ";lkc-n0m096:" + engagementPersonalization + ";" +
						"spatv:" + generatorSpaTV + ";lkc-j06opm:" + generatorSpaTV,
					"confluent.schema_registry.keys": "integration:" + integration + ";" +
						"qos:" + qos + ";" +
						"engagement:" + engagement + ";" +
						"generator:" + generator,
				},
			})
	}

	resourceId := c.Context.GetOrganization().GetResourceId()

	// Compute pool can be set as a flag or as default in the context
	computePool, err := cmd.Flags().GetString("compute-pool")
	if computePool == "" || err != nil {
		if c.Context.GetCurrentFlinkComputePool() == "" {
			return errors.NewErrorWithSuggestions("No compute pool set", "Please set a compute pool to be used. You can either set a default persitent compute pool \"confluent flink compute-pool use lfc-123\" or pass the flag \"--compute-pool lfcp-12345\".")
		} else {
			computePool = c.Context.GetCurrentFlinkComputePool()
		}
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
	if kafkaCluster == "" || err != nil {
		if c.Context.KafkaClusterContext.GetActiveKafkaClusterId() != "" {
			kafkaCluster = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		}
	}

	enviromentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client.StartApp(enviromentId, resourceId, kafkaCluster, computePool, c.AuthToken,
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd),
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
			USER_AGENT: c.Version.UserAgent,
		})
	return nil
}
