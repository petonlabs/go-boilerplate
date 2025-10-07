package testhelpers

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertTimestampsValid checks that created_at and updated_at fields are set
func AssertTimestampsValid(t *testing.T, obj interface{}) {
	t.Helper()

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	createdField := val.FieldByName("CreatedAt")
	if createdField.IsValid() {
		// Avoid calling Interface() on unexported fields which will panic.
		if !createdField.CanInterface() {
			t.Log("CreatedAt is unexported; skipping assertion")
		} else {
			switch v := createdField.Interface().(type) {
			case time.Time:
				assert.False(t, v.IsZero(), "CreatedAt should not be zero")
			case *time.Time:
				require.NotNil(t, v, "CreatedAt should not be nil")
				assert.False(t, v.IsZero(), "CreatedAt should not be zero")
			default:
				require.Failf(t, "CreatedAt has unexpected type", "expected time.Time or *time.Time, got %T", v)
			}
		}
	}

	updatedField := val.FieldByName("UpdatedAt")
	if updatedField.IsValid() {
		// Avoid calling Interface() on unexported fields which will panic.
		if !updatedField.CanInterface() {
			t.Log("UpdatedAt is unexported; skipping assertion")
		} else {
			switch v := updatedField.Interface().(type) {
			case time.Time:
				assert.False(t, v.IsZero(), "UpdatedAt should not be zero")
			case *time.Time:
				require.NotNil(t, v, "UpdatedAt should not be nil")
				assert.False(t, v.IsZero(), "UpdatedAt should not be zero")
			default:
				require.Failf(t, "UpdatedAt has unexpected type", "expected time.Time or *time.Time, got %T", v)
			}
		}
	}
}

// AssertValidUUID checks that the UUID is valid and not nil
func AssertValidUUID(t *testing.T, id uuid.UUID, message ...string) {
	t.Helper()

	msg := "UUID should not be nil"
	if len(message) > 0 {
		msg = message[0]
	}

	assert.NotEqual(t, uuid.Nil, id, msg)
}

// AssertEqualExceptTime asserts that two objects are equal, ignoring time fields
func AssertEqualExceptTime(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedVal := reflect.ValueOf(expected)
	if expectedVal.Kind() == reflect.Ptr {
		expectedVal = expectedVal.Elem()
	}

	actualVal := reflect.ValueOf(actual)
	if actualVal.Kind() == reflect.Ptr {
		actualVal = actualVal.Elem()
	}

	require.Equal(t, expectedVal.Type(), actualVal.Type(), "objects are not the same type")

	for i := 0; i < expectedVal.NumField(); i++ {
		field := expectedVal.Type().Field(i)

		// Skip time fields at the top-level
		if field.Type == reflect.TypeOf(time.Time{}) ||
			field.Type == reflect.TypeOf(&time.Time{}) {
			continue
		}

		expectedField := expectedVal.Field(i)
		actualField := actualVal.Field(i)

		// If either field is unexported, handle exportability mismatch or skip.
		if !expectedField.CanInterface() || !actualField.CanInterface() {
			if expectedField.CanInterface() != actualField.CanInterface() {
				require.Failf(t, "field mismatch", "field %s: exportability differs between expected and actual", field.Name)
			} else {
				// Both unexported; cannot compare via Interface(). Skip with a log.
				t.Logf("field %s is unexported; skipping comparison", field.Name)
			}
			continue
		}

		// Use recursive comparison for structs / pointer-to-structs to skip nested time fields
		if (expectedField.Kind() == reflect.Struct) || (expectedField.Kind() == reflect.Ptr && expectedField.Type().Elem().Kind() == reflect.Struct) {
			compareExceptTimeValue(t, expectedField, actualField, field.Name)
			continue
		}

		// Fallback: concrete value comparison
		assert.Equal(
			t,
			expectedField.Interface(),
			actualField.Interface(),
			fmt.Sprintf("field %s should be equal", field.Name),
		)
	}
}

// compareExceptTimeValue recursively compares two reflect.Values while skipping any
// fields that are time.Time or *time.Time. It avoids calling Interface() on
// unexported fields and handles pointer indirection.
func compareExceptTimeValue(t *testing.T, expected, actual reflect.Value, path string) {
	// Handle pointer indirection
	if expected.Kind() == reflect.Ptr {
		if actual.Kind() != reflect.Ptr {
			require.Failf(t, "kind mismatch", "%s: expected pointer, actual not pointer", path)
			return
		}
		if expected.IsNil() || actual.IsNil() {
			// Both nil -> equal, one nil -> fail
			if expected.IsNil() != actual.IsNil() {
				assert.Equal(t, expected.IsNil(), actual.IsNil(), fmt.Sprintf("field %s should be equal (nil mismatch)", path))
			}
			return
		}
		// Both non-nil: dereference
		expected = expected.Elem()
		actual = actual.Elem()
	}

	if expected.Kind() != reflect.Struct {
		// For non-structs, just compare the concrete values (should be safe to Interface because callers check CanInterface)
		switch {
		case expected.CanInterface() && actual.CanInterface():
			assert.Equal(t, expected.Interface(), actual.Interface(), fmt.Sprintf("field %s should be equal", path))
		case expected.CanInterface() != actual.CanInterface():
			require.Failf(t, "field mismatch", "field %s: exportability differs while recursing", path)
		default:
			t.Logf("field %s unexported while recursing; skipping", path)
		}
		return
	}

	// expected is a struct: iterate exported fields and recurse
	for i := 0; i < expected.NumField(); i++ {
		sf := expected.Type().Field(i)
		// Skip nested time fields
		if sf.Type == reflect.TypeOf(time.Time{}) || sf.Type == reflect.TypeOf(&time.Time{}) {
			continue
		}

		ev := expected.Field(i)
		av := actual.Field(i)
		childPath := fmt.Sprintf("%s.%s", path, sf.Name)

		// If field is unexported (PkgPath != ""), skip it
		if sf.PkgPath != "" {
			t.Logf("field %s is unexported; skipping comparison", childPath)
			continue
		}

		// If the child is a nested struct or pointer-to-struct, recurse
		if ev.Kind() == reflect.Ptr && ev.Type().Elem().Kind() == reflect.Struct {
			compareExceptTimeValue(t, ev, av, childPath)
			continue
		}
		if ev.Kind() == reflect.Struct {
			compareExceptTimeValue(t, ev, av, childPath)
			continue
		}

		// Concrete field: compare via Interface() — safe because exported
		switch {
		case ev.CanInterface() && av.CanInterface():
			assert.Equal(t, ev.Interface(), av.Interface(), fmt.Sprintf("field %s should be equal", childPath))
		case ev.CanInterface() != av.CanInterface():
			require.Failf(t, "field mismatch", "field %s: exportability differs between expected and actual while recursing", childPath)
		default:
			t.Logf("field %s unexported while recursing; skipping", childPath)
		}
	}
}

// AssertStringContains checks if a string contains all specified substrings
func AssertStringContains(t *testing.T, s string, substrings ...string) {
	t.Helper()

	for _, sub := range substrings {
		assert.True(
			t,
			strings.Contains(s, sub),
			fmt.Sprintf("expected string to contain '%s', but it didn't: %s", sub, s),
		)
	}
}
