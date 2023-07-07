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
	ConfigKeyCatalog          = "catalog"
	ConfigKeyDatabase         = "default_database"
	ConfigKeyOrgResourceId    = "org-resource-id"
	ConfigKeyExecutionRuntime = "execution.runtime-mode"
	ConfigKeyLocalTimeZone    = "table.local-time-zone"
	ConfigKeyResultsTimeout   = "table.results-timeout"
	ConfigKeyStatementOwner   = "statement-owner"
)
