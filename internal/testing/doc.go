// Package testing provides testing utilities for slog implementations.
//
// This package follows the darvaza.org/core testing patterns with two types of assertion functions:
//
//   - Assert functions (AssertMessage, AssertField, etc.) return boolean success indicators
//     and allow tests to continue execution after failures. Use these when you want to
//     collect multiple assertion failures or perform custom error handling.
//
//   - AssertMust functions (AssertMustMessage, AssertMustField, etc.) call the corresponding
//     Assert function and terminate test execution immediately on failure using t.FailNow().
//     Use these when the assertion failure makes continuing the test meaningless.
//
// All assertion functions use core.T interface and are compatible with both *testing.T
// and core.MockT for testing test code itself.
//
// # Assert vs AssertMust Pattern
//
// Standard assertions return boolean results:
//
//	if !AssertMessage(t, msg, slog.Info, "expected text") {
//		// Handle failure - test continues
//		return
//	}
//
// Fatal assertions terminate immediately on failure:
//
//	AssertMustMessage(t, msg, slog.Info, "expected text")
//	// If assertion fails, test stops here with t.FailNow()
//
// # Message Testing Utilities
//
// The package provides specialized assertion functions for testing slog messages:
//
//   - AssertMessage/AssertMustMessage: Verify log level and message text
//   - AssertField/AssertMustField: Verify field existence and value
//   - AssertNoField/AssertMustNoField: Verify field absence
//   - AssertMessageCount/AssertMustMessageCount: Verify message count with debugging output
//
// # Test Execution Helpers
//
// Helper functions for running tests with logger instances:
//
//   - RunWithLogger: Execute test function with specific logger
//   - RunWithLoggerFactory: Execute test with fresh logger instance
//   - TestLevelMethods: Standard tests for all log level methods
//   - TestFieldMethods: Standard tests for field handling methods
//
// # Message Comparison and Transformation
//
// Utilities for working with message collections:
//
//   - TransformMessages: Apply level transformations based on AdapterOptions
//   - CompareMessages: Set-based comparison of message arrays
//
// These utilities support comprehensive testing of slog implementations while maintaining
// consistency with the darvaza.org/core testing patterns and linting requirements.
package testing
