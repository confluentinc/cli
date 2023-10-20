package types

import (
	"fmt"
	"strings"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"
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
	Null                       StatementResultFieldType = "NULL"
)

type StatementResultFieldType string

func NewResultFieldType(obj flinkgatewayv1beta1.DataType) StatementResultFieldType {
	switch obj.Type {
	case "CHAR":
		return Char
	case "VARCHAR":
		return Varchar
	case "BOOLEAN":
		return Boolean
	case "BINARY":
		return Binary
	case "VARBINARY":
		return Varbinary
	case "DECIMAL":
		return Decimal
	case "TINYINT":
		return Tinyint
	case "SMALLINT":
		return Smallint
	case "INTEGER":
		return Integer
	case "BIGINT":
		return Bigint
	case "FLOAT":
		return Float
	case "DOUBLE":
		return Double
	case "DATE":
		return Date
	case "TIME_WITHOUT_TIME_ZONE":
		return TimeWithoutTimeZone
	case "TIMESTAMP_WITHOUT_TIME_ZONE":
		return TimestampWithoutTimeZone
	case "TIMESTAMP_WITH_TIME_ZONE":
		return TimestampWithTimeZone
	case "TIMESTAMP_WITH_LOCAL_TIME_ZONE":
		return TimestampWithLocalTimeZone
	case "INTERVAL_YEAR_MONTH":
		return IntervalYearMonth
	case "INTERVAL_DAY_TIME":
		return IntervalDayTime
	case "ARRAY":
		return Array
	case "MULTISET":
		return Multiset
	case "MAP":
		return Map
	case "ROW":
		return Row
	default:
		return Null
	}
}

type StatementResultField interface {
	GetType() StatementResultFieldType
	ToString() string
	ToSDKType() any
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
