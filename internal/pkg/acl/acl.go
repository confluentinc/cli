package acl

import (
	"io"
	"os"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/output"
)

func PrintACLsFromKafkaRestResponse(cmd *cobra.Command, aclGetResp krsdk.AclDataList, writer io.Writer) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, os.Stdout)
	if err != nil {
		return err
	}

	for _, aclData := range aclGetResp.Data {
		record := &struct {
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
		}{
			aclData.Principal,
			string(aclData.Permission),
			string(aclData.Operation),
			string(aclData.ResourceType),
			string(aclData.ResourceName),
			string(aclData.PatternType),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func PrintACLs(cmd *cobra.Command, bindingsObj []*schedv1.ACLBinding, writer io.Writer) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

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
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func PrintACLsFromKafkaRestResponseWithMap(cmd *cobra.Command, aclGetResp krsdk.AclDataList, writer io.Writer, IdMap map[int32]string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"UserId", "ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"user_id", "service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, os.Stdout)
	if err != nil {
		return err
	}

	for _, aclData := range aclGetResp.Data {
		UserId := aclData.Principal[5:]
		idp, err := strconv.Atoi(UserId)
		var resourceId string
		if err == nil {
			id := int32(idp)
			resourceId = IdMap[id]
		}
		record := &struct {
			UserId           string
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
		}{
			aclData.Principal,
			resourceId,
			string(aclData.Permission),
			string(aclData.Operation),
			string(aclData.ResourceType),
			string(aclData.ResourceName),
			string(aclData.PatternType),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func PrintACLsWithMap(cmd *cobra.Command, bindingsObj []*schedv1.ACLBinding, writer io.Writer, IdMap map[int32]string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"UserId", "ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"user_id", "service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, binding := range bindingsObj {
		UserId := binding.Entry.Principal[5:]
		idp, err := strconv.Atoi(UserId)
		var resourceId string
		if err == nil {
			id := int32(idp)
			resourceId = IdMap[id]
		}
		record := &struct {
			UserId           string
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
		}{
			binding.Entry.Principal,
			resourceId,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
			binding.Pattern.PatternType.String(),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}
