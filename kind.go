package apperr

// Kind represents the category of an application error.
// It is transport-agnostic — the transport layer maps it to HTTP/gRPC/etc.
type Kind uint8

const (
	KindValidation     Kind = iota + 1 // input or business rule violation
	KindNotFound                        // resource not found
	KindConflict                        // duplicate, constraint violation
	KindAuthentication                  // invalid credentials, expired token
	KindAuthorization                   // no permission for the resource
	KindInternal                        // unexpected system error
	KindRateLimited                     // rate limit exceeded
)

// String returns the name of the Kind (useful for logs and debug).
func (k Kind) String() string {
	switch k {
	case KindValidation:
		return "validation"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindAuthentication:
		return "authentication"
	case KindAuthorization:
		return "authorization"
	case KindInternal:
		return "internal"
	case KindRateLimited:
		return "rate_limited"
	default:
		return "unknown"
	}
}
