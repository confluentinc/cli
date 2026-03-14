package skillgen

import "strings"

// InferOperation extracts the operation type from a command path.
// It looks at the last token in the command path and checks if it matches
// a known operation verb (list, create, delete, describe, update, use, start, stop).
// If no match is found, it returns "other".
//
// Examples:
//   - "confluent kafka topic list" → "list"
//   - "confluent kafka cluster create" → "create"
//   - "confluent login" → "other"
func InferOperation(commandPath string) string {
	// Handle empty or whitespace-only paths
	commandPath = strings.TrimSpace(commandPath)
	if commandPath == "" {
		return "other"
	}

	// Split by whitespace and filter out empty tokens
	tokens := strings.Fields(commandPath)
	if len(tokens) == 0 {
		return "other"
	}

	// Extract the last token as the potential verb
	lastToken := tokens[len(tokens)-1]

	// Check against recognized verbs
	recognizedVerbs := map[string]bool{
		"list":     true,
		"create":   true,
		"delete":   true,
		"describe": true,
		"update":   true,
		"use":      true,
		"start":    true,
		"stop":     true,
	}

	if recognizedVerbs[lastToken] {
		return lastToken
	}

	return "other"
}

// InferResource extracts the resource type from a command path.
// It takes all tokens between "confluent" and the last token (verb),
// and joins them with hyphens.
//
// Examples:
//   - "confluent kafka topic list" → "kafka-topic"
//   - "confluent iam service-account create" → "iam-service-account"
//   - "confluent login" → "" (no resource)
func InferResource(commandPath string) string {
	// Handle empty or whitespace-only paths
	commandPath = strings.TrimSpace(commandPath)
	if commandPath == "" {
		return ""
	}

	// Split by whitespace and filter out empty tokens
	tokens := strings.Fields(commandPath)
	if len(tokens) == 0 {
		return ""
	}

	// Need at least 3 tokens for a resource: "confluent <resource> <verb>"
	if len(tokens) < 3 {
		return ""
	}

	// Skip first token ("confluent") and last token (verb)
	// Extract everything in between as the resource
	resourceTokens := tokens[1 : len(tokens)-1]

	// Join resource tokens with hyphens
	return strings.Join(resourceTokens, "-")
}
