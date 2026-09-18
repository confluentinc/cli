package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariantRendering(t *testing.T) {
	tests := []struct {
		name string
		wire string
		want string
	}{
		{"null", `[0]`, `null`},
		{"boolean true", `[3,"TRUE"]`, `true`},
		{"boolean false", `[3,"FALSE"]`, `false`},
		{"tinyint", `[4,"7"]`, `7`},
		{"bigint keeps precision", `[7,"9223372036854775807"]`, `9223372036854775807`},
		{"double", `[9,"21.5"]`, `21.5`},
		{"decimal strips trailing zeros", `[10,"100.00"]`, `100`},
		{"decimal without fraction", `[10,"42"]`, `42`},
		{"string quoted", `[11,"sensor-7"]`, `"sensor-7"`},
		{"date", `[12,"2026-07-28"]`, `"2026-07-28"`},
		{"timestamp uses T separator", `[13,"2026-07-28 09:14:02.117000"]`, `"2026-07-28T09:14:02.117"`},
		{"timestamp nanosecond", `[17,"2026-07-28 09:14:02.123456789"]`, `"2026-07-28T09:14:02.123456789"`},
		{"timestamp_ltz always utc", `[14,"2026-07-28 09:14:02.117000","-05:00"]`, `"2026-07-28T09:14:02.117Z"`},
		{"timestamp_ltz nanosecond always utc", `[18,"2026-07-28 09:14:02.123456789","-05:00"]`, `"2026-07-28T09:14:02.123456789Z"`},
		{"time", `[16,"09:14:02.123"]`, `"09:14:02.123"`},
		{"bytes as base64", `[15,"x'7f0203'"]`, `"fwID"`},
		{"bytes fallback on bad hex", `[15,"zz"]`, `"zz"`},
		{"non-finite number quoted", `[9,"NaN"]`, `"NaN"`},
		{"non-json number quoted", `[8,"1."]`, `"1."`},
		{"missing scalar quoted", `[6]`, `""`},
		{"unknown code marker", `[-1,"x'0102'","x'7f2a'"]`, `"<unsupported>"`},
		{"invalid marker", `[-2]`, `"<invalid>"`},
		{"unrecognized positive code", `[99,"x"]`, `"<unsupported>"`},
		{"empty object", `[1,[]]`, `{}`},
		{"empty array", `[2,[]]`, `[]`},
		{"array of scalars", `[2,[[11,"hot"],[0]]]`, `["hot",null]`},
		{"non-array is invalid", `"hello"`, `"<invalid>"`},
		{"empty node is invalid", `[]`, `"<invalid>"`},
		{"non-numeric code is invalid", `["x"]`, `"<invalid>"`},
		{"object without body", `[1]`, `{}`},
		{"object with non-list body", `[1,"x"]`, `{}`},
		{"object drops malformed pair", `[1,[["k"]]]`, `{}`},
		{"object drops non-string key", `[1,[[5,"v"]]]`, `{}`},
		{
			"nested object from spec",
			`[1,[["device",[11,"sensor-7"]],["meta",[1,[["n",[4,"7"]],["ok",[3,"TRUE"]]]]],["price",[10,"100.00"]],["seen_at",[13,"2026-07-28 09:14:02.117000"]],["seq",[7,"9223372036854775807"]],["tags",[2,[[11,"hot"],[0]]]],["temp",[9,"21.5"]],["weird",[-1,"x'0102'","x'7f2a'"]]]]`,
			`{"device":"sensor-7","meta":{"n":7,"ok":true},"price":100,"seen_at":"2026-07-28T09:14:02.117","seq":9223372036854775807,"tags":["hot",null],"temp":21.5,"weird":"<unsupported>"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var raw any
			require.NoError(t, json.Unmarshal([]byte(test.wire), &raw))
			field := NewVariantStatementResultField(raw)
			require.Equal(t, Variant, field.GetType())
			require.Equal(t, test.want, field.ToString())
		})
	}
}

func TestVariantToSDKTypeIsRenderedJSON(t *testing.T) {
	var raw any
	require.NoError(t, json.Unmarshal([]byte(`[1,[["a",[11,"x"]]]]`), &raw))
	require.Equal(t, `{"a":"x"}`, NewVariantStatementResultField(raw).ToSDKType())
}

func TestNewResultFieldTypeVariant(t *testing.T) {
	require.Equal(t, Variant, NewResultFieldType("VARIANT"))
}
