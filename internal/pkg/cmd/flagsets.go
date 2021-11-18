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
	set.String("context", "", "CLI context name.")
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
	set.Bool("prompt", false, "Bypass use of available login credentials and prompt for Kafka Rest credentials.")
	set.SortFlags = false
	return set
}

func OnPremAuthenticationSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("onprem-authentication", pflag.ExitOnError)
	set.String("bootstrap", "", "List of broker host or host:port.")
	set.String("protocol", "", "Security protocol used to communicate with brokers.")
	set.String("username", "", "sasl_ssl username for use with PLAIN mechanism.")
	set.String("password", "", "sasl_ssl password for use with PLAIN mechanism.")
	set.Bool("ssl-verification", false, "Enable OpenSSL's builtin broker (server) certificate verification.")
	set.String("ca-location", "", "File or directory path to CA certificate(s) for ssl verifying the broker's key.")
	set.String("cert-location", "", "Path to client's public key (PEM) used for ssl authentication.")
	set.String("key-location", "", "Path to client's private key (PEM) used for ssl authentication.")
	set.String("key-password", "", "Private key passphrase for ssl authentication.")
	set.SortFlags = false
	return set
}

func CombineFlagSet(flagSet *pflag.FlagSet, toAdd ...*pflag.FlagSet) *pflag.FlagSet {
	for _, set := range toAdd {
		flagSet.AddFlagSet(set)
	}
	return flagSet
}
