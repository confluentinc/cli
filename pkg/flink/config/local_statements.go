package config

const (
	// ops
	ConfigOpSet               = "SET"
	ConfigOpUse               = "USE"
	ConfigOpReset             = "RESET"
	ConfigOpExit              = "EXIT"
	ConfigOpUseCatalog        = "CATALOG"
	ConfigStatementTerminator = ";"

	// config namespaces
	ConfigNamespaceSql    = "sql."
	ConfigNamespaceClient = "client."

	// keys
	ConfigKeyCatalog        = "sql.current-catalog"
	ConfigKeyDatabase       = "sql.current-database"
	ConfigKeyLocalTimeZone  = "sql.local-time-zone"
	ConfigKeyResultsTimeout = "client.results-timeout"
	ConfigKeyServiceAccount = "client.service-account"
	ConfigKeyStatementName  = "client.statement-name"
)
