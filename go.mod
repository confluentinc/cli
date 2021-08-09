module github.com/confluentinc/cli

require (
	code.gitea.io/sdk/gitea v0.14.0 // indirect
	contrib.go.opencensus.io/resource v0.0.0-20190131005048-21591786a5e0 // indirect
	github.com/Azure/azure-amqp-common-go v1.1.4 // indirect
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/Shopify/sarama v1.23.1
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/antihax/optional v1.0.0
	github.com/apex/log v1.9.0 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da
	github.com/aws/aws-sdk-go v1.38.35
	github.com/blang/semver v3.5.1+incompatible
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/c-bata/go-prompt v0.2.3
	github.com/caarlos0/go-shellwords v1.0.12 // indirect
	github.com/chromedp/chromedp v0.5.2
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/codyaray/retag v0.0.0-20180529164156-4f3c7e6dfbe2 // indirect
	github.com/confluentinc/bincover v0.2.0
	github.com/confluentinc/cc-structs/kafka/billing v0.753.0
	github.com/confluentinc/cc-structs/kafka/clusterlink v0.753.0
	github.com/confluentinc/cc-structs/kafka/core v0.753.0
	github.com/confluentinc/cc-structs/kafka/flow v0.812.0
	github.com/confluentinc/cc-structs/kafka/org v0.812.0
	github.com/confluentinc/cc-structs/kafka/product/core v0.753.0
	github.com/confluentinc/cc-structs/kafka/scheduler v0.812.0
	github.com/confluentinc/cc-structs/kafka/util v0.753.0
	github.com/confluentinc/cc-structs/operator v0.753.0
	github.com/confluentinc/ccloud-sdk-go-v1 v0.0.84
	github.com/confluentinc/go-editor v0.4.0
	github.com/confluentinc/go-netrc v0.0.0-20201015001751-d8d220f17928
	github.com/confluentinc/go-printer v0.13.0
	github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3 v0.3.5
	github.com/confluentinc/mds-sdk-go/mdsv1 v0.0.27
	github.com/confluentinc/mds-sdk-go/mdsv2alpha1 v0.0.27
	github.com/confluentinc/properties v0.0.0-20190814194548-42c10394a787
	github.com/confluentinc/schema-registry-sdk-go v0.0.9
	github.com/davecgh/go-spew v1.1.1
	github.com/dghubble/sling v1.3.0
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/fatih/color v1.12.0
	github.com/gliderlabs/ssh v0.3.0 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/flect v0.1.3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.5.2
	github.com/golangci/golangci-lint v1.30.0
	github.com/golangci/revgrep v0.0.0-20180812185044-276a5c0a1039 // indirect
	github.com/google/go-cmp v0.5.5
	github.com/google/go-github/v25 v25.0.2
	github.com/google/go-github/v28 v28.1.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.2.0
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/goreleaser/goreleaser v0.162.1
	github.com/goreleaser/nfpm v0.12.0 // indirect
	github.com/goreleaser/nfpm/v2 v2.4.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-version v1.2.1
	github.com/havoc-io/gopass v0.0.0-20170602182606-9a121bec1ae7
	github.com/iancoleman/strcase v0.1.2
	github.com/imdario/mergo v0.3.12
	github.com/jarcoal/httpmock v1.0.8 // indirect
	github.com/jhump/protoreflect v1.7.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/kamilsk/retry/v4 v4.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.9.8
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/mattn/go-zglob v0.0.0-20180803001819-2ea3427bfa53 // indirect
	github.com/mitchellh/golicense v0.2.0
	github.com/olekukonko/tablewriter v0.0.0-20180912035003-be2c049b30cc
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/segmentio/analytics-go v3.1.0+incompatible
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/tidwall/gjson v1.6.5
	github.com/tidwall/pretty v1.0.2
	github.com/tidwall/sjson v1.0.4
	github.com/travisjeffery/mocker v1.1.0
	github.com/uber-go/atomic v1.3.2 // indirect
	github.com/uber/jaeger-client-go v2.15.0+incompatible // indirect
	github.com/uber/jaeger-lib v1.5.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xanzy/go-gitlab v0.48.0 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.mongodb.org/mongo-driver v1.0.1 // indirect
	gocloud.dev v0.22.0 // indirect
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	golang.org/x/tools v0.0.0-20210105210202-9ed45478a130 // indirect
	gonum.org/v1/netlib v0.0.0-20200317120129-c5a04cffd98a // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1
	k8s.io/kube-openapi v0.0.0-20200427153329-656914f816f9 // indirect
	mvdan.cc/sh/v3 v3.2.2
)

replace (
	github.com/shurcooL/sanitized_anchor_name => github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/spf13/cobra => github.com/spf13/cobra v1.1.3-0.20210218152603-eb3b6397b1b5
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20190316192920-e2bddce071ad
	k8s.io/api => k8s.io/api v0.0.0-20190126160459-e86510ea3fe7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20171026124306-e509bb64fe11
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20170925234155-019ae5ada31d
)

go 1.16
