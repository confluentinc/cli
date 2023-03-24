module github.com/confluentinc/cli

go 1.20

require (
	github.com/antihax/optional v1.0.0
	github.com/aws/aws-sdk-go v1.44.224
	github.com/billgraziano/dpapi v0.4.0
	github.com/brianstrauch/cobra-shell v0.4.0
	github.com/chromedp/chromedp v0.9.1
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/confluentinc/ccloud-sdk-go-v1-public v0.0.0-20230117212759-138da0e5aa56
	github.com/confluentinc/ccloud-sdk-go-v2/apikeys v0.4.0
	github.com/confluentinc/ccloud-sdk-go-v2/byok v0.0.1
	github.com/confluentinc/ccloud-sdk-go-v2/cdx v0.0.5
	github.com/confluentinc/ccloud-sdk-go-v2/cli v0.1.0
	github.com/confluentinc/ccloud-sdk-go-v2/cmk v0.7.0
	github.com/confluentinc/ccloud-sdk-go-v2/connect v0.3.0
	github.com/confluentinc/ccloud-sdk-go-v2/iam v0.10.0
	github.com/confluentinc/ccloud-sdk-go-v2/identity-provider v0.2.0
	github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas v0.4.0
	github.com/confluentinc/ccloud-sdk-go-v2/kafkarest v0.12.0
	github.com/confluentinc/ccloud-sdk-go-v2/ksql v0.2.0
	github.com/confluentinc/ccloud-sdk-go-v2/mds v0.4.0
	github.com/confluentinc/ccloud-sdk-go-v2/metrics v0.2.0
	github.com/confluentinc/ccloud-sdk-go-v2/org v0.6.0
	github.com/confluentinc/ccloud-sdk-go-v2/service-quota v0.2.0
	github.com/confluentinc/ccloud-sdk-go-v2/stream-designer v0.2.0
	github.com/confluentinc/confluent-kafka-go v1.9.3-RC3
	github.com/confluentinc/go-editor v0.11.0
	github.com/confluentinc/go-netrc v0.0.0-20220321173724-4d50f36ff450
	github.com/confluentinc/go-ps1 v1.0.2
	github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3 v0.3.14
	github.com/confluentinc/mds-sdk-go-public/mdsv1 v0.0.0-20230117192233-7e6d894d74a9
	github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1 v0.0.0-20230117192233-7e6d894d74a9
	github.com/confluentinc/properties v0.0.0-20190814194548-42c10394a787
	github.com/confluentinc/schema-registry-sdk-go v0.0.19
	github.com/davecgh/go-spew v1.1.1
	github.com/dghubble/sling v1.4.1
	github.com/fatih/color v1.15.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/flect v1.0.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.3
	github.com/google/go-github/v50 v50.2.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-hclog v1.4.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp v0.7.2
	github.com/hashicorp/go-version v1.6.0
	github.com/havoc-io/gopass v0.0.0-20170602182606-9a121bec1ae7
	github.com/iancoleman/strcase v0.2.0
	github.com/imdario/mergo v0.3.14
	github.com/jhump/protoreflect v1.15.1
	github.com/jonboulle/clockwork v0.3.0
	github.com/keybase/go-keychain v0.0.0-20221221221913-9be78f6c498b
	github.com/linkedin/goavro/v2 v2.12.0
	github.com/mattn/go-isatty v0.0.17
	github.com/olekukonko/tablewriter v0.0.5
	github.com/panta/machineid v1.0.2
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/sevlyar/retag v0.0.0-20190429052747-c3f10e304082
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.2
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/swaggest/go-asyncapi v0.8.0
	github.com/tidwall/gjson v1.14.4
	github.com/tidwall/pretty v1.2.1
	github.com/tidwall/sjson v1.2.5
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.7.0
	golang.org/x/exp v0.0.0-20230307190834-24139beb5833
	golang.org/x/oauth2 v0.6.0
	golang.org/x/term v0.6.0
	golang.org/x/text v0.8.0
	gopkg.in/launchdarkly/go-sdk-common.v2 v2.5.1
	gopkg.in/square/go-jose.v2 v2.6.0
)

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/bufbuild/protocompile v0.4.0 // indirect
	github.com/c-bata/go-prompt v0.2.6 // indirect
	github.com/chromedp/cdproto v0.0.0-20230220211738-2b1ec77315c9 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cloudflare/circl v1.1.0 // indirect
	github.com/confluentinc/proto-go-setter v0.3.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.9.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lyft/protoc-gen-star v0.6.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-tty v0.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/swaggest/jsonschema-go v0.3.45 // indirect
	github.com/swaggest/refl v1.1.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/travisjeffery/proto-go-sql v0.0.0-20190911121832-39ff47280e87 // indirect
	github.com/ugorji/go/codec v1.2.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230119192704-9d59e20e5cd1 // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.2-0.20230222093303-bc1253ad3743 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/launchdarkly/go-jsonstream.v1 v1.0.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.26.1 // indirect
	k8s.io/apimachinery v0.26.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/utils v0.0.0-20230115233650-391b47cb4029 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
