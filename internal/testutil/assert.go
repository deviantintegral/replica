package testutil

import (
	"strings"
	"testing"
)

// AssertEqual checks that got equals want and fails the test if they differ.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// AssertContains fails the test if s does not contain substr.
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
