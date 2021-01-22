package errors

const (
	// api commands
	APIKeyNotRetrievableMsg = "Save the API key and secret. The secret is not retrievable later."
	APIKeyTime              = "It may take a couple of minutes for the API key to be ready."

	// kafka commands
	KafkaClusterTime = "It may take up to 5 minutes for the Kafka cluster to be ready."

	// secret commands
	SaveTheMasterKeyMsg = "Save the master key. It cannot be retrieved later."

	//login command
	UsingLoginURLDefaults = "Assuming %s.\n"

	// ksql create warning
	KSQLCreateDeprecateWarning = "(DEPRECATED) In a future release, api-key and api-secret will be required flags when creating a ksql cluster."

	// audit log migration
	OtherCategoryWarning = "\\“Other\\” Category Warning: The OTHER event category rule from the route %q " +
		"for cluster %q has been dropped because it contains a MANAGEMENT event category. The OTHER event " +
		"category is deprecated in Confluent Platform 6.0, and is replaced by the MANAGEMENT event category."
	MultipleCRNWarning = "Multiple CRN Authorities Warning: Cluster %q had multiple CRN authorities " +
		"in its routes: %v. Multiple, different CRN authorities exist in routes from a single cluster. " +
		"This is unexpected in a configuration targeting a single cluster, but makes sense if you are reusing " +
		"the same routing rules on multiple clusters. If this is the case you can ignore this warning or consider " +
		"using CRN patterns with wildcard (empty) authority values in your audit log routes."
	MismatchedKafkaClusterWarning = "Mismatched Kafka Cluster Warning: Cluster %q has a route for a different cluster, " +
		"route: %q. Routes from one Kafka cluster ID on a completely different cluster ID are unexpected, " +
		"but not necessarily wrong. For example, this message might be returned if you reuse the same routing " +
		"configuration on multiple clusters."
	NewBootstrapWarning = "New Bootstrap Servers Warning: Cluster %q currently has bootstrap " +
		"servers = %v. Replacing with %v. Migrated clusters will use the specified bootstrap servers."
	MalformedConfigError = "Bad Input File: The audit log configuration for cluster %q " +
		"uses invalid JSON. Parsing error: %v"
	RepeatedRouteWarning = "Repeated Route Warning: Route Name : %q. There are duplicate routes specified " +
		"between different router configurations. Duplicate routes will be dropped."
	NewExcludedPrincipalsWarning = "New Excluded Principals Warning: Due to combining the excluded principals from " +
		"every input cluster, cluster %q will now also exclude the following principals: %v"
	RetentionTimeDiscrepancyWarning = "Retention Time Discrepancy Warning: Topic %q had discrepancies in retention time." +
		" Using max: %v. Discrepancies in retention time occur when two cluster configurations have the same topic in a" +
		" router configuration, but different retention times. The maximum specified retention time will be used."
)
