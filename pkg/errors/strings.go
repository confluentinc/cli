package errors

const (
	CreatedResourceMsg         = "Created %s \"%s\".\n"
	DeleteResourceConfirmMsg   = "Are you sure you want to delete %s \"%s\"?\nTo confirm, type \"%s\". To cancel, press Ctrl-C"
	LoggedInAsMsg              = "Logged in as \"%s\".\n"
	LoggedInAsMsgWithOrg       = "Logged in as \"%s\" for organization \"%s\" (\"%s\").\n"
	LoggedInUsingEnvMsg        = "Using environment \"%s\".\n"
	StartingConsumerMsg        = "Starting Kafka Consumer. Use Ctrl-C to exit."
	UndeleteResourceConfirmMsg = "Are you sure you want to undelete %s \"%s\"?\nTo confirm, type \"%s\". To cancel, press Ctrl-C."
	UnsetResourceMsg           = "Unset %s \"%s\".\n"
	UpdateSuccessMsg           = "Updated the %s of %s \"%s\" to \"%v\".\n"
	UpdatedResourceMsg         = "Updated %s \"%s\".\n"
	UsingResourceMsg           = "Using %s \"%s\".\n"
)
