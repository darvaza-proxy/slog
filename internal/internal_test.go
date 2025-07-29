package internal_test

import (
	"testing"

	"darvaza.org/slog/internal"
)

func TestHasFields(t *testing.T) {
	t.Run("EmptyMap", testHasFieldsEmpty)
	t.Run("NilMap", testHasFieldsNil)
	t.Run("ValidFields", testHasFieldsValid)
	t.Run("EmptyKeys", testHasFieldsEmptyKeys)
	t.Run("MixedKeys", testHasFieldsMixed)
}

func testHasFieldsEmpty(t *testing.T) {
	fields := map[string]any{}

	if internal.HasFields(fields) {
		t.Error("HasFields should return false for empty map")
	}
}

func testHasFieldsNil(t *testing.T) {
	var fields map[string]any

	if internal.HasFields(fields) {
		t.Error("HasFields should return false for nil map")
	}
}

func testHasFieldsValid(t *testing.T) {
	testCases := []struct {
		name   string
		fields map[string]any
		want   bool
	}{
		{
			"SingleField",
			map[string]any{"key": "value"},
			true,
		},
		{
			"MultipleFields",
			map[string]any{"key1": "value1", "key2": "value2"},
			true,
		},
		{
			"NilValue",
			map[string]any{"key": nil},
			true,
		},
		{
			"ZeroValue",
			map[string]any{"key": 0},
			true,
		},
		{
			"EmptyStringValue",
			map[string]any{"key": ""},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.HasFields(tc.fields)
			if got != tc.want {
				t.Errorf("HasFields() = %v, want %v", got, tc.want)
			}
		})
	}
}

func testHasFieldsEmptyKeys(t *testing.T) {
	testCases := []struct {
		name   string
		fields map[string]any
		want   bool
	}{
		{
			"OnlyEmptyKey",
			map[string]any{"": "value"},
			false,
		},
		{
			"OnlyEmptyKeySingle",
			map[string]any{"": "value1"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.HasFields(tc.fields)
			if got != tc.want {
				t.Errorf("HasFields() = %v, want %v for %v", got, tc.want, tc.fields)
			}
		})
	}
}

func testHasFieldsMixed(t *testing.T) {
	testCases := []struct {
		name   string
		fields map[string]any
		want   bool
	}{
		{
			"ValidAndEmptyKeys",
			map[string]any{"": "empty", "valid": "value"},
			true, // Should return true because "valid" is non-empty
		},
		{
			"OnlyEmptyKeysMultiple",
			map[string]any{"": "value"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.HasFields(tc.fields)
			if got != tc.want {
				t.Errorf("HasFields() = %v, want %v for %v", got, tc.want, tc.fields)
			}
		})
	}
}

func TestHasFieldsEdgeCases(t *testing.T) {
	t.Run("WhitespaceKeys", func(t *testing.T) {
		fields := map[string]any{
			" ":  "space",
			"\t": "tab",
			"\n": "newline",
			"  ": "spaces",
		}

		// Whitespace keys are not empty strings, so should return true
		if !internal.HasFields(fields) {
			t.Error("HasFields should return true for whitespace keys")
		}
	})

	t.Run("LargeMap", func(t *testing.T) {
		fields := make(map[string]any, 1000)

		// Fill with empty keys
		for i := 0; i < 999; i++ {
			fields[""] = i // All will overwrite to same empty key
		}

		// Add one valid key
		fields["valid"] = "value"

		if !internal.HasFields(fields) {
			t.Error("HasFields should return true when at least one key is non-empty")
		}
	})

	t.Run("UnicodeKeys", func(t *testing.T) {
		fields := map[string]any{
			"测试":   "chinese",
			"🔑":    "emoji",
			"café": "accented",
		}

		if !internal.HasFields(fields) {
			t.Error("HasFields should return true for unicode keys")
		}
	})
}
