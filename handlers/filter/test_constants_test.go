package filter_test

// Common test constants used across multiple test files
const (
	// redactedValue is the replacement value for sensitive fields
	redactedValue = "[REDACTED]"

	// sensitiveKey1 is a sensitive field name that should be redacted
	sensitiveKey1 = "password"

	// sensitiveKey2 is another sensitive field name that should be redacted
	sensitiveKey2 = "secret"
)
