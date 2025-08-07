package logr

import (
	"bytes"
	"testing"

	"github.com/go-logr/logr/funcr"

	"darvaza.org/core"
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
	core.AssertEqual(t, originalLen, len(fields), "input map length")

	// Check if empty key is still there
	_, hasEmpty := fields[""]
	core.AssertTrue(t, hasEmpty, "empty key present")
}
