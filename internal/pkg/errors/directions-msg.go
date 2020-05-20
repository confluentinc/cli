package errors

var (
	APIKeyStoreRefuseToOverrideExistingSecretDirectionsMsg = "If you would like to override the existing secret stored for API key \"%s\", please use `--force` flag."

	ResolveResourceIDResourceNotFoundDirectionsMsg            = "Please check the ID string \"%s\" for typos, and verify that the resource you are looking for exists."
	KafkaClusterCreateCloudRegionNotAvailableDirectionsMsg    = "You can view a list of available regions for \"%s\" with `ccloud kafka region list --cloud %s` command."
	KafkaClusterCreateCloudProviderNotAvailableDirectionsMsg = "You can view a list of available cloud providers and regions with the `ccloud kafka region list` command."
)
