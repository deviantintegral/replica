package mysql

import (
	"testing"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/rs/zerolog"
)

// testLogger returns a disabled logger for testing.
func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestContainsParam(t *testing.T) {
	tests := []struct {
		name  string
		dsn   string
		param string
		want  bool
	}{
		{
			name:  "param at start with value",
			dsn:   "user:pass@tcp(localhost)/db?parseTime=true",
			param: "parseTime",
			want:  true,
		},
		{
			name:  "param in middle with value",
			dsn:   "user:pass@tcp(localhost)/db?charset=utf8&parseTime=true&loc=Local",
			param: "parseTime",
			want:  true,
		},
		{
			name:  "param at end with value",
			dsn:   "user:pass@tcp(localhost)/db?charset=utf8&parseTime=true",
			param: "parseTime",
			want:  true,
		},
		{
			name:  "param at start followed by ampersand",
			dsn:   "user:pass@tcp(localhost)/db?parseTime&charset=utf8",
			param: "parseTime",
			want:  true,
		},
		{
			name:  "param at end without value",
			dsn:   "user:pass@tcp(localhost)/db?charset=utf8&parseTime",
			param: "parseTime",
			want:  true,
		},
		{
			name:  "param not present",
			dsn:   "user:pass@tcp(localhost)/db?charset=utf8",
			param: "parseTime",
			want:  false,
		},
		{
			name:  "similar param name should not match",
			dsn:   "user:pass@tcp(localhost)/db?parseTimeZone=true",
			param: "parseTime",
			want:  false,
		},
		{
			name:  "no query params",
			dsn:   "user:pass@tcp(localhost)/db",
			param: "parseTime",
			want:  false,
		},
		{
			name:  "empty dsn",
			dsn:   "",
			param: "parseTime",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsParam(tt.dsn, tt.param)
			if got != tt.want {
				t.Errorf("containsParam(%q, %q) = %v, want %v", tt.dsn, tt.param, got, tt.want)
			}
		})
	}
}

func TestContainsRune(t *testing.T) {
	tests := []struct {
		name string
		s    string
		r    rune
		want bool
	}{
		{
			name: "contains question mark",
			s:    "user:pass@tcp(localhost)/db?parseTime=true",
			r:    '?',
			want: true,
		},
		{
			name: "does not contain question mark",
			s:    "user:pass@tcp(localhost)/db",
			r:    '?',
			want: false,
		},
		{
			name: "contains ampersand",
			s:    "?foo=bar&baz=qux",
			r:    '&',
			want: true,
		},
		{
			name: "empty string",
			s:    "",
			r:    '?',
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsRune(tt.s, tt.r)
			if got != tt.want {
				t.Errorf("containsRune(%q, %q) = %v, want %v", tt.s, tt.r, got, tt.want)
			}
		})
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "adds parseTime when no params exist",
			url:  "user:pass@tcp(localhost:3306)/dbname",
			want: "user:pass@tcp(localhost:3306)/dbname?parseTime=true",
		},
		{
			name: "adds parseTime when other params exist",
			url:  "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4",
			want: "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=true",
		},
		{
			name: "does not duplicate parseTime when already present at start",
			url:  "user:pass@tcp(localhost:3306)/dbname?parseTime=true",
			want: "user:pass@tcp(localhost:3306)/dbname?parseTime=true",
		},
		{
			name: "does not duplicate parseTime when present in middle",
			url:  "user:pass@tcp(localhost:3306)/dbname?charset=utf8&parseTime=true&loc=Local",
			want: "user:pass@tcp(localhost:3306)/dbname?charset=utf8&parseTime=true&loc=Local",
		},
		{
			name: "does not duplicate parseTime=false",
			url:  "user:pass@tcp(localhost:3306)/dbname?parseTime=false",
			want: "user:pass@tcp(localhost:3306)/dbname?parseTime=false",
		},
		{
			name: "handles multiple params correctly",
			url:  "user:pass@tcp(localhost)/db?charset=utf8mb4&collation=utf8mb4_unicode_ci",
			want: "user:pass@tcp(localhost)/db?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDSN(tt.url)
			if got != tt.want {
				t.Errorf("buildDSN(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "masks password in standard format",
			dsn:  "user:secretpassword@tcp(localhost:3306)/dbname",
			want: "user:***@tcp(localhost:3306)/dbname",
		},
		{
			name: "masks password with special chars",
			dsn:  "admin:p@ss:word!@tcp(localhost)/db",
			want: "admin:***@tcp(localhost)/db",
		},
		{
			name: "no password present",
			dsn:  "user@tcp(localhost:3306)/dbname",
			want: "user@tcp(localhost:3306)/dbname",
		},
		{
			name: "no at symbol",
			dsn:  "localhost:3306/dbname",
			want: "localhost:3306/dbname",
		},
		{
			name: "empty string",
			dsn:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskDSN(tt.dsn)
			if got != tt.want {
				t.Errorf("maskDSN(%q) = %q, want %q", tt.dsn, got, tt.want)
			}
		})
	}
}

func TestNew_EmptyURL(t *testing.T) {
	cfg := database.Config{
		Driver: "mysql",
		URL:    "",
	}

	_, err := New(cfg, testLogger())
	if err == nil {
		t.Error("New() with empty URL should return error, got nil")
	}

	wantMsg := "database URL is required"
	if err.Error() != wantMsg {
		t.Errorf("New() error = %q, want %q", err.Error(), wantMsg)
	}
}

func TestMariaDB_Driver(t *testing.T) {
	// Create a MariaDB instance directly without connecting
	// This tests the Driver() method without requiring a real database
	m := &MariaDB{
		db:     nil,
		logger: testLogger(),
	}

	got := m.Driver()
	want := "mariadb"
	if got != want {
		t.Errorf("Driver() = %q, want %q", got, want)
	}
}

func TestMariaDB_ImplementsDatabaseInterface(t *testing.T) {
	// This test verifies at compile time that MariaDB implements database.Database
	// We use a nil pointer since we only need to verify interface compliance
	var _ database.Database = (*MariaDB)(nil)
}

func TestMariadbTx_ImplementsTxInterface(t *testing.T) {
	// This test verifies at compile time that mariadbTx implements database.Tx
	var _ database.Tx = (*mariadbTx)(nil)
}

func TestDefaultConstants(t *testing.T) {
	// Verify default constants have expected values
	if DefaultMaxOpenConns != 25 {
		t.Errorf("DefaultMaxOpenConns = %d, want 25", DefaultMaxOpenConns)
	}
	if DefaultMaxIdleConns != 5 {
		t.Errorf("DefaultMaxIdleConns = %d, want 5", DefaultMaxIdleConns)
	}
	// 5 minutes in nanoseconds
	expectedLifetime := 5 * 60 * 1000 * 1000 * 1000
	if int64(DefaultConnMaxLifetime) != int64(expectedLifetime) {
		t.Errorf("DefaultConnMaxLifetime = %v, want 5 minutes", DefaultConnMaxLifetime)
	}
}
