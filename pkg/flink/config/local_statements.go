package config

const (
	// ops
	ConfigOpSet               = "SET"
	ConfigOpUse               = "USE"
	ConfigOpReset             = "RESET"
	ConfigOpExit              = "EXIT"
	ConfigOpUseCatalog        = "CATALOG"
	ConfigStatementTerminator = ";"

	// keys
	ConfigKeyCatalog        = "sql.current-catalog"
	ConfigKeyDatabase       = "sql.current-database"
	ConfigKeyLocalTimeZone  = "sql.local-time-zone"
	ConfigKeyResultsTimeout = "client.results-timeout"
)
