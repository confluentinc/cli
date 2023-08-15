package ccstructs

type Sku int32

const (
	Sku_UNKNOWN          Sku = 0
	Sku_BASIC_LEGACY     Sku = 1
	Sku_BASIC            Sku = 2
	Sku_STANDARD         Sku = 3
	Sku_DEDICATED        Sku = 4
	Sku_DEDICATED_LEGACY Sku = 5
	Sku_ENTERPRISE       Sku = 6
)

var Sku_name = map[int32]string{
	0: "UNKNOWN",
	1: "BASIC_LEGACY",
	2: "BASIC",
	3: "STANDARD",
	4: "DEDICATED",
	5: "DEDICATED_LEGACY",
	6: "ENTERPRISE",
}

var Sku_value = map[string]int32{
	"UNKNOWN":          0,
	"BASIC_LEGACY":     1,
	"BASIC":            2,
	"STANDARD":         3,
	"DEDICATED":        4,
	"DEDICATED_LEGACY": 5,
	"ENTERPRISE":       6,
}
