package dot

const (
	// DefaultCommitPrompt is the default system prompt used to generate Conventional Commits messages.
	DefaultCommitPrompt = "Write ONE Conventional Commits message for this diff. Format: type(scope): subject, then a blank line and a short body if useful. Allowed types: %s. Output ONLY the raw commit message, absolutely no markdown code fences, no backticks, and no conversational preamble."

	// DefaultPRPrompt is the default system prompt used to generate GitHub Pull Request descriptions.
	DefaultPRPrompt = "Write a comprehensive, professional GitHub Pull Request description based on this diff. Format the description in Markdown. Include sections: Description, Context & Motivation, Key Changes, and Testing. Output ONLY the raw markdown content, absolutely no markdown code fences wrapping the entire output, no backticks surrounding it, and no conversational preamble."
)
