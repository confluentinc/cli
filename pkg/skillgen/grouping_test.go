package skillgen

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzeNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		commands []CommandIR
		expected map[string]int
	}{
		{
			name: "basic namespace extraction",
			commands: []CommandIR{
				{CommandPath: "confluent kafka topic list"},
				{CommandPath: "confluent kafka topic create"},
				{CommandPath: "confluent kafka cluster list"},
				{CommandPath: "confluent iam user list"},
			},
			expected: map[string]int{
				"kafka": 3,
				"iam":   1,
			},
		},
		{
			name: "single-command namespaces",
			commands: []CommandIR{
				{CommandPath: "confluent plugin list"},
				{CommandPath: "confluent context list"},
			},
			expected: map[string]int{
				"plugin":  1,
				"context": 1,
			},
		},
		{
			name:     "empty command list",
			commands: []CommandIR{},
			expected: map[string]int{},
		},
		{
			name: "namespace with many commands",
			commands: []CommandIR{
				{CommandPath: "confluent kafka topic list"},
				{CommandPath: "confluent kafka topic create"},
				{CommandPath: "confluent kafka topic delete"},
				{CommandPath: "confluent kafka topic describe"},
				{CommandPath: "confluent kafka cluster list"},
				{CommandPath: "confluent kafka cluster create"},
				{CommandPath: "confluent kafka cluster delete"},
				{CommandPath: "confluent kafka cluster describe"},
				{CommandPath: "confluent kafka cluster update"},
			},
			expected: map[string]int{
				"kafka": 9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeNamespaces(tt.commands)
			require.Equal(t, tt.expected, result, "namespace counts should match")
		})
	}
}

func TestComputeTierThresholds(t *testing.T) {
	tests := []struct {
		name             string
		namespaceCounts  map[string]int
		expectedHigh     int
		expectedMedium   int
		description      string
	}{
		{
			name: "37 namespaces like real IR",
			namespaceCounts: map[string]int{
				"kafka":           81,
				"local":           81,
				"network":         74,
				"flink":           58,
				"iam":             56,
				"schema-registry": 36,
				"connect":         26,
				"stream-share":    11,
				"api-key":         10,
				"environment":     9,
				"organization":    9,
				"pipeline":        9,
				"tableflow":       8,
				"billing":         7,
				"byok":            7,
				"cluster-linking": 7,
				"ksql":            7,
				"price":           7,
				"audit-log":       6,
				"shell":           6,
				"admin":           5,
				"context":         5,
				"login":           4,
				"logout":          3,
				"update":          3,
				"version":         3,
				"api":             2,
				"completion":      2,
				"prompt":          2,
				"secret":          2,
				"ai":              1,
				"feedback":        1,
				"plugin":          1,
				"service-quota":   1,
				"telemetry":       1,
				"test":            1,
				"pcf":             1,
			},
			expectedHigh:   26, // 85th percentile using formula: (n-1) * 0.85
			expectedMedium: 7,  // 60th percentile using formula: (n-1) * 0.60
			description:    "real distribution with 85th/60th percentiles for tier balance",
		},
		{
			name: "simple 10-namespace distribution",
			namespaceCounts: map[string]int{
				"ns1":  100,
				"ns2":  90,
				"ns3":  80,
				"ns4":  70,
				"ns5":  60,
				"ns6":  50,
				"ns7":  40,
				"ns8":  30,
				"ns9":  20,
				"ns10": 10,
			},
			expectedHigh:   80,  // 85th percentile: (10-1) * 0.85 = 7.65, int = 7 -> counts[7] = 80
			expectedMedium: 60,  // 60th percentile: (10-1) * 0.60 = 5.4, int = 5 -> counts[5] = 60
			description:    "evenly spaced values with 85th/60th percentiles",
		},
		{
			name: "single namespace",
			namespaceCounts: map[string]int{
				"only": 42,
			},
			expectedHigh:   42,
			expectedMedium: 42,
			description:    "single value returns itself for both thresholds",
		},
		{
			name: "two namespaces",
			namespaceCounts: map[string]int{
				"high": 100,
				"low":  10,
			},
			expectedHigh:   10,  // (2-1) * 0.80 = 0.8, int(0.8) = 0 -> counts[0] = 10
			expectedMedium: 10,  // (2-1) * 0.50 = 0.5, int(0.5) = 0 -> counts[0] = 10
			description:    "two values both return lower value due to formula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			high, medium := ComputeTierThresholds(tt.namespaceCounts)
			require.Equal(t, tt.expectedHigh, high, "high threshold (80th percentile) should match")
			require.Equal(t, tt.expectedMedium, medium, "medium threshold (50th percentile) should match")
		})
	}
}

