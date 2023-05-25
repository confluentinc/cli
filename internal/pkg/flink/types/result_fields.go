package types

import (
	"fmt"
	"math"
	"strings"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
)

const (
	CHAR                           StatementResultFieldType = "CHAR"
	VARCHAR                        StatementResultFieldType = "VARCHAR"
	BOOLEAN                        StatementResultFieldType = "BOOLEAN"
	BINARY                         StatementResultFieldType = "BINARY"
	VARBINARY                      StatementResultFieldType = "VARBINARY"
	DECIMAL                        StatementResultFieldType = "DECIMAL"
	TINYINT                        StatementResultFieldType = "TINYINT"
	SMALLINT                       StatementResultFieldType = "SMALLINT"
	INTEGER                        StatementResultFieldType = "INTEGER"
	BIGINT                         StatementResultFieldType = "BIGINT"
	FLOAT                          StatementResultFieldType = "FLOAT"
	DOUBLE                         StatementResultFieldType = "DOUBLE"
	DATE                           StatementResultFieldType = "DATE"
	TIME_WITHOUT_TIME_ZONE         StatementResultFieldType = "TIME_WITHOUT_TIME_ZONE"
	TIMESTAMP_WITHOUT_TIME_ZONE    StatementResultFieldType = "TIMESTAMP_WITHOUT_TIME_ZONE"
	TIMESTAMP_WITH_TIME_ZONE       StatementResultFieldType = "TIMESTAMP_WITH_TIME_ZONE"
	TIMESTAMP_WITH_LOCAL_TIME_ZONE StatementResultFieldType = "TIMESTAMP_WITH_LOCAL_TIME_ZONE"
	INTERVAL_YEAR_MONTH            StatementResultFieldType = "INTERVAL_YEAR_MONTH"
	INTERVAL_DAY_TIME              StatementResultFieldType = "INTERVAL_DAY_TIME"
	ARRAY                          StatementResultFieldType = "ARRAY"
	MULTISET                       StatementResultFieldType = "MULTISET"
	MAP                            StatementResultFieldType = "MAP"
	ROW                            StatementResultFieldType = "ROW"
	NULL                           StatementResultFieldType = "NULL"
)

type StatementResultFieldType string

func NewResultFieldType(obj v1.DataType) StatementResultFieldType {
	if obj.ArrayType != nil {
		return ARRAY
	}
	if obj.BigIntType != nil {
		return BIGINT
	}
	if obj.BinaryType != nil {
		return BINARY
	}
	if obj.BooleanType != nil {
		return BOOLEAN
	}
	if obj.CharType != nil {
		return CHAR
	}
	if obj.DateType != nil {
		return DATE
	}
	if obj.DecimalType != nil {
		return DECIMAL
	}
	if obj.DoubleType != nil {
		return DOUBLE
	}
	if obj.FloatType != nil {
		return FLOAT
	}
	if obj.IntegerType != nil {
		return INTEGER
	}
	if obj.IntervalDayTimeType != nil {
		return INTERVAL_DAY_TIME
	}
	if obj.IntervalYearMonthType != nil {
		return INTERVAL_YEAR_MONTH
	}
	if obj.MapType != nil {
		return MAP
	}
	if obj.MultisetType != nil {
		return MULTISET
	}
	if obj.RowType != nil {
		return ROW
	}
	if obj.SmallIntType != nil {
		return SMALLINT
	}
	if obj.TimeWithoutTimeZoneType != nil {
		return TIME_WITHOUT_TIME_ZONE
	}
	if obj.TimestampWithLocalTimeZoneType != nil {
		return TIMESTAMP_WITH_LOCAL_TIME_ZONE
	}
	if obj.TimestampWithTimeZoneType != nil {
		return TIMESTAMP_WITH_TIME_ZONE
	}
	if obj.TimestampWithoutTimeZoneType != nil {
		return TIMESTAMP_WITHOUT_TIME_ZONE
	}
	if obj.TinyIntType != nil {
		return TINYINT
	}
	if obj.VarbinaryType != nil {
		return VARBINARY
	}
	if obj.VarcharType != nil {
		return VARCHAR
	}
	return NULL
}

type FormatterOptions struct {
	MaxCharCountToDisplay int
}

