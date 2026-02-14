package xai_test

import (
	"testing"

	xai "github.com/roelfdiedericks/xai-go"
)

func TestSecureString(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		s := xai.NewSecureString("secret123")
		if got := s.Value(); got != "secret123" {
			t.Errorf("Value() = %q, want %q", got, "secret123")
		}
	})

	t.Run("Len", func(t *testing.T) {
		s := xai.NewSecureString("secret")
		if got := s.Len(); got != 6 {
			t.Errorf("Len() = %d, want %d", got, 6)
		}
	})

	t.Run("IsZero", func(t *testing.T) {
		var nilSecure *xai.SecureString
		if !nilSecure.IsZero() {
			t.Error("nil SecureString should be zero")
		}

		empty := xai.NewSecureString("")
		if !empty.IsZero() {
			t.Error("empty SecureString should be zero")
		}

		nonEmpty := xai.NewSecureString("value")
		if nonEmpty.IsZero() {
			t.Error("non-empty SecureString should not be zero")
		}
	})

	t.Run("Redacted", func(t *testing.T) {
		short := xai.NewSecureString("short")
		if got := short.Redacted(); got != "****" {
			t.Errorf("Redacted() for short string = %q, want %q", got, "****")
		}

		long := xai.NewSecureString("sk_test_1234567890abcdef")
		got := long.Redacted()
		if got != "sk_t****cdef" {
			t.Errorf("Redacted() = %q, want %q", got, "sk_t****cdef")
		}
	})

	t.Run("Close", func(t *testing.T) {
		s := xai.NewSecureString("secret")
		s.Close()

		if got := s.Value(); got != "" {
			t.Errorf("Value() after Close = %q, want empty string", got)
		}

		if !s.IsZero() {
			t.Error("IsZero() after Close should be true")
		}

		// Should be safe to call Close multiple times
		s.Close()
	})
}
