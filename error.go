package e

import "google.golang.org/grpc/codes"

type LogicError struct {
	onlyRead bool
	code     int
	message  string
	details  []string
	wrapErr  []error
	data     any
}

func (e *LogicError) Error() string {
	if len(e.details) > 0 {
		return e.message
	}
	return e.message
}

func IsAppError(err error) (bool, *LogicError) {
	errC, ok := err.(*LogicError)
	if ok {
		return true, errC
	}

	return false, nil
}

func NewError(code int, message string, details []string) *LogicError {
	return &LogicError{
		code:    code,
		message: message,
		details: details,
	}
}

func NewErrorFrom(parentErr *LogicError) *LogicError {
	e := &LogicError{
		code:    parentErr.code,
		message: parentErr.message,
		details: parentErr.details,
	}
	return e.Wrap(parentErr)
}

func (e *LogicError) Wrap(err error) *LogicError {
	if e.onlyRead {
		return e
	}

	e.wrapErr = append(e.wrapErr, err)
	return e
}

func (e *LogicError) Unwrap() []error {
	return e.wrapErr
}

func (e *LogicError) Is(err error) bool {
	if e == err {
		return true
	}
	for _, wErr := range e.wrapErr {
		if wErr == err {
			return true
		}
	}
	return false
}

func (e *LogicError) Lock() *LogicError {
	e.onlyRead = true
	return e
}

func (e *LogicError) Code() int {
	return e.code
}

func (e *LogicError) Message() string {
	return e.message
}

func (e *LogicError) Details() []string {
	return e.details
}

func (e *LogicError) SetMessage(message string) *LogicError {
	if e.onlyRead {
		return e
	}
	e.message = message
	return e
}

func (e *LogicError) AddDetails(details []string) *LogicError {
	if e.onlyRead {
		return e
	}
	if e.details == nil {
		e.details = make([]string, 0, len(details))
	}
	e.details = append(e.details, details...)
	return e
}

func (e *LogicError) Data() any {
	return e.data
}

func (e *LogicError) SetData(data any) *LogicError {
	if e.onlyRead {
		return e
	}

	e.data = data
	return e
}

func (e *LogicError) GetGRPCCode() codes.Code {
	switch e.code {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 406:
		return codes.FailedPrecondition
	case 409:
		return codes.Aborted
	case 422:
		return codes.InvalidArgument
	case 500:
		return codes.Internal
	case 503:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}
