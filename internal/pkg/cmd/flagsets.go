package cmd

import "github.com/spf13/pflag"

func EnvironmentSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("environment state", pflag.ExitOnError)
	set.String("environment", "", "Environment ID.")
	set.SortFlags = false
	return set
}

func ClusterSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("cluster state", pflag.ExitOnError)
	set.String("cluster", "", "Kafka cluster ID.")
	set.SortFlags = false
	return set
}

func ContextSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("context state", pflag.ExitOnError)
	set.String("context", "", "CLI Context name.")
	set.SortFlags = false
	return set
}

func EnvironmentContextSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("env-context state", pflag.ExitOnError)
	set.AddFlagSet(EnvironmentSet())
	set.AddFlagSet(ContextSet())
	set.SortFlags = false
	return set
}

func ClusterEnvironmentContextSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("cluster-env-context state", pflag.ExitOnError)
	set.AddFlagSet(EnvironmentSet())
	set.AddFlagSet(ClusterSet())
	set.AddFlagSet(ContextSet())
	set.SortFlags = false
	return set
}

func KeySecretSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("key-secret", pflag.ExitOnError)
	set.String("api-key", "", "API key.")
	set.String("api-secret", "", "API key secret.")
	set.SortFlags = false
	return set
}

func OnPremKafkaRestSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("onprem-kafkarest", pflag.ExitOnError)
	set.String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.")
	set.String("ca-cert-path", "", "Path to a PEM-encoded CA to verify the Confluent REST Proxy.")
	set.String("client-cert-path", "", "Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.")
	set.String("client-key-path", "", "Path to client private key, include for mTLS authentication.")
	set.Bool("no-auth", false, "Include if requests should be made without authentication headers, and user will not be prompted for credentials.")
	set.SortFlags = false
	return set
}

func CombineFlagSet(flagSet *pflag.FlagSet, toAdd ...*pflag.FlagSet) *pflag.FlagSet {
	for _, set := range toAdd {
		flagSet.AddFlagSet(set)
	}
	return flagSet
}
