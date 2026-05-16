package agent

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

var ErrXrayDryRunFailed = errors.New("xray dry-run validation failed")

type XrayDryRunError struct {
	Reason string
}

func (e XrayDryRunError) Error() string {
	if e.Reason == "" {
		return ErrXrayDryRunFailed.Error()
	}
	return ErrXrayDryRunFailed.Error() + ": " + e.Reason
}

func (e XrayDryRunError) Unwrap() error {
	return ErrXrayDryRunFailed
}

type XrayDryRunValidator interface {
	Validate(ctx context.Context, configPath string) error
}

type CommandXrayDryRunValidator struct {
	Binary  string
	Timeout time.Duration
}

func (v CommandXrayDryRunValidator) Validate(ctx context.Context, configPath string) error {
	binary := strings.TrimSpace(v.Binary)
	if binary == "" {
		return nil
	}
	if strings.TrimSpace(configPath) == "" {
		return XrayDryRunError{Reason: "missing_config_path"}
	}

	timeout := v.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, binary, "run", "-test", "-config", configPath)
	output, err := cmd.CombinedOutput()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return XrayDryRunError{Reason: "timeout"}
	}
	if err != nil {
		return XrayDryRunError{Reason: compactXrayDryRunReason(output)}
	}
	return nil
}

func compactXrayDryRunReason(output []byte) string {
	text := strings.ToLower(strings.TrimSpace(string(output)))
	if text == "" {
		return "command_failed"
	}

	var builder strings.Builder
	lastUnderscore := false
	for _, r := range text {
		allowed := unicode.IsLetter(r) || unicode.IsDigit(r)
		if allowed {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "command_failed"
	}
	if len(result) > 96 {
		result = strings.TrimRight(result[:96], "_")
	}
	if result == "" {
		return "command_failed"
	}
	return result
}
