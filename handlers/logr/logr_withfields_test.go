package logr

import (
	"bytes"
	"testing"

	"github.com/go-logr/logr/funcr"
)

// TestWithFieldsDoesNotModifyInput tests that WithFields doesn't modify the input map
func TestWithFieldsDoesNotModifyInput(t *testing.T) {
	var buf bytes.Buffer
	logrLogger := funcr.New(func(prefix, args string) {
		_, _ = buf.WriteString(prefix)
		if args != "" {
			_, _ = buf.WriteString(" ")
			_, _ = buf.WriteString(args)
		}
		_, _ = buf.WriteString("\n")
	}, funcr.Options{
		Verbosity: 2,
	})

	logger := New(logrLogger)

	// Create a map with an empty key
	fields := map[string]any{
		"":      "empty",
		"valid": "value",
	}

	// Store original length
	originalLen := len(fields)

	// Call WithFields
	logger.Info().WithFields(fields).Print("test")

	// Check if the map was modified
	if len(fields) != originalLen {
		t.Errorf("WithFields modified the input map: original length %d, new length %d", originalLen, len(fields))
	}

	// Check if empty key is still there
	if _, hasEmpty := fields[""]; !hasEmpty {
		t.Error("WithFields removed the empty key from the input map")
	}
}
