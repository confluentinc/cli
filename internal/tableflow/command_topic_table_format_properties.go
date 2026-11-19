package tableflow

import (
	"fmt"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"
)

const icebergFormat = "ICEBERG"

// addTableFormatPropertiesFlags declares --iceberg-format-version. On update, the flag can also enable a disabled ICEBERG.
func addTableFormatPropertiesFlags(cmd *cobra.Command, isUpdate bool) {
	help := "Specify the Iceberg format version, 2 or 3. It can be raised but not lowered. Raising it drops the Iceberg snapshot history of the table."
	if isUpdate {
		help += " It also enables ICEBERG if ICEBERG is disabled."
	}
	cmd.Flags().Int32("iceberg-format-version", 0, help)
}

func tableFormatPropertiesForCreate(cmd *cobra.Command) ([]tableflowv1.TableflowV1TableflowTopicSpecTableFormatPropertiesOneOf, error) {
	iceberg, err := icebergEntry(cmd)
	if err != nil || iceberg == nil {
		return nil, err
	}

	return []tableflowv1.TableflowV1TableflowTopicSpecTableFormatPropertiesOneOf{
		tableflowv1.TableflowV1IcebergTableFormatPropertiesAsTableflowV1TableflowTopicSpecTableFormatPropertiesOneOf(iceberg),
	}, nil
}

// The API merges entries by format, so the update does not change the other formats.
func tableFormatPropertiesForUpdate(cmd *cobra.Command) ([]tableflowv1.TableflowV1TableflowTopicSpecUpdateTableFormatPropertiesOneOf, error) {
	iceberg, err := icebergEntry(cmd)
	if err != nil || iceberg == nil {
		return nil, err
	}

	return []tableflowv1.TableflowV1TableflowTopicSpecUpdateTableFormatPropertiesOneOf{
		tableflowv1.TableflowV1IcebergTableFormatPropertiesAsTableflowV1TableflowTopicSpecUpdateTableFormatPropertiesOneOf(iceberg),
	}, nil
}

// icebergEntry returns the ICEBERG entry of "table_format_properties", enabled, with the format version of --iceberg-format-version.
// It returns nil if the flag is not set.
// The entry enables ICEBERG, so it fails if --table-formats is set without ICEBERG, because "table_formats" and
// "table_format_properties" would contradict. If --table-formats is not set, enable sends its default ICEBERG and update does not send it.
func icebergEntry(cmd *cobra.Command) (*tableflowv1.TableflowV1IcebergTableFormatProperties, error) {
	if !cmd.Flags().Changed("iceberg-format-version") {
		return nil, nil
	}

	formatVersion, err := cmd.Flags().GetInt32("iceberg-format-version")
	if err != nil {
		return nil, err
	}

	tableFormats, err := cmd.Flags().GetString("table-formats")
	if err != nil {
		return nil, err
	}

	if cmd.Flags().Changed("table-formats") && tableFormats != icebergFormat {
		return nil, fmt.Errorf("`--iceberg-format-version` requires `--table-formats ICEBERG`")
	}

	entry := tableflowv1.NewTableflowV1IcebergTableFormatProperties(icebergFormat, true)
	entry.SetFormatVersion(formatVersion)
	return entry, nil
}

// getIcebergFormatVersion returns the format version of the ICEBERG entry in "table_format_properties".
// For disabled ICEBERG, API returns stored version, but this method returns 0 so the user does not
// see an iceberg version for a table which does not have ICEBERG enabled at the moment.
func getIcebergFormatVersion(topic tableflowv1.TableflowV1TableflowTopic) int32 {
	for _, properties := range topic.Spec.GetTableFormatProperties() {
		if iceberg := properties.TableflowV1IcebergTableFormatProperties; iceberg != nil {
			if !iceberg.GetEnabled() {
				return 0
			}
			return iceberg.GetFormatVersion()
		}
	}

	return 0
}
