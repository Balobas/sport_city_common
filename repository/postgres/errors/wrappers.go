package pgErrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("database error: (%s) %w", msg, err)
}

func WrapWithHandleExistsFlag[T any](obj T, err error, msg string) (T, bool, error) {
	if err == nil {
		return obj, true, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return obj, false, nil
	}
	return obj, false, Wrap(err, msg)
}