func (f *FormatterOptions) GetMaxCharCountToDisplay() int {
	if f == nil {
		return math.MaxInt32
	}
	return f.MaxCharCountToDisplay
}

type StatementResultField interface {
	GetType() StatementResultFieldType
	Format(*FormatterOptions) string
	ToSDKType() v1.SqlV1alpha1ResultItemRowOneOf
}

type AtomicStatementResultField struct {
	Type  StatementResultFieldType
	Value string
}

func (f AtomicStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f AtomicStatementResultField) Format(options *FormatterOptions) string {
	if options != nil {
		return truncateString(f.Value, options.GetMaxCharCountToDisplay())
	}
	return f.Value
}

func (f AtomicStatementResultField) ToSDKType() v1.SqlV1alpha1ResultItemRowOneOf {
	val := v1.SqlV1alpha1ResultItemString(f.Value)
	if f.Type == NULL {
		return v1.SqlV1alpha1ResultItemRowOneOf{}
	}
	return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemString: &val}
}

type ArrayStatementResultField struct {
	Type        StatementResultFieldType
	ElementType StatementResultFieldType
	Values      []StatementResultField
}

func (f ArrayStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f ArrayStatementResultField) Format(options *FormatterOptions) string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for idx, item := range f.Values {
		sb.WriteString(item.Format(nil))
		if idx != len(f.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")

	if options != nil {
		return truncateString(sb.String(), options.GetMaxCharCountToDisplay())
	}
	return sb.String()
}

func (f ArrayStatementResultField) ToSDKType() v1.SqlV1alpha1ResultItemRowOneOf {
	items := make([]v1.SqlV1alpha1ResultItemRowOneOf, len(f.Values))
	for idx, item := range f.Values {
		items[idx] = item.ToSDKType()
	}
	return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: items}}
}

type MapStatementResultFieldEntry struct {
	Key   StatementResultField
	Value StatementResultField
}

type MapStatementResultField struct {
	Type      StatementResultFieldType
	KeyType   StatementResultFieldType
	ValueType StatementResultFieldType
	Entries   []MapStatementResultFieldEntry
}

func (f MapStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f MapStatementResultField) Format(options *FormatterOptions) string {
	sb := strings.Builder{}
	sb.WriteString("{")
	for idx, entry := range f.Entries {
		sb.WriteString(fmt.Sprintf("%s=%s", entry.Key.Format(nil), entry.Value.Format(nil)))
		if idx != len(f.Entries)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")

	if options != nil {
		return truncateString(sb.String(), options.GetMaxCharCountToDisplay())
	}
	return sb.String()
}

func (f MapStatementResultField) ToSDKType() v1.SqlV1alpha1ResultItemRowOneOf {
	mapItems := make([]v1.SqlV1alpha1ResultItemRowOneOf, len(f.Entries))
	for idx, entry := range f.Entries {
		var keyValuePair []v1.SqlV1alpha1ResultItemRowOneOf
		keyValuePair = append(keyValuePair, entry.Key.ToSDKType(), entry.Value.ToSDKType())
		resultItemRowOneOf := v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: keyValuePair}}
		mapItems[idx] = resultItemRowOneOf
	}
	return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: mapItems}}
}

type RowStatementResultField struct {
	Type         StatementResultFieldType
	ElementTypes []StatementResultFieldType
	Values       []StatementResultField
}

func (f RowStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f RowStatementResultField) Format(options *FormatterOptions) string {
	sb := strings.Builder{}
	sb.WriteString("(")
	for idx, item := range f.Values {
		sb.WriteString(item.Format(nil))
		if idx != len(f.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

	if options != nil {
		return truncateString(sb.String(), options.GetMaxCharCountToDisplay())
	}
	return sb.String()
}

func (f RowStatementResultField) ToSDKType() v1.SqlV1alpha1ResultItemRowOneOf {
	rowItems := make([]v1.SqlV1alpha1ResultItemRowOneOf, len(f.Values))
	for idx, value := range f.Values {
		rowItems[idx] = value.ToSDKType()
	}
	return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: rowItems}}
}

func truncateString(str string, maxCharCountToDisplay int) string {
	if len(str) > maxCharCountToDisplay {
		if maxCharCountToDisplay <= 3 {
			return "..."
		}
		return str[:maxCharCountToDisplay-3] + "..."
	}
	return str
}
