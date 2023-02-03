package cmd

import "github.com/spf13/pflag"

func OnPremKafkaRestSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("onprem-kafkarest", pflag.ExitOnError)
	set.String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.")
	set.String("ca-cert-path", "", "Path to a PEM-encoded CA to verify the Confluent REST Proxy.")
	set.String("client-cert-path", "", "Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.")
	set.String("client-key-path", "", "Path to client private key, include for mTLS authentication.")
	set.Bool("no-authentication", false, "Include if requests should be made without authentication headers, and user will not be prompted for credentials.")
	set.Bool("prompt", false, "Bypass use of available login credentials and prompt for Kafka Rest credentials.")
	set.SortFlags = false
	return set
}

func OnPremAuthenticationSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("onprem-authentication", pflag.ExitOnError)
	set.String("bootstrap", "", `Comma-separated list of broker hosts, each formatted as "host" or "host:port".`)
	set.String("ca-location", "", "File or directory path to CA certificate(s) for SSL verifying the broker's key.")
	set.String("username", "", "SASL_SSL username for use with PLAIN mechanism.")
	set.String("password", "", "SASL_SSL password for use with PLAIN mechanism.")
	set.String("cert-location", "", "Path to client's public key (PEM) used for SSL authentication.")
	set.String("key-location", "", "Path to client's private key (PEM) used for SSL authentication.")
	set.String("key-password", "", "Private key passphrase for SSL authentication.")
	set.SortFlags = false
	return set
}

func OnPremSchemaRegistrySet() *pflag.FlagSet {
	set := pflag.NewFlagSet("onprem-schemaregistry", pflag.ExitOnError)
	set.String("ca-location", "", "File or directory path to CA certificate(s) to authenticate the Schema Registry client.")
	set.String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")
	set.SortFlags = false
	return set
}
