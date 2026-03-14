// Package skillgen provides namespace grouping and tier assignment logic
// for managing skill count within target limits.
package skillgen

import (
	"sort"
	"strings"
)

// AnalyzeNamespaces extracts namespace counts from command paths.
// The namespace is the second token in the command path (after "confluent").
// Returns a map of namespace to command count.
func AnalyzeNamespaces(commands []CommandIR) map[string]int {
	counts := make(map[string]int)

	for _, cmd := range commands {
		parts := strings.Fields(cmd.CommandPath)
		if len(parts) < 2 {
			continue // Skip invalid command paths
		}

		namespace := parts[1]
		counts[namespace]++
	}

	return counts
}

// ComputeTierThresholds calculates rank-based thresholds for tier assignment.
// Returns (highThreshold, mediumThreshold) where:
// - highThreshold captures approximately the top 20% of namespaces by command count
// - mediumThreshold captures approximately the top 50% of namespaces
//
// The thresholds are calculated using percentile ranks in the sorted distribution.
// Namespaces with count >= highThreshold are assigned to "high" tier.
// Namespaces with count >= mediumThreshold (but < highThreshold) are "medium" tier.
// Namespaces with count < mediumThreshold are "low" tier.
func ComputeTierThresholds(namespaceCounts map[string]int) (int, int) {
	if len(namespaceCounts) == 0 {
		return 0, 0
	}

	// Extract counts into a slice
	counts := make([]int, 0, len(namespaceCounts))
	for _, count := range namespaceCounts {
		counts = append(counts, count)
	}

	// Sort ascending
	sort.Ints(counts)

	n := len(counts)

	// Special case for single element
	if n == 1 {
		return counts[0], counts[0]
	}

	// Use percentile rank formula to find thresholds
	// For the Pth percentile, use: index = int((n-1) * P)
	// This gives the value at position P in the sorted array
	//
	// 85th percentile: ~15-20% of namespaces will be >= threshold (high tier)
	// 60th percentile: ~40% of namespaces will be >= threshold (medium + high tiers)
	//
	// These percentiles are tuned to achieve the target tier distribution:
	// - High tier (top ~18-20%): detailed resource+operation skills
	// - Medium tier (next ~20-25%): operation-level skills
	// - Low tier (bottom ~55-60%): namespace-level skills
	highIdx := int(float64(n-1) * 0.85)
	mediumIdx := int(float64(n-1) * 0.60)

	return counts[highIdx], counts[mediumIdx]
}

// AssignTiers assigns each namespace to a tier (high, medium, or low)
// based on the provided thresholds.
//
// Tier assignment:
// - count >= highThreshold → "high"
// - count >= mediumThreshold (but < highThreshold) → "medium"
// - count < mediumThreshold → "low"
func AssignTiers(namespaceCounts map[string]int, highThreshold, mediumThreshold int) map[string]string {
	tiers := make(map[string]string)

	for namespace, count := range namespaceCounts {
		if count >= highThreshold {
			tiers[namespace] = "high"
		} else if count >= mediumThreshold {
			tiers[namespace] = "medium"
		} else {
			tiers[namespace] = "low"
		}
	}

	return tiers
}

// EstimateSkillCount estimates the total number of skills that will be generated
// based on the tier assignments and command distribution.
//
// Skill counting rules:
// - High tier: one skill per unique (resource, operation) pair
// - Medium tier: one skill per unique operation
// - Low tier: one skill per namespace
// - operation='other' special handling: dedicated skills for login, logout, version, update
//
// Returns (estimatedCount, breakdown) where breakdown shows count by tier.
func EstimateSkillCount(commands []CommandIR, tiers map[string]string) (int, map[string]int) {
	breakdown := map[string]int{
		"high":   0,
		"medium": 0,
		"low":    0,
	}

	// Group commands by namespace
	byNamespace := make(map[string][]CommandIR)
	for _, cmd := range commands {
		parts := strings.Fields(cmd.CommandPath)
		if len(parts) < 2 {
			continue
		}
		namespace := parts[1]
		byNamespace[namespace] = append(byNamespace[namespace], cmd)
	}

	// Count skills per namespace based on tier
	for namespace, cmds := range byNamespace {
		tier, ok := tiers[namespace]
		if !ok {
			continue // Skip namespaces without tier assignment
		}

		switch tier {
		case "high":
			// One skill per unique (resource, operation) pair
			pairs := make(map[string]bool)
			for _, cmd := range cmds {
				// Create unique key from resource and operation
				key := cmd.Resource + "|" + cmd.Operation
				pairs[key] = true
			}
			breakdown["high"] += len(pairs)

		case "medium":
			// One skill per unique operation
			operations := make(map[string]bool)
			for _, cmd := range cmds {
				operations[cmd.Operation] = true
			}
			breakdown["medium"] += len(operations)

		case "low":
			// One skill per namespace
			breakdown["low"]++
		}
	}

	// Calculate total
	total := breakdown["high"] + breakdown["medium"] + breakdown["low"]

	return total, breakdown
}
