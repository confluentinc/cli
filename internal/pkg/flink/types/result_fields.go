package types

import (
	"fmt"
	"strings"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
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

func NewResultFieldType(obj flinkgatewayv1alpha1.DataType) StatementResultFieldType {
	switch obj.Type {
	case "CHAR":
		return CHAR
	case "VARCHAR":
		return VARCHAR
	case "BOOLEAN":
		return BOOLEAN
	case "BINARY":
		return BINARY
	case "VARBINARY":
		return VARBINARY
	case "DECIMAL":
		return DECIMAL
	case "TINYINT":
		return TINYINT
	case "SMALLINT":
		return SMALLINT
	case "INTEGER":
		return INTEGER
	case "BIGINT":
		return BIGINT
	case "FLOAT":
		return FLOAT
	case "DOUBLE":
		return DOUBLE
	case "DATE":
		return DATE
	case "TIME_WITHOUT_TIME_ZONE":
		return TIME_WITHOUT_TIME_ZONE
	case "TIMESTAMP_WITHOUT_TIME_ZONE":
		return TIMESTAMP_WITHOUT_TIME_ZONE
	case "TIMESTAMP_WITH_TIME_ZONE":
		return TIMESTAMP_WITH_TIME_ZONE
	case "TIMESTAMP_WITH_LOCAL_TIME_ZONE":
		return TIMESTAMP_WITH_LOCAL_TIME_ZONE
	case "INTERVAL_YEAR_MONTH":
		return INTERVAL_YEAR_MONTH
	case "INTERVAL_DAY_TIME":
		return INTERVAL_DAY_TIME
	case "ARRAY":
		return ARRAY
	case "MULTISET":
		return MULTISET
	case "MAP":
		return MAP
	case "ROW":
		return ROW
	default:
		return NULL
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
	if f.Type == NULL {
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
