package errors

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// ==================== juju adaptor start ========================

// Trace annotates err with a stack trace at the point WithStack was called.
// If err is nil or already contain stack trace return directly.
func Trace(err error) error {
	if err == nil {
		return nil
	}
	if errHasStack(err) {
		return err
	}
	return &withStack{
		err,
		callers(),
	}
}

// error passed as the parameter is not an annotated error, the result is		 // error passed as the parameter is not an annotated error, the result is
// simply the result of the Error() method on that error.		 // simply the result of the Error() method on that error.
func ErrorStack(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%+v", err)
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
