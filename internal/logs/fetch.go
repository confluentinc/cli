package logs

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

const (
	scopeOrganization = "ORGANIZATION"
	scopeEnvironment  = "ENVIRONMENT"
	scopeResource     = "RESOURCE"

	maxTail = 1000

	searchTextKey = "search_text"
)

// resolveTimeWindow turns --since / --from / --to into absolute RFC 3339 UTC timestamps.
func resolveTimeWindow(since, from, to string, now time.Time) (string, string, error) {
	now = now.UTC()

	if since != "" {
		if from != "" || to != "" {
			return "", "", errors.New(sinceWithFromToErrorMsg)
		}
		d, err := parseSince(since)
		if err != nil {
			return "", "", err
		}
		return now.Add(-d).Format(time.RFC3339), now.Format(time.RFC3339), nil
	}

	if from == "" {
		return "", "", errors.NewErrorWithSuggestions(missingWindowErrorMsg, missingWindowSuggestions)
	}

	start, err := parseTimestamp("--from", from)
	if err != nil {
		return "", "", err
	}
	end := now
	if to != "" {
		if end, err = parseTimestamp("--to", to); err != nil {
			return "", "", err
		}
	}
	if end.Before(start) {
		return "", "", errors.New(toBeforeFromErrorMsg)
	}
	return start.Format(time.RFC3339), end.Format(time.RFC3339), nil
}

// parseSince accepts Go durations plus a day suffix, e.g. 90m, 24h, 7d.
func parseSince(value string) (time.Duration, error) {
	if days, ok := strings.CutSuffix(value, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil || n <= 0 {
			return 0, errors.NewErrorWithSuggestions(fmt.Sprintf(invalidSinceErrorMsg, value), invalidSinceSuggestions)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, errors.NewErrorWithSuggestions(fmt.Sprintf(invalidSinceErrorMsg, value), invalidSinceSuggestions)
	}
	return d, nil
}

func parseTimestamp(flag, value string) (time.Time, error) {
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, errors.NewErrorWithSuggestions(fmt.Sprintf(invalidTimestampErrorMsg, flag, value), invalidTimestampSuggestions)
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.NewErrorWithSuggestions(fmt.Sprintf(invalidTimestampErrorMsg, flag, value), invalidTimestampSuggestions)
	}
	return t.UTC(), nil
}

// buildSearch converts --search field=value pairs and --search-text into the API's search map.
// A field given once maps to a string; a repeated field maps to an array, which the API treats as OR.
func buildSearch(pairs []string, searchText string) (map[string]any, error) {
	values := map[string][]string{}
	var order []string

	for _, pair := range pairs {
		field, value, ok := strings.Cut(pair, "=")
		field, value = strings.TrimSpace(field), strings.TrimSpace(value)
		if !ok || field == "" || value == "" {
			return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(invalidSearchErrorMsg, pair), invalidSearchSuggestions)
		}
		if field == searchTextKey {
			return nil, errors.NewErrorWithSuggestions(reservedSearchFieldErrorMsg, reservedSearchFieldSuggestion)
		}
		if _, seen := values[field]; !seen {
			order = append(order, field)
		}
		values[field] = append(values[field], value)
	}

	search := map[string]any{}
	for _, field := range order {
		if len(values[field]) == 1 {
			search[field] = values[field][0]
		} else {
			search[field] = values[field]
		}
	}
	if searchText != "" {
		search[searchTextKey] = searchText
	}

	if len(search) == 0 {
		return nil, nil
	}
	return search, nil
}

// validateSearch checks the search fields and values against the log source's declared schema, so mistakes fail
// fast. The server validates every request again.
func validateSearch(source *loggingv1.LoggingV1Source, search map[string]any) error {
	fields := map[string]loggingv1.LoggingV1SearchFieldDescriptor{}
	names := make([]string, 0, len(source.SearchFields))
	for _, field := range source.SearchFields {
		fields[field.Name] = field
		names = append(names, field.Name)
	}

	keys := make([]string, 0, len(search))
	for key := range search {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		if key == searchTextKey {
			if !source.FullTextSearchSupported {
				return fmt.Errorf(searchTextUnsupportedErrorMsg, source.Name)
			}
			continue
		}

		field, ok := fields[key]
		if !ok {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(unknownSearchFieldErrorMsg, key, source.Name), fmt.Sprintf(unknownSearchFieldSuggestions, strings.Join(names, ", "), source.Name))
		}

		if field.Type != "enum" || len(field.GetAllowedValues()) == 0 {
			continue
		}
		for _, value := range searchValues(search[key]) {
			if !slices.Contains(field.GetAllowedValues(), value) {
				return errors.NewErrorWithSuggestions(fmt.Sprintf(invalidSearchValueErrorMsg, value, key), fmt.Sprintf(invalidSearchValueSuggestions, strings.Join(field.GetAllowedValues(), ", ")))
			}
		}
	}

	return nil
}