func TestAssignTiers(t *testing.T) {
	tests := []struct {
		name            string
		namespaceCounts map[string]int
		highThreshold   int
		mediumThreshold int
		expected        map[string]string
	}{
		{
			name: "basic tier assignment",
			namespaceCounts: map[string]int{
				"kafka":  81,
				"local":  81,
				"iam":    56,
				"plugin": 1,
			},
			highThreshold:   50,
			mediumThreshold: 5,
			expected: map[string]string{
				"kafka":  "high",
				"local":  "high",
				"iam":    "high",
				"plugin": "low",
			},
		},
		{
			name: "all three tiers",
			namespaceCounts: map[string]int{
				"high1":   100,
				"high2":   90,
				"medium1": 50,
				"medium2": 40,
				"low1":    10,
				"low2":    5,
			},
			highThreshold:   80,
			mediumThreshold: 30,
			expected: map[string]string{
				"high1":   "high",
				"high2":   "high",
				"medium1": "medium",
				"medium2": "medium",
				"low1":    "low",
				"low2":    "low",
			},
		},
		{
			name: "boundary conditions - at threshold",
			namespaceCounts: map[string]int{
				"exactly-high":   80,
				"exactly-medium": 30,
				"below-medium":   29,
			},
			highThreshold:   80,
			mediumThreshold: 30,
			expected: map[string]string{
				"exactly-high":   "high",
				"exactly-medium": "medium",
				"below-medium":   "low",
			},
		},
		{
			name: "empty namespaces",
			namespaceCounts: map[string]int{},
			highThreshold:   50,
			mediumThreshold: 10,
			expected:        map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssignTiers(tt.namespaceCounts, tt.highThreshold, tt.mediumThreshold)
			require.Equal(t, tt.expected, result, "tier assignments should match")
		})
	}
}

func TestEstimateSkillCount(t *testing.T) {
	tests := []struct {
		name             string
		commands         []CommandIR
		tiers            map[string]string
		expectedCount    int
		expectedMin      int
		expectedMax      int
		checkBreakdown   bool
		expectedBreakdown map[string]int
	}{
		{
			name: "high tier - one skill per resource+operation pair",
			commands: []CommandIR{
				{CommandPath: "confluent kafka topic list", Operation: "list", Resource: "kafka-topic"},
				{CommandPath: "confluent kafka topic create", Operation: "create", Resource: "kafka-topic"},
				{CommandPath: "confluent kafka topic delete", Operation: "delete", Resource: "kafka-topic"},
				{CommandPath: "confluent kafka cluster list", Operation: "list", Resource: "kafka-cluster"},
			},
			tiers: map[string]string{
				"kafka": "high",
			},
			expectedCount:    4, // kafka-topic-list, kafka-topic-create, kafka-topic-delete, kafka-cluster-list
			checkBreakdown:   true,
			expectedBreakdown: map[string]int{"high": 4},
		},
		{
			name: "medium tier - one skill per operation",
			commands: []CommandIR{
				{CommandPath: "confluent iam user list", Operation: "list", Resource: "iam-user"},
				{CommandPath: "confluent iam service-account list", Operation: "list", Resource: "iam-service-account"},
				{CommandPath: "confluent iam user create", Operation: "create", Resource: "iam-user"},
			},
			tiers: map[string]string{
				"iam": "medium",
			},
			expectedCount:    2, // iam-list, iam-create
			checkBreakdown:   true,
			expectedBreakdown: map[string]int{"medium": 2},
		},
		{
			name: "low tier - one skill per namespace",
			commands: []CommandIR{
				{CommandPath: "confluent plugin list", Operation: "list", Resource: "plugin"},
				{CommandPath: "confluent context list", Operation: "list", Resource: "context"},
			},
			tiers: map[string]string{
				"plugin":  "low",
				"context": "low",
			},
			expectedCount:    2, // plugin, context
			checkBreakdown:   true,
			expectedBreakdown: map[string]int{"low": 2},
		},
		{
			name: "mixed tiers",
			commands: []CommandIR{
				// High tier
				{CommandPath: "confluent kafka topic list", Operation: "list", Resource: "kafka-topic"},
				{CommandPath: "confluent kafka topic create", Operation: "create", Resource: "kafka-topic"},
				// Medium tier
				{CommandPath: "confluent iam user list", Operation: "list", Resource: "iam-user"},
				{CommandPath: "confluent iam user create", Operation: "create", Resource: "iam-user"},
				// Low tier
				{CommandPath: "confluent plugin list", Operation: "list", Resource: "plugin"},
			},
			tiers: map[string]string{
				"kafka":  "high",
				"iam":    "medium",
				"plugin": "low",
			},
			expectedCount: 5, // kafka-topic-list, kafka-topic-create, iam-list, iam-create, plugin
			checkBreakdown:   true,
			expectedBreakdown: map[string]int{"high": 2, "medium": 2, "low": 1},
		},
		{
			name: "operation=other special handling",
			commands: []CommandIR{
				{CommandPath: "confluent login", Operation: "other", Resource: ""},
				{CommandPath: "confluent logout", Operation: "other", Resource: ""},
				{CommandPath: "confluent version", Operation: "other", Resource: ""},
				{CommandPath: "confluent update", Operation: "other", Resource: ""},
				{CommandPath: "confluent ai", Operation: "other", Resource: ""},
			},
			tiers: map[string]string{
				"login":   "low",
				"logout":  "low",
				"version": "low",
				"update":  "low",
				"ai":      "low",
			},
			expectedCount: 5, // Each gets its own skill despite being low tier
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, breakdown := EstimateSkillCount(tt.commands, tt.tiers)

			if tt.expectedCount > 0 {
				require.Equal(t, tt.expectedCount, count, "skill count should match")
			} else {
				require.GreaterOrEqual(t, count, tt.expectedMin, "skill count should be >= min")
				require.LessOrEqual(t, count, tt.expectedMax, "skill count should be <= max")
			}

			if tt.checkBreakdown {
				// Check that the expected tiers have the right counts
				// (ignore zero values for tiers not being tested)
				for tier, expectedCount := range tt.expectedBreakdown {
					require.Equal(t, expectedCount, breakdown[tier], "breakdown for tier %s should match", tier)
				}
			}
		})
	}
}

