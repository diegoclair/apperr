package apperr

// AppError is the common interface between Definition (static error) and Error (error with context).
// Both can be returned as error. Helpers like IsNotFound(), AsAppError(), and the httpmap package
// check for this interface.
type AppError interface {
	error
	Kind() Kind
	Code() Code
	Message() string
}
