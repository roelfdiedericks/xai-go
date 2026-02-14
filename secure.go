// Package xai provides a Go client for the xAI gRPC API.
package xai

import (
	"sync"
	"unsafe"
)

// SecureString holds a sensitive string value and clears it from memory when closed.
// This provides defense-in-depth for API keys and other secrets.
type SecureString struct {
	mu    sync.RWMutex
	value []byte
}

// NewSecureString creates a new SecureString from the given value.
func NewSecureString(value string) *SecureString {
	s := &SecureString{
		value: []byte(value),
	}
	return s
}

// Value returns the string value. Returns empty string if closed.
func (s *SecureString) Value() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.value == nil {
		return ""
	}
	return string(s.value)
}

// Close zeroes out the memory and marks the string as closed.
// Safe to call multiple times.
func (s *SecureString) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.value != nil {
		// Zero out the memory
		for i := range s.value {
			s.value[i] = 0
		}
		s.value = nil
	}
}

// Len returns the length of the value, or 0 if closed.
func (s *SecureString) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.value)
}

// IsZero returns true if the SecureString is nil or closed.
func (s *SecureString) IsZero() bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.value) == 0
}

// Redacted returns a redacted version of the key for logging.
// Shows first 4 and last 4 characters with asterisks in between.
func (s *SecureString) Redacted() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.value) < 12 {
		return "****"
	}
	return string(s.value[:4]) + "****" + string(s.value[len(s.value)-4:])
}

// compile-time check that we're not accidentally copying
var _ = unsafe.Sizeof(SecureString{})
