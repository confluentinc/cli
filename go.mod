module github.com/confluentinc/cli

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/antihax/optional v1.0.0
	github.com/armon/go-metrics v0.3.10
	github.com/aws/aws-sdk-go v1.42.9
	github.com/c-bata/go-prompt v0.2.6
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/confluentinc/bincover v0.2.0
	github.com/confluentinc/cc-structs/kafka/billing v0.1033.0
	github.com/confluentinc/cc-structs/kafka/clusterlink v0.1033.0
	github.com/confluentinc/cc-structs/kafka/core v0.1033.0
	github.com/confluentinc/cc-structs/kafka/flow v0.1033.0
	github.com/confluentinc/cc-structs/kafka/org v0.1033.0
	github.com/confluentinc/cc-structs/kafka/product/core v0.1033.0
	github.com/confluentinc/cc-structs/kafka/scheduler v0.1033.0
	github.com/confluentinc/cc-structs/kafka/util v0.1033.0
	github.com/confluentinc/cc-structs/operator v0.1033.0
	github.com/confluentinc/ccloud-sdk-go-v1 v0.0.93
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/confluentinc/countrycode v0.0.0-20211121160605-23262b771ab0
	github.com/confluentinc/go-editor v0.9.0
	github.com/confluentinc/go-netrc v0.0.0-20211121160620-ec37f663ea18
	github.com/confluentinc/go-printer v0.16.0
	github.com/confluentinc/go-ps1 v1.0.2
	github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3 v0.3.12
	github.com/confluentinc/mds-sdk-go/mdsv1 v0.0.39
	github.com/confluentinc/mds-sdk-go/mdsv2alpha1 v0.0.39
	github.com/confluentinc/properties v0.0.0-20190814194548-42c10394a787
	github.com/confluentinc/schema-registry-sdk-go v0.0.12
	github.com/davecgh/go-spew v1.1.1
	github.com/dghubble/sling v1.4.0
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/fatih/color v1.13.0
	github.com/gliderlabs/ssh v0.3.0 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/flect v0.2.4
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/golangci/golangci-lint v1.43.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v25 v25.1.3
	github.com/google/uuid v1.3.0
	github.com/goreleaser/goreleaser v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-hclog v1.0.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/havoc-io/gopass v0.0.0-20170602182606-9a121bec1ae7
	github.com/iancoleman/strcase v0.2.0
	github.com/imdario/mergo v0.3.12
	github.com/jhump/protoreflect v1.10.1
	github.com/jonboulle/clockwork v0.2.2
	github.com/linkedin/goavro/v2 v2.10.1
	github.com/mattn/go-isatty v0.0.14
	github.com/mitchellh/golicense v0.2.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/segmentio/analytics-go v3.1.0+incompatible
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/tidwall/gjson v1.11.0
	github.com/tidwall/pretty v1.2.0
	github.com/tidwall/sjson v1.2.3
	github.com/travisjeffery/mocker v1.1.1
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gonum.org/v1/netlib v0.0.0-20200317120129-c5a04cffd98a // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
	mvdan.cc/sh/v3 v3.4.0
)

replace (
	github.com/influxdata/influxdb1-client => github.com/influxdata/influxdb1-client v0.0.0-20190124185755-16c852ea613f
	github.com/shurcooL/sanitized_anchor_name => github.com/shurcooL/sanitized_anchor_name v1.0.0
	k8s.io/api => k8s.io/api v0.0.0-20190126160459-e86510ea3fe7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20171026124306-e509bb64fe11
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20170925234155-019ae5ada31d
)

go 1.16
