package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// Variant type codes are a frozen wire contract: a VARIANT is self-describing, so
// each node carries its own type code.
const (
	VariantCodeNull           = 0
	VariantCodeObject         = 1
	VariantCodeArray          = 2
	VariantCodeBoolean        = 3
	VariantCodeTinyint        = 4
	VariantCodeSmallint       = 5
	VariantCodeInt            = 6
	VariantCodeBigint         = 7
	VariantCodeFloat          = 8
	VariantCodeDouble         = 9
	VariantCodeDecimal        = 10
	VariantCodeString         = 11
	VariantCodeDate           = 12
	VariantCodeTimestamp      = 13
	VariantCodeTimestampLtz   = 14
	VariantCodeBytes          = 15
	VariantCodeTime           = 16
	VariantCodeTimestampNs    = 17
	VariantCodeTimestampLtzNs = 18
	VariantCodeUnknown        = -1
	VariantCodeInvalid        = -2
)

const (
	variantUnsupportedMarker = "<unsupported>"
	variantInvalidMarker     = "<invalid>"
)

// variantNode is one decoded node of a VARIANT value; each shape is its own type
// and the interface is sealed by an unexported method.
type variantNode interface {
	appendJSON(sb *strings.Builder)
}

type variantNull struct{}

func (variantNull) appendJSON(sb *strings.Builder) {
	sb.WriteString("null")
}

type variantScalar struct {
	code int
	raw  string
}

func (s variantScalar) appendJSON(sb *strings.Builder) {
	switch s.code {
	case VariantCodeBoolean:
		if strings.EqualFold(s.raw, "TRUE") {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case VariantCodeTinyint, VariantCodeSmallint, VariantCodeInt, VariantCodeBigint, VariantCodeFloat, VariantCodeDouble:
		sb.WriteString(renderVariantNumber(s.raw))
	case VariantCodeDecimal:
		sb.WriteString(renderVariantNumber(stripTrailingZeros(s.raw)))
	case VariantCodeString, VariantCodeDate, VariantCodeTime:
		sb.WriteString(jsonQuote(s.raw))
	case VariantCodeTimestamp, VariantCodeTimestampNs:
		sb.WriteString(jsonQuote(formatVariantTimestamp(s.raw)))
	case VariantCodeTimestampLtz, VariantCodeTimestampLtzNs:
		// LTZ renders at UTC with a trailing Z; the wire's session offset is dropped.
		sb.WriteString(jsonQuote(formatVariantTimestamp(s.raw) + "Z"))
	case VariantCodeBytes:
		sb.WriteString(jsonQuote(formatVariantBytes(s.raw)))
	default:
		// Safety net for an unrecognized code; keeps the cell valid JSON.
		sb.WriteString(jsonQuote(variantUnsupportedMarker))
	}
}

type variantMarker struct {
	text string
}

func (m variantMarker) appendJSON(sb *strings.Builder) {
	sb.WriteString(jsonQuote(m.text))
}

type variantField struct {
	key   string
	value variantNode
}

type variantObject struct {
	fields []variantField
}

func (o variantObject) appendJSON(sb *strings.Builder) {
	sb.WriteString("{")
	for idx, field := range o.fields {
		if idx != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(jsonQuote(field.key))
		sb.WriteString(":")
		field.value.appendJSON(sb)
	}
	sb.WriteString("}")
}

type variantArray struct {
	elements []variantNode
}

func (a variantArray) appendJSON(sb *strings.Builder) {
	sb.WriteString("[")
	for idx, element := range a.elements {
		if idx != 0 {
			sb.WriteString(",")
		}
		element.appendJSON(sb)
	}
	sb.WriteString("]")
}

type VariantStatementResultField struct {
	Type StatementResultFieldType
	root variantNode
}

func NewVariantStatementResultField(raw any) VariantStatementResultField {
	return VariantStatementResultField{Type: Variant, root: decodeVariant(raw)}
}

func (f VariantStatementResultField) GetType() StatementResultFieldType {
	return f.Type
}

func (f VariantStatementResultField) ToString() string {
	sb := strings.Builder{}
	f.root.appendJSON(&sb)
	return sb.String()
}

func (f VariantStatementResultField) ToSDKType() any {
	// Return the rendered JSON text so machine-readable output keeps exact
	// formatting, including large-integer precision a float round-trip would lose.
	return f.ToString()
}

func (f VariantStatementResultField) ToSerializedValue() any {
	// A VARIANT renders as JSON text, per the Result Schema and Payload Format spec
	// (https://confluentinc.atlassian.net/wiki/spaces/FLINK/pages/3037888565): it is
	// semi-structured and self-describing, so it is deliberately rendered as a JSON
	// string rather than decomposed like ARRAY/ROW/STRUCTURED.
	return f.ToSDKType()
}

// ToPrettyString renders the value as indented JSON for the row-details view.
func (f VariantStatementResultField) ToPrettyString() string {
	compact := f.ToString()
	indented := bytes.Buffer{}
	if err := json.Indent(&indented, []byte(compact), "", "  "); err != nil {
		return compact
	}
	return indented.String()
}

// decodeVariant walks the self-describing [code, ...] wire format. It is
// SDK-agnostic and total: anything unreadable becomes an INVALID marker and a bad
// child degrades in place, so a single bad node never fails the row.
func decodeVariant(raw any) variantNode {
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return variantMarker{text: variantInvalidMarker}
	}
	code, ok := variantCodeOf(arr[0])
	if !ok {
		return variantMarker{text: variantInvalidMarker}
	}
	switch code {
	case VariantCodeNull:
		return variantNull{}
	case VariantCodeObject:
		return variantObject{fields: decodeVariantFields(arr)}
	case VariantCodeArray:
		return variantArray{elements: decodeVariantElements(arr)}
	case VariantCodeUnknown:
		return variantMarker{text: variantUnsupportedMarker}
	case VariantCodeInvalid:
		return variantMarker{text: variantInvalidMarker}
	default:
		// Scalar [code, value]; the LTZ codes also carry an ignored offset at index 2.
		return variantScalar{code: code, raw: variantStringAt(arr, 1)}
	}
}

