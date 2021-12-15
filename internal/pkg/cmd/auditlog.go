package cmd

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func AreAuditLogsEnabled(state *v1.ContextState) (*orgv1.AuditLog, bool) {
	if state.Auth == nil || state.Auth.Organization == nil || state.Auth.Organization.AuditLog == nil || state.Auth.Organization.AuditLog.ServiceAccountId == 0 {
		return nil, false
	}
	return state.Auth.Organization.AuditLog, true
}
