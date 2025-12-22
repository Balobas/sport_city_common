package commonErrors

type ErrorCode int

const (
	ErrorCodeAlreadyExists = iota + 1
	ErrorCodeNotFound
	ErrorCodeInternal
	ErrorCodeBadRequest
)
