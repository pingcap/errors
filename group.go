package errors

// ErrorGroup is an interface for multiple errors that are not a chain.
// This happens for example when executing multiple operations in parallel.
type ErrorGroup interface {
	Errors() []error
}

// WalkDeep does a depth-first traversal of all errors.
// Any ErrorGroup is traversed (after going deep).
// The visitor function can return false to end the traversal early
// In that case, WalkDeep will return false, otherwise true
func WalkDeep(err error, visitor func(err error) bool) bool {
	// Go deep
	unErr := err
	for unErr != nil {
		if more := visitor(unErr); more == true {
			return true
		}
		unErr = Unwrap(unErr)
	}

	// Go wide
	if hasGroup, ok := err.(ErrorGroup); ok {
		for _, err := range hasGroup.Errors() {
			if more := WalkDeep(err, visitor); more == true {
				return true
			}
		}
	}

	return true
}
