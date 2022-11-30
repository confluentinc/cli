package v1

import (
	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
)

func GetAuditLog(context *Context) *ccloudv1.AuditLog {
	if auditLog := context.GetOrganization().GetAuditLog(); auditLog.GetServiceAccountId() != 0 {
		return auditLog
	}
	return nil
}
