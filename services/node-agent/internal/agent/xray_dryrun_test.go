package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommandXrayDryRunValidatorMissingBinary(t *testing.T) {
	validator := CommandXrayDryRunValidator{Binary: filepath.Join(t.TempDir(), "missing-xray")}

	err := validator.Validate(context.Background(), "config.json")
	if !errors.Is(err, ErrXrayDryRunFailed) {
		t.Fatalf("expected dry-run error, got %v", err)
	}
	var dryRunErr XrayDryRunError
	if !errors.As(err, &dryRunErr) || dryRunErr.Reason != "xray_binary_not_found" {
		t.Fatalf("expected missing binary reason, got %v", err)
	}
}

func TestCommandXrayDryRunValidatorCommandFailureReason(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fixture is unix-specific")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "xray")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho 'Xray failed: invalid config shape' >&2\nexit 23\n"), 0o700); err != nil {
		t.Fatalf("expected fake xray binary: %v", err)
	}

	validator := CommandXrayDryRunValidator{Binary: binary}
	err := validator.Validate(context.Background(), "config.json")
	if !errors.Is(err, ErrXrayDryRunFailed) {
		t.Fatalf("expected dry-run error, got %v", err)
	}
	var dryRunErr XrayDryRunError
	if !errors.As(err, &dryRunErr) || dryRunErr.Reason != "xray_failed_invalid_config_shape" {
		t.Fatalf("expected compact command reason, got %v", err)
	}
}
