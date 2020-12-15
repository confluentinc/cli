package errors

import "text/template"

var CopyBYOKGCPPermissionsHeaderTmpl = template.Must(template.New("byok_gcp_permissions").Parse()
