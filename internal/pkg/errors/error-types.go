package errors

/*
List of Error types

1. BackendError
2. ResourceValidationError
3. ProhibitedActionError
4. FlagUseError
5. CLIStateError
6. CorruptedCLIConfigError
7. ResourceNotReadyError
8. UnexpectedCLIBehavior

Errors that don't quite fit in any group should just be put under UnexpectedCLIBehavior error, unless a new error
category should be defined for that type.
*/

var (
	backendErrorPrefix = "Backend Error"
	backendErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       backendErrorPrefix,
	}

	resourceValidationErrorPrefix = "Resource Validation Error"
	resourceValidationErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       resourceValidationErrorPrefix,
	}


	prohibitedActionErrorPrefix = "Prohibited Action"
	prohibitedActionErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       prohibitedActionErrorPrefix,
	}

	cliStateErrorPrefix = "CLI State"
	cliStateErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       cliStateErrorPrefix,
	}

	corruptedCLIConfigErrorPrefix = "Corrupted CLI Config"
	corruptedCLIConfigErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       corruptedCLIConfigErrorPrefix,
	}

	flagUseErrorPrefix = "Incorrect Flag Use"
	flagUseErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: false,
		prefix:       flagUseErrorPrefix,
	}

	resourceNotReadyErrorPrefix = "Resource Not Ready"
	resourceNotREadyErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       resourceNotReadyErrorPrefix,
	}

	unexpectedCLIBehaviorErrorPrefix = "Unexpected CLI Behavior"
	unexpectedCLIBehaviorErrorParams = &cliDefinedErrorImplParams{
		silenceUsage: true,
		prefix:       unexpectedCLIBehaviorErrorPrefix,
	}
)

type CLIDefinedError interface {
	error
	SilenceUsage() bool
	SetDirectionsMsg(format string, args ...interface{})
	GetDirectionsMsg() string
}

/*
Errors from the backend that are not yet defined under CommonBackendBugError, as they are not yet known.
*/
type BackendError struct {
	*cliDefinedErrorImpl
}

/*
Errors thrown when users specify an invalid resource. e.g. resource not found, do not have access to that resource
*/
type ResourceValidationError struct {
	*cliDefinedErrorImpl
}

/*
ACL and permission related stuff...
Errors thrown when the users does not have permission to perform the action.
*/
type ProhibitedActionError struct {
	*cliDefinedErrorImpl
}

/*
Flag related erros.
*/
type FlagUseError struct {
	*cliDefinedErrorImpl
}

/*
Errors thrown when the command fails because of the CLI state even though the command is called correctly.
e.g. no API key selected for a cluster that the user is trying to produce from, user not logged in
*/
type CLIStateError struct {
	*cliDefinedErrorImpl
}

/*
CLI config file is corrupted.
*/
type CorruptedCLIConfigError struct  {
	*cliDefinedErrorImpl
}

/*
Resource is not yet read, need to let user know that they may have to wait.
*/
type ResourceNotReadyError struct  {
	*cliDefinedErrorImpl
}

/*
CLI unexpected bugs. e.g. client is null
*/
type UnexpectedCLIBehaviorError struct {
	*cliDefinedErrorImpl
}
