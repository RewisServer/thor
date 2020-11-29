package slog

import "testing"

func TestSetVerbosity(t *testing.T) {
	SetVerbosity(0)
	if V(1) {
		t.Errorf("Verbosity check is incorrect, expected 'false' but got '%v', level: %v", V(1), 0)
	}

	SetVerbosity(1)
	if !V(1) {
		t.Errorf("Verbosity check is incorrect, expected 'true' but got '%v', level: %v", V(1), 1)
	}
}
