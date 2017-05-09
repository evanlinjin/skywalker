package skywalker

import "errors"

var (
	// ErrRootNotFound happens when root is not found with public key.
	ErrRootNotFound = errors.New("root not found")

	// ErrObjNotFound happens when specified object is not found.
	ErrObjNotFound = errors.New("object not found")

	// ErrFieldNotFound happens when an object's field by name is not found.
	ErrFieldNotFound = errors.New("field not found")

	// ErrFieldNotProvided occurs when the field name of a struct is not provided.
	ErrFieldNotProvided = errors.New("field not provided")
)
