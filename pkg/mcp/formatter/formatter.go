package formatter

// FormattedOutput represents the result of formatting CLI output.
// It provides three views of the same data:
// - Summary: Human-readable text optimized for chat (markdown formatted)
// - Structured: Parsed data structure for programmatic access
// - Raw: Original unmodified output
type FormattedOutput struct {
	Summary    string      // Human-readable formatted output (ANSI stripped, JSON summarized)
	Structured interface{} // Parsed structured data (nil for plain text, map/slice for JSON)
	Raw        string      // Original raw output with ANSI codes
}

// Formatter handles the conversion of raw CLI output into formatted results.
// It strips ANSI codes, detects JSON, and creates human-readable summaries.
type Formatter struct {
	// Empty for now - future expansion for configuration options
	// (e.g., column limits, ID shortening preferences, table styles)
}

// NewFormatter creates a new Formatter instance.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format processes raw CLI output and returns a formatted result.
// It performs the following steps:
// 1. Strips ANSI escape codes from the output
// 2. If an error occurred, formats the error message
// 3. Otherwise, attempts to parse and summarize JSON
// 4. Returns FormattedOutput with Summary, Structured data, and Raw output
func (f *Formatter) Format(rawOutput string, execErr error, commandPath string) *FormattedOutput {
	// Always preserve raw output
	result := &FormattedOutput{
		Raw: rawOutput,
	}

	// Step 1: Strip ANSI codes first
	cleanOutput := StripANSI(rawOutput)

	// Step 2: Handle errors
	if execErr != nil {
		result.Summary = f.formatError(cleanOutput, execErr, commandPath)
		return result
	}

	// Step 3: Parse and summarize JSON (or pass through plain text)
	summary, structured := SummarizeJSON(cleanOutput)
	result.Summary = summary
	result.Structured = structured

	return result
}

// formatError creates a human-readable error message.
// Delegates to FormatError() for consistent error formatting.
func (f *Formatter) formatError(cleanOutput string, execErr error, commandPath string) string {
	return FormatError(execErr, commandPath, cleanOutput)
}
