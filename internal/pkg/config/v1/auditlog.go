package v1

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

func GetAuditLog(context *Context) *orgv1.AuditLog {
	if auditLog := context.GetOrganization().GetAuditLog(); auditLog.GetServiceAccountId() != 0 {
		return auditLog
	}
	return nil
}
