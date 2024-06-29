package config

const (
	// ops
	OpSet               = "SET"
	OpUse               = "USE"
	OpReset             = "RESET"
	OpExit              = "EXIT"
	OpQuit              = "QUIT"
	OpUseCatalog        = "CATALOG"
	StatementTerminator = ";"

	// config namespaces
	NamespaceClient = "client."

	// keys
	KeyCatalog        = "sql.current-catalog"
	KeyDatabase       = "sql.current-database"
	KeyLocalTimeZone  = "sql.local-time-zone"
	KeySqlSecrets     = "sql.secrets."
	KeyResultsTimeout = "client.results-timeout"
	KeyServiceAccount = "client.service-account"
	KeyStatementName  = "client.statement-name"
	KeyOutputFormat   = "client.output-format"
)

type OutputFormat string

const (
	OutputFormatStandard  OutputFormat = "standard"
	OutputFormatPlainText OutputFormat = "plain-text"
)
