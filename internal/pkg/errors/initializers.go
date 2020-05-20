package errors

/*
Backend Error
*/

func NewBackendErrorf(format string, args ...interface{}) *BackendError {
	backendErrorParams.msgFormat = format
	return &BackendError{newCliDefinedErrorImpl(backendErrorParams, args...)}
}

func NewBackendErrorWrapf(err error, format string, args ...interface{}) *BackendError {
	backendErrorParams.wrappedErr = err
	backendErrorParams.msgFormat = format
	return &BackendError{newCliDefinedErrorImpl(backendErrorParams, args...)}
}


/*
Resource Validation Error
*/
func NewResourceValidationErrorf(format string, args ...interface{}) *ResourceValidationError {
	resourceValidationErrorParams.msgFormat = format
	return &ResourceValidationError{newCliDefinedErrorImpl(resourceValidationErrorParams, args...)}
}

func NewResourceValidationErrorWrapf(err error, format string, args ...interface{}) *ResourceValidationError {
	resourceValidationErrorParams.wrappedErr = err
	resourceValidationErrorParams.msgFormat = format
	return &ResourceValidationError{newCliDefinedErrorImpl(resourceValidationErrorParams, args...)}
}


/*
Prohibited Action Error
*/
func NewProhibitedActionErrorf(format string, args ...interface{}) *ProhibitedActionError {
	prohibitedActionErrorParams.msgFormat = format
	return &ProhibitedActionError{newCliDefinedErrorImpl(prohibitedActionErrorParams, args...)}
}

func NewProhibitedActionErrorWrapf(err error, format string, args ...interface{}) *ProhibitedActionError {
	prohibitedActionErrorParams.wrappedErr = err
	prohibitedActionErrorParams.msgFormat = format
	return &ProhibitedActionError{newCliDefinedErrorImpl(prohibitedActionErrorParams, args...)}
}


/*
Flag Use Error
*/
func NewFlagUseErrorf(format string, args ...interface{}) *FlagUseError {
	flagUseErrorParams.msgFormat = format
	return &FlagUseError{newCliDefinedErrorImpl(flagUseErrorParams, args...)}
}

func NewFlagUseErrorWrapf(err error, format string, args ...interface{}) *FlagUseError {
	flagUseErrorParams.wrappedErr = err
	flagUseErrorParams.msgFormat = format
	return &FlagUseError{newCliDefinedErrorImpl(flagUseErrorParams, args...)}
}


/*
Corrupted CLI Config Error
*/
func NewCorruptedCLIConfigErrorf(format string, args ...interface{}) *CorruptedCLIConfigError {
	corruptedCLIConfigErrorParams.msgFormat = format
	return &CorruptedCLIConfigError{newCliDefinedErrorImpl(corruptedCLIConfigErrorParams, args...)}
}

func NewCorruptedCLIConfigErrorWrapf(err error, format string, args ...interface{}) *CorruptedCLIConfigError {
	corruptedCLIConfigErrorParams.wrappedErr = err
	corruptedCLIConfigErrorParams.msgFormat = format
	return &CorruptedCLIConfigError{newCliDefinedErrorImpl(corruptedCLIConfigErrorParams, args...)}
}


/*
CLI State Error
*/
func NewCLIStateErrorf(format string, args ...interface{}) *CLIStateError {
	cliStateErrorParams.msgFormat = format
	return &CLIStateError{newCliDefinedErrorImpl(cliStateErrorParams, args...)}
}

func NewCLIStateErrorWrapf(err error, format string, args ...interface{}) *CLIStateError {
	cliStateErrorParams.wrappedErr = err
	cliStateErrorParams.msgFormat = format
	return &CLIStateError{newCliDefinedErrorImpl(cliStateErrorParams, args...)}
}


/*
Resource Not Ready Error
*/
func NewResourceNotReadyErrof(format string, args ...interface{}) *ResourceNotReadyError {
	resourceNotREadyErrorParams.msgFormat = format
	return &ResourceNotReadyError{newCliDefinedErrorImpl(resourceNotREadyErrorParams, args...)}
}

func NewResourceNotReadyErrorWrapf(err error, format string, args ...interface{}) *ResourceNotReadyError {
	resourceNotREadyErrorParams.wrappedErr = err
	resourceNotREadyErrorParams.msgFormat = format
	return &ResourceNotReadyError{newCliDefinedErrorImpl(resourceNotREadyErrorParams, args...)}
}


/*
Unexpected CLI Behavior
*/
func NewUnexpectedCLIBehaviorErrorf(format string, args ...interface{}) *UnexpectedCLIBehaviorError {
	unexpectedCLIBehaviorErrorParams.msgFormat = format
	return &UnexpectedCLIBehaviorError{newCliDefinedErrorImpl(unexpectedCLIBehaviorErrorParams, args...)}
}

func NewUnexpectedCLIBehaviorErrorWrapf(err error, format string, args ...interface{}) *UnexpectedCLIBehaviorError {
	unexpectedCLIBehaviorErrorParams.wrappedErr = err
	unexpectedCLIBehaviorErrorParams.msgFormat = format
	return &UnexpectedCLIBehaviorError{newCliDefinedErrorImpl(unexpectedCLIBehaviorErrorParams, args...)}
}
