package xai_test

import (
	"errors"
	"testing"

	xai "github.com/roelfdiedericks/xai-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorCode(t *testing.T) {
	tests := []struct {
		code xai.ErrorCode
		want string
	}{
		{xai.ErrAuth, "authentication_error"},
		{xai.ErrRateLimit, "rate_limit_error"},
		{xai.ErrInvalidRequest, "invalid_request_error"},
		{xai.ErrNotFound, "not_found_error"},
		{xai.ErrServerError, "server_error"},
		{xai.ErrUnavailable, "unavailable_error"},
		{xai.ErrTimeout, "timeout_error"},
		{xai.ErrCanceled, "canceled_error"},
		{xai.ErrResourceExhausted, "resource_exhausted_error"},
		{xai.ErrUnknown, "unknown_error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.code.String(); got != tt.want {
				t.Errorf("ErrorCode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	t.Run("Error message", func(t *testing.T) {
		err := &xai.Error{
			Code:    xai.ErrAuth,
			Message: "invalid api key",
		}
		if got := err.Error(); got != "authentication_error: invalid api key" {
			t.Errorf("Error() = %q, want %q", got, "authentication_error: invalid api key")
		}
	})

	t.Run("Error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &xai.Error{
			Code:    xai.ErrServerError,
			Message: "server failed",
			Cause:   cause,
		}
		if got := err.Error(); got != "server_error: server failed: underlying error" {
			t.Errorf("Error() = %q, want server_error: server failed: underlying error", got)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := errors.New("underlying")
		err := &xai.Error{
			Code:    xai.ErrServerError,
			Message: "test",
			Cause:   cause,
		}
		if got := err.Unwrap(); got != cause {
			t.Errorf("Unwrap() = %v, want %v", got, cause)
		}
	})

	t.Run("IsRetryable", func(t *testing.T) {
		retryable := []xai.ErrorCode{
			xai.ErrRateLimit,
			xai.ErrUnavailable,
			xai.ErrTimeout,
			xai.ErrServerError,
		}
		for _, code := range retryable {
			err := &xai.Error{Code: code}
			if !err.IsRetryable() {
				t.Errorf("ErrorCode %v should be retryable", code)
			}
		}

		notRetryable := []xai.ErrorCode{
			xai.ErrAuth,
			xai.ErrInvalidRequest,
			xai.ErrNotFound,
		}
		for _, code := range notRetryable {
			err := &xai.Error{Code: code}
			if err.IsRetryable() {
				t.Errorf("ErrorCode %v should not be retryable", code)
			}
		}
	})
}

func TestFromGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		grpcCode codes.Code
		wantCode xai.ErrorCode
	}{
		{"Unauthenticated", codes.Unauthenticated, xai.ErrAuth},
		{"PermissionDenied", codes.PermissionDenied, xai.ErrAuth},
		{"ResourceExhausted", codes.ResourceExhausted, xai.ErrRateLimit},
		{"InvalidArgument", codes.InvalidArgument, xai.ErrInvalidRequest},
		{"NotFound", codes.NotFound, xai.ErrNotFound},
		{"Internal", codes.Internal, xai.ErrServerError},
		{"Unavailable", codes.Unavailable, xai.ErrUnavailable},
		{"DeadlineExceeded", codes.DeadlineExceeded, xai.ErrTimeout},
		{"Canceled", codes.Canceled, xai.ErrCanceled},
		{"Unknown", codes.Unknown, xai.ErrUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := status.Error(tt.grpcCode, "test message")
			err := xai.FromGRPCError(grpcErr)
			if err.Code != tt.wantCode {
				t.Errorf("FromGRPCError() code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.GRPCCode != tt.grpcCode {
				t.Errorf("FromGRPCError() GRPCCode = %v, want %v", err.GRPCCode, tt.grpcCode)
			}
		})
	}

	t.Run("nil error", func(t *testing.T) {
		if got := xai.FromGRPCError(nil); got != nil {
			t.Errorf("FromGRPCError(nil) = %v, want nil", got)
		}
	})

	t.Run("non-grpc error", func(t *testing.T) {
		plainErr := errors.New("plain error")
		got := xai.FromGRPCError(plainErr)
		if got.Code != xai.ErrUnknown {
			t.Errorf("FromGRPCError() code = %v, want ErrUnknown", got.Code)
		}
	})
}

func TestErrorIs(t *testing.T) {
	authErr := &xai.Error{Code: xai.ErrAuth, Message: "test"}

	if !errors.Is(authErr, xai.ErrAuthSentinel) {
		t.Error("Auth error should match ErrAuthSentinel")
	}

	if errors.Is(authErr, xai.ErrRateLimitSentinel) {
		t.Error("Auth error should not match ErrRateLimitSentinel")
	}
}
