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

	/*
	   Create Alter model regex. Match create|alter model with
	     1. allow any space,return,tab between them
	     2. They are not surrounded by unescaped single quote
	     3. Single quote is escaped by doubling them like ''.
	     4. Start beginning of line, with space or `;`

	     (?msi) -> enable multiline, . match \n and case insensitive
	     \s -> matches space, \n, \t, \r, \f
	     '' -> match escaped single quote
	     ?: -> ignore group
	     '(?:''|[^'])*' -> match anything inside single quote including escaped quote
	     ?P<query> -> give group match a name
	     (?:^|;|\\s)(?:create|alter)\\s+model\\s+ -> start with beginning of line, spaces or ;. Match create or alter, then multiple \s, then model, then multiple \s

	     Reference: https://github.com/google/re2/wiki/Syntax
	                https://pkg.go.dev/regexp#Regexp.FindSubmatch
	                https://stackoverflow.com/questions/6462578/regex-to-match-all-instances-not-inside-quotes
	*/
	CreateAlterModelRegex = "(?msi)''|'(?:''|[^'])*'|(?P<query>(?:^|;|\\s)(?:create|alter)\\s+model\\s+)"
)

type OutputFormat string

const (
	OutputFormatStandard  OutputFormat = "standard"
	OutputFormatPlainText OutputFormat = "plain-text"
)
