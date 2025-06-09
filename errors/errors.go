package commonErrors

import "github.com/pkg/errors"

var (
	ErrAlreadyExists = errors.New("already exists")
)