func decodeVariantFields(arr []any) []variantField {
	pairs := variantChildrenOf(arr)
	fields := make([]variantField, 0, len(pairs))
	for _, entry := range pairs {
		pair, ok := entry.([]any)
		if !ok || len(pair) != 2 {
			continue
		}
		key, ok := pair[0].(string)
		if !ok {
			continue
		}
		fields = append(fields, variantField{key: key, value: decodeVariant(pair[1])})
	}
	return fields
}

func decodeVariantElements(arr []any) []variantNode {
	children := variantChildrenOf(arr)
	elements := make([]variantNode, 0, len(children))
	for _, child := range children {
		elements = append(elements, decodeVariant(child))
	}
	return elements
}

// variantChildrenOf returns the child list at index 1, or nil (read as an empty
// container) when it is absent or not a list.
func variantChildrenOf(arr []any) []any {
	if len(arr) < 2 {
		return nil
	}
	children, _ := arr[1].([]any)
	return children
}

func variantStringAt(arr []any, i int) string {
	if i >= len(arr) {
		return ""
	}
	s, _ := arr[i].(string)
	return s
}

// variantCodeOf reads the leading type code, which encoding/json yields as a float64.
func variantCodeOf(raw any) (int, bool) {
	code, ok := raw.(float64)
	if !ok {
		return 0, false
	}
	return int(code), true
}

func jsonQuote(s string) string {
	// HTML escaping off so <, > and & survive verbatim; Encode appends a newline.
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(s); err != nil {
		return `""`
	}
	return strings.TrimRight(buf.String(), "\n")
}

// renderVariantNumber emits the scalar as a bare JSON number when it is one, else
// quotes it, so an empty, malformed, or non-finite scalar stays valid JSON. The
// numeric first-byte check excludes true/false/null that json.Valid also accepts.
func renderVariantNumber(scalar string) string {
	if scalar != "" && (scalar[0] == '-' || (scalar[0] >= '0' && scalar[0] <= '9')) && json.Valid([]byte(scalar)) {
		return scalar
	}
	return jsonQuote(scalar)
}

// stripTrailingZeros drops trailing fractional zeros ("100.00" -> "100"). A string
// with no decimal point is returned unchanged so integer zeros are never touched.
func stripTrailingZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimRight(s, ".")
}

func formatVariantTimestamp(s string) string {
	return stripTrailingZeros(strings.Replace(s, " ", "T", 1))
}

// formatVariantBytes converts the backend's hex literal (x'7f0203') to base64,
// falling back to the raw value on a parse failure.
func formatVariantBytes(s string) string {
	digits := s
	if strings.HasPrefix(digits, "x'") && strings.HasSuffix(digits, "'") {
		digits = digits[2 : len(digits)-1]
	}
	raw, err := hex.DecodeString(digits)
	if err != nil {
		return s
	}
	return base64.StdEncoding.EncodeToString(raw)
}
