module github.com/confluentinc/cli

require (
	github.com/DataDog/zstd v1.3.5 // indirect
	github.com/Shopify/sarama v1.20.1
	github.com/Shopify/toxiproxy v2.1.3+incompatible // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da
	github.com/codyaray/go-editor v0.3.0
	github.com/codyaray/go-printer v0.8.0
	github.com/codyaray/retag v0.0.0-20180529164156-4f3c7e6dfbe2 // indirect
	github.com/confluentinc/cc-structs v0.0.0-20181109155559-7cfce9602e5d
	github.com/confluentinc/ccloud-sdk-go v0.0.5-0.20190205193832-1f6d876cdc3f
	github.com/confluentinc/ccloudapis v0.0.0-20190205225915-829aeb3a907d
	github.com/confluentinc/proto-go-setter v0.0.0-20180912191759-fb17e76fc076 // indirect
	github.com/confluentinc/protoc-gen-ccloud v0.0.1 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful v2.8.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/spec v0.18.0 // indirect
	github.com/gogo/protobuf v1.2.0
	github.com/golang/mock v1.2.0 // indirect
	github.com/golang/protobuf v1.2.1-0.20181128192352-1d3f30b51784
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/golangci/errcheck v0.0.0-20181003203344-ef45e06d44b6 // indirect
	github.com/golangci/golangci-lint v1.12.2
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/google/uuid v1.1.0
	github.com/goreleaser/goreleaser v0.101.0
	github.com/hashicorp/go-hclog v0.0.0-20180910232447-e45cbeb79f04
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-plugin v0.0.0-20181030172320-54b6ff97d818
	github.com/hashicorp/yamux v0.0.0-20180826203732-cc6d2ea263b2 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.0-20180912035003-be2c049b30cc // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/pascaldekloe/goe v0.0.0-20180627143212-57f6aae5913c // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/shurcooL/go v0.0.0-20190121191506-3fef8c783dec // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.2
	github.com/spf13/viper v1.2.0
	github.com/stretchr/testify v1.3.0
	github.com/travisjeffery/proto-go-sql v0.0.0-20180327175836-681792161a58 // indirect
	github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43 // indirect
	golang.org/x/crypto v0.0.0-20190131182504-b8fe1690c613
	golang.org/x/net v0.0.0-20190125091013-d26f9f9a57f3
	golang.org/x/oauth2 v0.0.0-20190130055435-99b60b757ec1 // indirect
	golang.org/x/sys v0.0.0-20190204203706-41f3e6584952 // indirect
	golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2 // indirect
	golang.org/x/tools v0.0.0-20190205201329-379209517ffe // indirect
	google.golang.org/genproto v0.0.0-20190201180003-4b09977fb922 // indirect
	google.golang.org/grpc v1.18.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20181221193117-173ce66c1e39 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190103235604-e7617803aceb // indirect
	k8s.io/apimachinery v0.0.0-20190109170643-c3a4c8673eae // indirect
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be // indirect
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed // indirect
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	mvdan.cc/unparam v0.0.0-20190124213536-fbb59629db34 // indirect
	sourcegraph.com/sourcegraph/go-diff v0.5.0 // indirect
	sourcegraph.com/sqs/pbtypes v1.0.0 // indirect
)

replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20171026124306-e509bb64fe11
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20170925234155-019ae5ada31d
)
