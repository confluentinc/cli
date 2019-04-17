package acl

import (
	"io"

	"github.com/codyaray/go-printer"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
)

func PrintAcls(bindingsObj []*kafkav1.ACLBinding, writer io.Writer) {
	var bindings [][]string
	for _, binding := range bindingsObj {

		record := &struct {
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
		}{
			binding.Entry.Principal,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
			binding.Pattern.PatternType.String(),
		}
		bindings = append(bindings, printer.ToRow(record,
			[]string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}))
	}
	printer.RenderCollectionTableOut(bindings, []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}, writer)
}
