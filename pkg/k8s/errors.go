package k8s

import (
	"fmt"
	"strings"
)

// AlreadyExistsError resource already exists on kubernetes
type AlreadyExistsError struct {
	K8sError
}

// NotFoundError requested kubernetes resource not found
type NotFoundError struct {
	K8sError
}

// K8sError generic kubectl error, which also provides the error interface for AlreadyExistsError and NotFoundError
type K8sError struct {
	message string
	err     error
}

func (e K8sError) Error() string {
	return fmt.Sprintf("k8s: %s, %s", e.err.Error(), e.message)
}

func newK8sError(out []byte, err error) error {
	message := string(out)
	if strings.Contains(message, "Error from server (AlreadyExists):") {
		return AlreadyExistsError{K8sError{message, err}}
	}

	if strings.Contains(message, "Error from server (NotFound):") {
		return NotFoundError{K8sError{message, err}}
	}

	return K8sError{message, err}
}