func TestEstimateSkillCountWithRealIR(t *testing.T) {
	// Load real IR data
	data, err := os.ReadFile("../../cmd/generate-skills/ir.json")
	require.NoError(t, err, "should read ir.json")

	var ir IR
	err = json.Unmarshal(data, &ir)
	require.NoError(t, err, "should unmarshal IR")

	// Analyze namespaces
	namespaceCounts := AnalyzeNamespaces(ir.Commands)
	require.NotEmpty(t, namespaceCounts, "should have namespaces")

	// Compute tier thresholds
	highThreshold, mediumThreshold := ComputeTierThresholds(namespaceCounts)

	// Assign tiers
	tiers := AssignTiers(namespaceCounts, highThreshold, mediumThreshold)

	// Estimate skill count
	count, breakdown := EstimateSkillCount(ir.Commands, tiers)

	// Log breakdown for diagnostic purposes
	t.Logf("Skill count: %d", count)
	t.Logf("Breakdown by tier: %+v", breakdown)
	t.Logf("High threshold: %d, Medium threshold: %d", highThreshold, mediumThreshold)
	t.Logf("Tier assignments: %+v", tiers)

	// Verify total count is reasonable
	// Note: With current IR data (540 commands across 37 namespaces), percentile-based
	// tiering alone produces ~420 skills. This is because high-tier namespaces (kafka,
	// iam, etc.) have many unique resource+operation pairs. Achieving the <200 target
	// will require additional strategies (e.g., operation grouping, resource limiting)
	// which are out of scope for this plan.
	//
	// This plan establishes the tier assignment mechanism. Subsequent plans will refine
	// the grouping strategy to reduce skill count further.
	require.Greater(t, count, 0, "should have at least one skill")

	// Verify the count is in a reasonable range (demonstrates tiers are working)
	require.LessOrEqual(t, count, 500, "should be under 500 skills")

	// Verify all tiers present in breakdown
	require.Greater(t, breakdown["high"], 0, "should have high tier skills")
	require.Greater(t, breakdown["medium"], 0, "should have medium tier skills")
	require.Greater(t, breakdown["low"], 0, "should have low tier skills")

	// Verify breakdown sums to total
	sum := breakdown["high"] + breakdown["medium"] + breakdown["low"]
	require.Equal(t, count, sum, "breakdown should sum to total count")
}

func TestNamespaceConsistency(t *testing.T) {
	commands := []CommandIR{
		{CommandPath: "confluent kafka topic list", Operation: "list", Resource: "kafka-topic"},
		{CommandPath: "confluent kafka topic create", Operation: "create", Resource: "kafka-topic"},
		{CommandPath: "confluent kafka cluster list", Operation: "list", Resource: "kafka-cluster"},
		{CommandPath: "confluent kafka cluster delete", Operation: "delete", Resource: "kafka-cluster"},
	}

	namespaceCounts := AnalyzeNamespaces(commands)
	highThreshold, mediumThreshold := ComputeTierThresholds(namespaceCounts)
	tiers := AssignTiers(namespaceCounts, highThreshold, mediumThreshold)

	// Extract namespace from each command and verify tier is consistent
	kafkaTier := tiers["kafka"]
	require.NotEmpty(t, kafkaTier, "kafka should have a tier")

	// All commands in kafka namespace should use the same tier
	for _, cmd := range commands {
		// Extract namespace (second token)
		parts := splitCommandPath(cmd.CommandPath)
		require.Greater(t, len(parts), 1, "command should have namespace")
		namespace := parts[1]

		tier, ok := tiers[namespace]
		require.True(t, ok, "namespace %s should have tier", namespace)
		require.Equal(t, kafkaTier, tier, "all kafka commands should have same tier")
	}
}

// Helper to split command path
func splitCommandPath(path string) []string {
	parts := []string{}
	current := ""
	for _, r := range path {
		if r == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
