package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyDotEnv_ParsesExampleShapes(t *testing.T) {
	for _, key := range []string{"FIELDOPS_TEST_A", "FIELDOPS_TEST_B", "FIELDOPS_TEST_C", "FIELDOPS_TEST_D"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
	t.Setenv("FIELDOPS_TEST_D", "already-set")
	input := strings.Join([]string{
		"# comment",
		"",
		"FIELDOPS_TEST_A=plain",
		"FIELDOPS_TEST_B=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=   # 32 bytes, base64 (dev only)",
		`FIELDOPS_TEST_C="quoted # not a comment"`,
		"FIELDOPS_TEST_D=from-file",
		"not a pair",
	}, "\n")
	if err := applyDotEnv(bufio.NewScanner(strings.NewReader(input))); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"FIELDOPS_TEST_A": "plain",
		"FIELDOPS_TEST_B": "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"FIELDOPS_TEST_C": "quoted # not a comment",
		"FIELDOPS_TEST_D": "already-set",
	}
	for key, value := range want {
		if got := os.Getenv(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
}

func TestLoadConfig_Validates(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("ENCRYPTION_KEY", "key")
	t.Setenv("PORT", "")
	t.Setenv("SEED_ON_BOOT", "true")
	if _, err := loadConfig(); err == nil {
		t.Error("short JWT_SECRET accepted")
	}
	t.Setenv("JWT_SECRET", "a-secret-that-is-long-enough-for-hs256!!")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || !cfg.SeedOnBoot {
		t.Errorf("config = %+v", cfg)
	}
	t.Setenv("SEED_ON_BOOT", "sometimes")
	if _, err := loadConfig(); err == nil {
		t.Error("bad SEED_ON_BOOT accepted")
	}
}

func TestLoadDotEnv_MissingFilesAreFine(t *testing.T) {
	dir := t.TempDir()
	if err := loadDotEnv(filepath.Join(dir, "nope"), filepath.Join(dir, "also-nope")); err != nil {
		t.Fatal(err)
	}
}
