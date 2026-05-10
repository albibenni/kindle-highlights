package types

import (
	"os"
	"testing"
)

func TestEnvVarMethods(t *testing.T) {
	key := "TEST_ENV_VAR"
	val := "test_value"
	
	// Test Value()
	_ = os.Setenv(key, val)
	defer func() { _ = os.Unsetenv(key) }()
	
	ev := EnvVar(key)
	if ev.Value() != val {
		t.Errorf("EnvVar.Value() = %q, want %q", ev.Value(), val)
	}

	// Test String()
	if ev.String() != key {
		t.Errorf("EnvVar.String() = %q, want %q", ev.String(), key)
	}
}

func TestEnvFileMethods(t *testing.T) {
	file := "test.env"
	ef := EnvFile(file)

	// Test String()
	if ef.String() != file {
		t.Errorf("EnvFile.String() = %q, want %q", ef.String(), file)
	}

	// Test Value() - Note: this actually looks up an ENV VAR named "test.env"
	// which is a bit unusual but matches the implementation.
	_ = os.Setenv(file, "some_path")
	defer func() { _ = os.Unsetenv(file) }()
	
	if ef.Value() != "some_path" {
		t.Errorf("EnvFile.Value() = %q, want %q", ef.Value(), "some_path")
	}
}

func TestGetEnvFile(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		{"darwin", string(Mac)},
		{"linux", string(Linux)},
		{"windows", "wrong pc"},
		{"other", string(Linux)},
	}

	for _, tt := range tests {
		got := getEnvFileForOS(tt.os)
		if got != tt.want {
			t.Errorf("getEnvFileForOS(%q) = %q, want %q", tt.os, got, tt.want)
		}
	}

	// Also call the main function just to cover it
	_ = GetEnvFile()
}

func TestConstants(t *testing.T) {
	// Verify indent.go constants exist and are not empty
	if Bold == "" || Italic == "" || Reset == "" {
		t.Error("ANSI escape constants should not be empty")
	}
}
