package skywalker

import "errors"

var (
	// ErrRootNotFound happens when root is not found with public key.
	ErrRootNotFound = errors.New("root not found")

	// ErrObjNotFound happens when specified object is not found.
	ErrObjNotFound = errors.New("object not found")
)
