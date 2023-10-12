package config

const (
	// ops
	OpSet               = "SET"
	OpUse               = "USE"
	OpReset             = "RESET"
	OpExit              = "EXIT"
	OpUseCatalog        = "CATALOG"
	StatementTerminator = ";"

	// config namespaces
	NamespaceClient = "client."

	// keys
	KeyCatalog        = "sql.current-catalog"
	KeyDatabase       = "sql.current-database"
	KeyLocalTimeZone  = "sql.local-time-zone"
	KeyResultsTimeout = "client.results-timeout"
	KeyServiceAccount = "client.service-account"
	KeyStatementName  = "client.statement-name"
)
