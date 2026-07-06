package dot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
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

// GenerateText calls the configured AI provider binary with the given prompt and input content, limiting the input size if necessary.
// It automatically resolves the AI binary path from the global state.
func GenerateText(ctx context.Context, state *GlobalState, prompt, input string, maxSize int) (string, error) {
	binary := GetAIBinary(state)
	aiPath, err := state.Runner.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolNotInstalled, binary)
	}

	if maxSize <= 0 {
		maxSize = DefaultMaxDiffSize
	}

	inputLimit := input
	if len(inputLimit) > maxSize {
		// Back off to a rune boundary so truncation never splits a multi-byte UTF-8
		// character and streams invalid UTF-8 to the AI binary.
		cut := maxSize
		for cut > 0 && !utf8.RuneStart(inputLimit[cut]) {
			cut--
		}
		inputLimit = inputLimit[:cut]
	}

	output, err := state.Runner.Run(ctx, "", strings.NewReader(inputLimit), aiPath, "--prompt", prompt)
	if err != nil {
		return "", fmt.Errorf("AI invocation failed: %w", err)
	}

	msg := strings.TrimSpace(output)
	if msg == "" {
		return "", errors.New("AI returned empty output")
	}

	return msg, nil
}
