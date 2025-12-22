package commonErrors

import (
	"github.com/pkg/errors"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type BaseError struct {
	code ErrorCode
	msg  string
	err  error
}

func NewBaseError(code ErrorCode, msg string, err error, opts ...ErrorOption) *BaseError {
	baseErr := &BaseError{
		code: code,
		msg:  msg,
		err:  err,
	}
	baseErr.handleOptions(opts...)
	return baseErr
}

func (be *BaseError) handleOptions(opts ...ErrorOption) {
	options := &errorOptions{}

	for _, apply := range opts {
		apply(options)
	}

	if options.withSpanRecord {
		options.spanForRecord.RecordError(be.err)
	}

	if options.withDebugLog {
		options.loggerForDebug.Debug().Str("error", be.Error()).Send()
	}
}

func (be *BaseError) Error() string {
	return be.err.Error()
}

func (be *BaseError) Code() ErrorCode {
	return be.code
}

func (be *BaseError) Unwrap() error {
	return be.err
}
