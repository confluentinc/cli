module github.com/confluentinc/cli

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/DataDog/zstd v1.3.5 // indirect
	github.com/Shopify/sarama v1.20.1
	github.com/Shopify/toxiproxy v2.1.3+incompatible // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da
	github.com/codyaray/go-editor v0.3.0
	github.com/codyaray/go-printer v0.8.0
	github.com/codyaray/retag v0.0.0-20180529164156-4f3c7e6dfbe2 // indirect
	github.com/confluentinc/cc-structs v0.0.0-20181109155559-7cfce9602e5d
	github.com/confluentinc/ccloud-sdk-go v0.0.5-0.20190111135442-0c88bb94860a
	github.com/confluentinc/ccloudapis v0.0.0-20190115160518-100a9b43e5be
	github.com/confluentinc/proto-go-setter v0.0.0-20180912191759-fb17e76fc076 // indirect
	github.com/dghubble/sling v1.1.0
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful v2.8.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/spec v0.18.0 // indirect
	github.com/gogo/protobuf v1.2.0
	github.com/golang/protobuf v1.2.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/google/uuid v1.1.0
	github.com/hashicorp/go-hclog v0.0.0-20180910232447-e45cbeb79f04
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-plugin v0.0.0-20181030172320-54b6ff97d818
	github.com/hashicorp/yamux v0.0.0-20180826203732-cc6d2ea263b2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.0-20180912035003-be2c049b30cc // indirect
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/pierrec/lz4 v0.0.0-20180906185208-bb6bfd13c6a2 // indirect
	github.com/pkg/errors v0.8.0
	github.com/rcrowley/go-metrics v0.0.0-20180503174638-e2704e165165 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.2
	github.com/spf13/viper v1.2.0
	github.com/stretchr/testify v1.2.2
	github.com/travisjeffery/proto-go-sql v0.0.0-20180327175836-681792161a58 // indirect
	golang.org/x/crypto v0.0.0-20180910181607-0e37d006457b
	golang.org/x/net v0.0.0-20181113165502-88d92db4c548
	google.golang.org/genproto v0.0.0-20180912233945-5a2fd4cab2d6 // indirect
	google.golang.org/grpc v1.16.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20181221193117-173ce66c1e39 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190103235604-e7617803aceb // indirect
	k8s.io/apimachinery v0.0.0-20190109170643-c3a4c8673eae // indirect
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be // indirect
)

replace (
	github.com/dghubble/sling => github.com/codyaray/sling v0.0.0-20180507231946-0b86fc2ffcc6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20171026124306-e509bb64fe11
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20170925234155-019ae5ada31d
)
