package serdes

var DekAlgorithms = []string{
	"AES128_GCM",
	"AES256_GCM",
	"AES256_SIV",
}

var KmsTypes = []string{
	"aws-kms",
	"azure-kms",
	"gcp-kms",
}

var Formats = []string{
	"string",
	"avro",
	"double",
	"integer",
	"jsonschema",
	"protobuf",
}
