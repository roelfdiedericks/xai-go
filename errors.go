package xai

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorCode represents the category of an error.
type ErrorCode int

const (
	// ErrUnknown indicates an unknown error.
	ErrUnknown ErrorCode = iota
	// ErrAuth indicates an authentication error (invalid/expired API key).
	ErrAuth
	// ErrRateLimit indicates the request was rate limited.
	ErrRateLimit
	// ErrInvalidRequest indicates the request was malformed or invalid.
	ErrInvalidRequest
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound
	// ErrServerError indicates an internal server error.
	ErrServerError
	// ErrUnavailable indicates the service is temporarily unavailable.
	ErrUnavailable
	// ErrTimeout indicates the request timed out.
	ErrTimeout
	// ErrCanceled indicates the request was canceled.
	ErrCanceled
	// ErrResourceExhausted indicates quota or resource limits exceeded.
	ErrResourceExhausted
)

// String returns a human-readable name for the error code.
func (c ErrorCode) String() string {
	switch c {
	case ErrAuth:
		return "authentication_error"
	case ErrRateLimit:
		return "rate_limit_error"
	case ErrInvalidRequest:
		return "invalid_request_error"
	case ErrNotFound:
		return "not_found_error"
	case ErrServerError:
		return "server_error"
	case ErrUnavailable:
		return "unavailable_error"
	case ErrTimeout:
		return "timeout_error"
	case ErrCanceled:
		return "canceled_error"
	case ErrResourceExhausted:
		return "resource_exhausted_error"
	default:
		return "unknown_error"
	}
}

// Error represents an xAI API error with structured information.
type Error struct {
	// Code is the error category.
	Code ErrorCode
	// Message is a human-readable error message.
	Message string
	// Cause is the underlying error, if any.
	Cause error
	// RetryAfter is set for rate limit errors, indicating when to retry.
	RetryAfter time.Duration
	// GRPCCode is the original gRPC status code.
	GRPCCode codes.Code
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *Error) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error is transient and the request can be retried.
func (e *Error) IsRetryable() bool {
	switch e.Code {
	case ErrRateLimit, ErrUnavailable, ErrTimeout, ErrServerError:
		return true
	default:
		return false
	}
}

// IsAuth returns true if this is an authentication error.
func (e *Error) IsAuth() bool {
	return e.Code == ErrAuth
}

// IsRateLimit returns true if this is a rate limit error.
func (e *Error) IsRateLimit() bool {
	return e.Code == ErrRateLimit
}

// Sentinel errors for errors.Is checks.
var (
	ErrAuthSentinel         = &Error{Code: ErrAuth}
	ErrRateLimitSentinel    = &Error{Code: ErrRateLimit}
	ErrInvalidSentinel      = &Error{Code: ErrInvalidRequest}
	ErrNotFoundSentinel     = &Error{Code: ErrNotFound}
	ErrServerSentinel       = &Error{Code: ErrServerError}
	ErrUnavailableSentinel  = &Error{Code: ErrUnavailable}
	ErrTimeoutSentinel      = &Error{Code: ErrTimeout}
	ErrCanceledSentinel     = &Error{Code: ErrCanceled}
	ErrExhaustedSentinel    = &Error{Code: ErrResourceExhausted}
)

// Is implements errors.Is for Error matching by code.
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// FromGRPCError converts a gRPC error to an xAI Error.
func FromGRPCError(err error) *Error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return &Error{
			Code:    ErrUnknown,
			Message: err.Error(),
			Cause:   err,
		}
	}

	xaiErr := &Error{
		Message:  st.Message(),
		Cause:    err,
		GRPCCode: st.Code(),
	}

	switch st.Code() {
	case codes.Unauthenticated:
		xaiErr.Code = ErrAuth
		xaiErr.Message = "authentication failed: " + st.Message()
	case codes.PermissionDenied:
		xaiErr.Code = ErrAuth
		xaiErr.Message = "permission denied: " + st.Message()
	case codes.ResourceExhausted:
		// Could be rate limit or quota
		xaiErr.Code = ErrRateLimit
		xaiErr.Message = "rate limit exceeded: " + st.Message()
		// TODO: Parse retry-after from metadata if available
	case codes.InvalidArgument:
		xaiErr.Code = ErrInvalidRequest
	case codes.NotFound:
		xaiErr.Code = ErrNotFound
	case codes.Internal:
		xaiErr.Code = ErrServerError
	case codes.Unavailable:
		xaiErr.Code = ErrUnavailable
	case codes.DeadlineExceeded:
		xaiErr.Code = ErrTimeout
	case codes.Canceled:
		xaiErr.Code = ErrCanceled
	case codes.FailedPrecondition:
		xaiErr.Code = ErrInvalidRequest
	case codes.Aborted:
		xaiErr.Code = ErrUnavailable
	case codes.OutOfRange:
		xaiErr.Code = ErrInvalidRequest
	case codes.Unimplemented:
		xaiErr.Code = ErrInvalidRequest
		xaiErr.Message = "method not implemented: " + st.Message()
	case codes.DataLoss:
		xaiErr.Code = ErrServerError
	default:
		xaiErr.Code = ErrUnknown
	}

	return xaiErr
}

// WrapError wraps an error with additional context.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	var xaiErr *Error
	if errors.As(err, &xaiErr) {
		return &Error{
			Code:       xaiErr.Code,
			Message:    message + ": " + xaiErr.Message,
			Cause:      xaiErr.Cause,
			RetryAfter: xaiErr.RetryAfter,
			GRPCCode:   xaiErr.GRPCCode,
		}
	}
	return fmt.Errorf("%s: %w", message, err)
}
