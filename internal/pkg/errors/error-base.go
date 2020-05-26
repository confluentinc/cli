package errors

import (
	"fmt"
)

var (
	errorFormat = "[%s] %s"
	wrapErrorFormat = "[%s] %s: %s"

)

func newCliDefinedErrorImpl(params *cliDefinedErrorImplParams, args ...interface{}) *cliDefinedErrorImpl {
	return &cliDefinedErrorImpl{
		msg:          constructErrorMessage(params, args...),
		silenceUsage: params.silenceUsage,

	}
}

func constructErrorMessage(params *cliDefinedErrorImplParams, args ...interface{}) string {
	var msg string
	if params.wrappedErr == nil {
		msg = fmt.Sprintf(errorFormat, params.prefix, params.msgFormat)
		msg = fmt.Sprintf(msg, args...)
	} else {
		msg = fmt.Sprintf(wrapErrorFormat, params.prefix, params.msgFormat, params.wrappedErr.Error())
		msg = fmt.Sprintf(msg, args...)
	}
	return msg
}

type cliDefinedErrorImplParams struct {
	wrappedErr error
	silenceUsage bool
	prefix    string
	msgFormat string
}

type cliDefinedErrorImpl struct {
	msg        string
	directions string
	silenceUsage bool
}

func (b *cliDefinedErrorImpl) Error() string {
	return b.msg
}

func (b *cliDefinedErrorImpl) SilenceUsage() bool {
	return b.silenceUsage
}

func (b *cliDefinedErrorImpl) SetDirectionsMsg(format string, args ...interface{}) {
	b.directions = fmt.Sprintf(format, args...)
}

func (b *cliDefinedErrorImpl) GetDirectionsMsg() string {
	return b.directions
}
