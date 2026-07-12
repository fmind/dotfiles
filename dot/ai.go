package dot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// AIConfig represents the configuration for the AI feature.
type AIConfig struct {
	Binary string `yaml:"binary"`
}

// DefaultAIBinary is the default binary used for AI commands.
const DefaultAIBinary = "agy"

// DefaultMaxDiffSize bounds the diff bytes fed to the AI so a huge changeset can't blow past
// the model's context window. Shared by GenerateText's fallback and the commit config default.
const DefaultMaxDiffSize = 200000

func defaultAIConfig() AIConfig {
	return AIConfig{
		Binary: DefaultAIBinary,
	}
}

// GetAIBinary returns the configured or default AI binary name.
func GetAIBinary(state *GlobalState) string {
	return cmp.Or(state.Config.AI.Binary, DefaultAIBinary)
}

// ScanDiffForSecrets runs the exact diff destined for the AI provider through
// gitleaks first. The scan fails closed: missing tooling, scanner errors, and
// detected secrets all prevent the diff from leaving the machine.
func ScanDiffForSecrets(ctx context.Context, state *GlobalState, diff string) error {
	gitleaksPath, err := state.Runner.LookPath("gitleaks")
	if err != nil {
		return fmt.Errorf("%w: gitleaks", ErrToolNotInstalled)
	}
	if _, err := state.Runner.Run(ctx, "", strings.NewReader(diff), gitleaksPath, "stdin", "--no-banner", "--redact"); err != nil {
		return fmt.Errorf("outgoing diff secret scan failed: %w", err)
	}
	return nil
}

func limitAIInput(input string, maxSize int) string {
	if maxSize <= 0 {
		maxSize = DefaultMaxDiffSize
	}
	if len(input) <= maxSize {
		return input
	}

	// Back off to a rune boundary so truncation never splits a multi-byte UTF-8
	// character and streams invalid UTF-8 to the AI binary.
	cut := maxSize
	for cut > 0 && !utf8.RuneStart(input[cut]) {
		cut--
	}
	return input[:cut]
}

// GenerateText calls the configured AI provider binary with the given prompt and input content, limiting the input size if necessary.
// It automatically resolves the AI binary path from the global state.
func GenerateText(ctx context.Context, state *GlobalState, prompt, input string, maxSize int) (string, error) {
	binary := GetAIBinary(state)
	aiPath, err := state.Runner.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolNotInstalled, binary)
	}

	inputLimit := limitAIInput(input, maxSize)
	workDir, err := os.MkdirTemp("", "dot-ai-")
	if err != nil {
		return "", fmt.Errorf("failed to create isolated AI workspace: %w", err)
	}

	args := []string{"--prompt", prompt}
	if filepath.Base(binary) == DefaultAIBinary {
		args = append([]string{"--sandbox"}, args...)
	}
	output, invocationErr := state.Runner.Run(ctx, workDir, strings.NewReader(inputLimit), aiPath, args...)
	cleanupErr := os.RemoveAll(workDir)
	if invocationErr != nil {
		invocationErr = fmt.Errorf("AI invocation failed: %w", invocationErr)
		if cleanupErr != nil {
			return "", errors.Join(invocationErr, fmt.Errorf("failed to remove isolated AI workspace %s: %w", workDir, cleanupErr))
		}
		return "", invocationErr
	}
	if cleanupErr != nil {
		return "", fmt.Errorf("failed to remove isolated AI workspace %s: %w", workDir, cleanupErr)
	}

	msg := strings.TrimSpace(output)
	if msg == "" {
		return "", errors.New("AI returned empty output")
	}

	return msg, nil
}
