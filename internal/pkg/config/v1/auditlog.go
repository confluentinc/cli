package v1

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

func GetAuditLogs(state *ContextState) *orgv1.AuditLog {
	if state.Auth == nil || state.Auth.Organization == nil || state.Auth.Organization.GetAuditLog() == nil || state.Auth.Organization.GetAuditLog().GetServiceAccountId() == 0 {
		return nil
	}
	return state.Auth.Organization.AuditLog
}
