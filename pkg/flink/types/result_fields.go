package types

import (
	"fmt"
	"strings"
)

const (
	Char                       StatementResultFieldType = "CHAR"
	Varchar                    StatementResultFieldType = "VARCHAR"
	Boolean                    StatementResultFieldType = "BOOLEAN"
	Binary                     StatementResultFieldType = "BINARY"
	Varbinary                  StatementResultFieldType = "VARBINARY"
	Decimal                    StatementResultFieldType = "DECIMAL"
	Tinyint                    StatementResultFieldType = "TINYINT"
	Smallint                   StatementResultFieldType = "SMALLINT"
	Integer                    StatementResultFieldType = "INTEGER"
	Bigint                     StatementResultFieldType = "BIGINT"
	Float                      StatementResultFieldType = "FLOAT"
	Double                     StatementResultFieldType = "DOUBLE"
	Date                       StatementResultFieldType = "DATE"
	TimeWithoutTimeZone        StatementResultFieldType = "TIME_WITHOUT_TIME_ZONE"
	TimestampWithoutTimeZone   StatementResultFieldType = "TIMESTAMP_WITHOUT_TIME_ZONE"
	TimestampWithTimeZone      StatementResultFieldType = "TIMESTAMP_WITH_TIME_ZONE"
	TimestampWithLocalTimeZone StatementResultFieldType = "TIMESTAMP_WITH_LOCAL_TIME_ZONE"
	IntervalYearMonth          StatementResultFieldType = "INTERVAL_YEAR_MONTH"
	IntervalDayTime            StatementResultFieldType = "INTERVAL_DAY_TIME"
	Array                      StatementResultFieldType = "ARRAY"
	Multiset                   StatementResultFieldType = "MULTISET"
	Map                        StatementResultFieldType = "MAP"
	Row                        StatementResultFieldType = "ROW"
	StructuredType             StatementResultFieldType = "STRUCTURED_TYPE"
	Variant                    StatementResultFieldType = "VARIANT"
	Null                       StatementResultFieldType = "NULL"
)

type StatementResultFieldType string

// NewResultFieldType maps the Flink Gateway's wire type name to our internal
// enum. An unrecognized name errors instead of silently becoming Null: without
// this, an older CLI build would render a newly introduced SQL type's non-null
// values as null, with no indication anything was wrong.
func NewResultFieldType(objType string) (StatementResultFieldType, error) {
	switch objType {
	case "CHAR":
		return Char, nil
	case "VARCHAR":
		return Varchar, nil
	case "BOOLEAN":
		return Boolean, nil
	case "BINARY":
		return Binary, nil
	case "VARBINARY":
		return Varbinary, nil
	case "DECIMAL":
		return Decimal, nil
	case "TINYINT":
		return Tinyint, nil
	case "SMALLINT":
		return Smallint, nil
	case "INTEGER":
		return Integer, nil
	case "BIGINT":
		return Bigint, nil
	case "FLOAT":
		return Float, nil
	case "DOUBLE":
		return Double, nil
	case "DATE":
		return Date, nil
	case "TIME_WITHOUT_TIME_ZONE":
		return TimeWithoutTimeZone, nil
	case "TIMESTAMP_WITHOUT_TIME_ZONE":
		return TimestampWithoutTimeZone, nil
	case "TIMESTAMP_WITH_TIME_ZONE":
		return TimestampWithTimeZone, nil
	case "TIMESTAMP_WITH_LOCAL_TIME_ZONE":
		return TimestampWithLocalTimeZone, nil
	case "INTERVAL_YEAR_MONTH":
		return IntervalYearMonth, nil
	case "INTERVAL_DAY_TIME":
		return IntervalDayTime, nil
	case "ARRAY":
		return Array, nil
	case "MULTISET":
		return Multiset, nil
	case "MAP":
		return Map, nil
	case "ROW":
		return Row, nil
	case "STRUCTURED_TYPE":
		return StructuredType, nil
	case "VARIANT":
		return Variant, nil
	case "NULL":
		return Null, nil
	default:
		return "", fmt.Errorf("unsupported result field type %q: this CLI may be out of date", objType)
	}
}

type StatementResultField interface {
	GetType() StatementResultFieldType
	ToString() string
	ToSDKType() any
	ToSerializedValue() any
}

type AtomicStatementResultField struct {
	Type  StatementResultFieldType
	Value string
}

func (f AtomicStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f AtomicStatementResultField) ToString() string {
	return f.Value
}

func (f AtomicStatementResultField) ToSDKType() any {
	if f.Type == Null {
		return nil
	}
	return f.Value
}

type ArrayStatementResultField struct {
	Type        StatementResultFieldType
	ElementType StatementResultFieldType
	Values      []StatementResultField
}

func (f ArrayStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f ArrayStatementResultField) ToString() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for idx, item := range f.Values {
		sb.WriteString(item.ToString())
		if idx != len(f.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func (f ArrayStatementResultField) ToSDKType() any {
	items := make([]any, len(f.Values))
	for idx, item := range f.Values {
		items[idx] = item.ToSDKType()
	}
	return items
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

func (f MapStatementResultField) ToString() string {
	sb := strings.Builder{}
	sb.WriteString("{")
	for idx, entry := range f.Entries {
		sb.WriteString(fmt.Sprintf("%s=%s", entry.Key.ToString(), entry.Value.ToString()))
		if idx != len(f.Entries)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (f MapStatementResultField) ToSDKType() any {
	mapItems := make([]any, len(f.Entries))
	for idx, entry := range f.Entries {
		mapItems[idx] = []any{entry.Key.ToSDKType(), entry.Value.ToSDKType()}
	}
	return mapItems
}

type RowStatementResultField struct {
	Type         StatementResultFieldType
	ElementTypes []StatementResultFieldType
	Values       []StatementResultField
}

func (f RowStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f RowStatementResultField) ToString() string {
	sb := strings.Builder{}
	sb.WriteString("(")
	for idx, item := range f.Values {
		sb.WriteString(item.ToString())
		if idx != len(f.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (f RowStatementResultField) ToSDKType() any {
	rowItems := make([]any, len(f.Values))
	for idx, value := range f.Values {
		rowItems[idx] = value.ToSDKType()
	}
	return rowItems
}

type StructuredTypeStatementResultField struct {
	Type       StatementResultFieldType
	FieldNames []string
	FieldTypes []StatementResultFieldType
	Values     []StatementResultField
}

func (f StructuredTypeStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f StructuredTypeStatementResultField) ToString() string {
	sb := strings.Builder{}
	sb.WriteString("(")
	for idx, item := range f.Values {
		sb.WriteString(f.FieldNames[idx] + "=" + item.ToString())
		if idx != len(f.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (f StructuredTypeStatementResultField) ToSDKType() any {
	items := make(map[string]any, len(f.FieldNames))
	for idx, value := range f.Values {
		items[f.FieldNames[idx]] = value.ToSDKType()
	}
	return items
}
