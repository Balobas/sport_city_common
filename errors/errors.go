package commonErrors

import (
	"fmt"

	"github.com/pkg/errors"
)

type InternalError struct {
	*BaseError
}

func NewInternalError(err error, opts ...ErrorOption) *InternalError {
	return &InternalError{
		BaseError: NewBaseError(
			ErrorCodeInternal,
			"",
			err,
			opts...,
		),
	}
}

func (ie *InternalError) Unwrap() error {
	return ie.BaseError
}

type AlreadyExistsError struct {
	*BaseError
}

func NewAlreadyExistsError(msg string, opts ...ErrorOption) *AlreadyExistsError {
	return &AlreadyExistsError{
		BaseError: NewBaseError(
			ErrorCodeAlreadyExists,
			msg,
			errors.New(msg),
			opts...,
		),
	}
}

func (aer *AlreadyExistsError) Unwrap() error {
	return aer.BaseError
}

type NotFoundError struct {
	*BaseError
}

func NewNotFoundError(msg string, opts ...ErrorOption) *NotFoundError {
	return &NotFoundError{
		BaseError: NewBaseError(
			ErrorCodeNotFound,
			msg,
			errors.New(msg),
			opts...,
		),
	}
}

func (nfe *NotFoundError) Unwrap() error {
	return nfe.BaseError
}

type DatabaseError struct {
	*BaseError
}

func NewDatabaseError(err error, opts ...ErrorOption) *DatabaseError {
	return &DatabaseError{
		BaseError: NewBaseError(
			ErrorCodeInternal,
			"database error",
			fmt.Errorf("database error: %w", err),
			opts...,
		),
	}
}

func (dbe *DatabaseError) Unwrap() error {
	return dbe.BaseError
}

func IsDatabaseError(err error) bool {
	var target *DatabaseError
	return errors.As(err, target)
}