func searchValues(value any) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	}
	return nil
}

// scopeFor derives the query scope from the inputs, as the HLD specifies: a resource ID means resource scope,
// otherwise an explicit --environment means environment scope, otherwise organization scope. The active context's
// environment is never used to pick the scope.
func scopeFor(resourceId string, environmentSet bool) string {
	switch {
	case resourceId != "":
		return scopeResource
	case environmentSet:
		return scopeEnvironment
	default:
		return scopeOrganization
	}
}

func validateScope(source *loggingv1.LoggingV1Source, scope, resourceId string) error {
	if slices.Contains(source.AllowedScopes, scope) {
		return nil
	}

	if resourceId == "" && slices.Contains(source.AllowedScopes, scopeResource) {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(resourceIdRequiredErrorMsg, source.Name), fmt.Sprintf(resourceIdRequiredSuggestions, source.Name))
	}
	return errors.NewErrorWithSuggestions(fmt.Sprintf(unsupportedScopeErrorMsg, source.Name, strings.ToLower(scope)), fmt.Sprintf(unsupportedScopeSuggestions, strings.ToLower(strings.Join(source.AllowedScopes, ", "))))
}

func validateResourceId(source *loggingv1.LoggingV1Source, resourceId string) error {
	if resourceId == "" || source.GetIdPrefix() == "" {
		return nil
	}
	if !strings.HasPrefix(resourceId, source.GetIdPrefix()+"-") {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(resourceIdMismatchErrorMsg, resourceId, source.Name), fmt.Sprintf(resourceIdMismatchSuggestions, source.GetIdPrefix()))
	}
	return nil
}

func validateLookback(source *loggingv1.LoggingV1Source, since string) error {
	if since == "" || source.MaxLookbackDays <= 0 {
		return nil
	}
	d, err := parseSince(since)
	if err != nil {
		return err
	}
	if d > time.Duration(source.MaxLookbackDays)*24*time.Hour {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(lookbackExceededErrorMsg, source.MaxLookbackDays, source.Name), fmt.Sprintf(lookbackExceededSuggestions, source.MaxLookbackDays))
	}
	return nil
}

// crnValues carries the values the CRN's elements are built from.
type crnValues struct {
	Organization string
	Environment  string
	Cluster      string
	Leaf         string
}

// buildCRN builds the CRN the API expects. It follows the log source's declared crn_elements in order and stops at
// the element the requested scope is truncated to, so the same code serves every log source.
func buildCRN(source *loggingv1.LoggingV1Source, scope string, values crnValues) (string, error) {
	parts := make([]string, 0, len(source.CrnElements))
	for i, element := range source.CrnElements {
		var value string
		switch {
		case i == len(source.CrnElements)-1:
			value = values.Leaf
		case element == "organization":
			value = values.Organization
		case element == "environment":
			value = values.Environment
		case element == "cloud-cluster":
			value = values.Cluster
		default:
			return "", errors.NewErrorWithSuggestions(fmt.Sprintf(unsupportedCrnElementErrorMsg, source.Name, element), unsupportedCrnElementSuggestions)
		}

		parts = append(parts, fmt.Sprintf("%s=%s", element, value))

		if (scope == scopeOrganization && element == "organization") || (scope == scopeEnvironment && element == "environment") {
			break
		}
	}

	if scope != scopeResource && len(parts) == len(source.CrnElements) {
		return "", fmt.Errorf(missingScopeElementErrorMsg, source.Name, strings.ToLower(scope))
	}

	return "crn://confluent.cloud/" + strings.Join(parts, "/"), nil
}
