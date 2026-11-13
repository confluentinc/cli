//go:build eval

package scenarios

import (
	"fmt"
	"regexp"

	"github.com/confluentinc/cli/v4/test/eval"
)

// All is the eval matrix's scenario registry, in report order.
var All = []eval.Scenario{
	environmentCrosstalk,
	selectionCrossField,
	crudLostUpdate,
	authRace,
	mixedWorkload,
}

const (
	envA = "env-596"
	envB = "env-595"
)

// loginStep builds the login command for the mock backend URL, so the "--url" format string is not
// restated across scenario closures.
func loginStep(cloudURL string) string {
	return "login --url " + cloudURL
}

// observeActiveEnvironment reports the session's current context's active environment.
func observeActiveEnvironment(r eval.SessionResult) (string, error) {
	return eval.ReadActiveEnvironment(r.HomeDir)
}

// observeLoggedInEnvironment reports the session's active environment, or "logged out" when a
// concurrent logout cleared its credentials even though the environment survived.
func observeLoggedInEnvironment(r eval.SessionResult) (string, error) {
	cleared, err := eval.CredsCleared(r.HomeDir)
	if err != nil {
		return "", err
	}
	if cleared {
		return "logged out", nil
	}
	return observeActiveEnvironment(r)
}

var apiKeyPattern = regexp.MustCompile(`MYKEY[0-9]+`)

// createdGlobalKey extracts the mock-assigned key id this session's `api-key create` printed.
func createdGlobalKey(r eval.SessionResult) string {
	for _, inv := range r.Invocations {
		if m := apiKeyPattern.FindString(inv.Stdout); m != "" {
			return m
		}
	}
	return ""
}

// observeGlobalKeyPresent returns an observe probe for a session's own created global api-key. A key
// that never parsed from stdout is treated as a read error (routes to a visible corruption verdict),
// so a stdout-format change can't silently mask a dropped key as OK.
func observeGlobalKeyPresent(key string) func(eval.SessionResult) (string, error) {
	return func(r eval.SessionResult) (string, error) {
		if key == "" {
			return "", fmt.Errorf("no created api-key parsed from stdout")
		}
		present, err := eval.GlobalAPIKeyPresent(r.HomeDir, key)
		if err != nil {
			return "", err
		}
		if present {
			return key, nil
		}
		return "", nil
	}
}

// observeCredsCleared returns an observe probe that reports whether the session's context is logged
// out ("cleared") or still has credentials ("resurrected" by a concurrent write).
func observeCredsCleared() func(eval.SessionResult) (string, error) {
	return func(r eval.SessionResult) (string, error) {
		cleared, err := eval.CredsCleared(r.HomeDir)
		if err != nil {
			return "", err
		}
		if cleared {
			return "cleared", nil
		}
		return "resurrected", nil
	}
}
