package testing

import "testing"

func TestExtractLibpqParam(t *testing.T) {
	tests := []struct {
		dsn  string
		key  string
		want string
	}{
		{"host=localhost port=5432 sslmode=disable", "sslmode", "disable"},
		{"host=localhost port=5432 sslmode='disable'", "sslmode", "disable"},
		{"host=localhost sslmode=\"require\" user=foo", "sslmode", "require"},
		{"postgres://user:pass@localhost/db?sslmode=disable", "sslmode", "disable"}, // raw DSN contains sslmode
		{"", "sslmode", ""},
	}

	for _, tc := range tests {
		got := extractLibpqParam(tc.dsn, tc.key)
		if got != tc.want {
			t.Fatalf("extractLibpqParam(%q, %q) = %q; want %q", tc.dsn, tc.key, got, tc.want)
		}
	}
}
