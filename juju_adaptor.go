package errors

import (
	log "github.com/sirupsen/logrus"
)

// ==================== juju adaptor start ========================

// Trace annotates err with a stack trace at the point WithStack was called.
// If err is nil or already contain stack trace return directly.
func Trace(err error) error {
	if err == nil {
		return nil
	}
	errWithStack, hasStack := err.(withStackAware)
	if hasStack {
		hasStack = errWithStack.hasStack()
	}
	if hasStack {
		return err
	}
	return &withStack{
		err,
		callers(),
	}
}

// Wrap changes the Cause of the error, old error stack also be output.
func Wrap(oldErr, newErr error) error {
	log.Errorf("%+v", oldErr)
	return Trace(newErr)
}

// NotFoundf represents an error with not found message.
func NotFoundf(format string, args ...interface{}) error {
	return Errorf(format+" not found", args...)
}

// BadRequestf represents an error with bad request message.
func BadRequestf(format string, args ...interface{}) error {
	return Errorf(format+" bad request", args...)
}

// NotSupportedf represents an error with not supported message.
func NotSupportedf(format string, args ...interface{}) error {
	return Errorf(format+" not supported", args...)
}

// ==================== juju adaptor end ========================
