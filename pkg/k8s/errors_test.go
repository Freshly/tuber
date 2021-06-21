package k8s

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewK8sError(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		err         error
		expectedErr error
	}{
		{
			name:        "already exists",
			input:       "Error from server (AlreadyExists): blah",
			err:         errors.New("it exists"),
			expectedErr: AlreadyExistsError{K8sError{"Error from server (AlreadyExists): blah", errors.New("it exists")}},
		},
		{
			name:        "not found",
			input:       "Error from server (NotFound): where is it",
			err:         errors.New("it's nowhere"),
			expectedErr: NotFoundError{K8sError{"Error from server (NotFound): where is it", errors.New("it's nowhere")}},
		},
		{
			name:        "only error provided",
			input:       "",
			err:         errors.New("exec: \"kubectl\": executable file not found in $PATH"),
			expectedErr: K8sError{"", errors.New("exec: \"kubectl\": executable file not found in $PATH")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := newK8sError([]byte(tc.input), tc.err)

			assert.Equal(t, tc.expectedErr, actual)
		})
	}
}
