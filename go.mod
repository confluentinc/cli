module github.com/confluentinc/cli

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Shopify/sarama v0.0.0-20180730132037-e7238b119b7d
	github.com/Shopify/toxiproxy v2.1.3+incompatible // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da
	github.com/codyaray/go-editor v0.3.0
	github.com/codyaray/go-printer v0.8.0
	github.com/codyaray/retag v0.0.0-20180529164156-4f3c7e6dfbe2 // indirect
	github.com/confluentinc/cc-structs v0.0.0-20181109155559-7cfce9602e5d
	github.com/confluentinc/ccloud-sdk-go v0.0.1
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.1.1
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/hashicorp/go-hclog v0.0.0-20180910232447-e45cbeb79f04
	github.com/hashicorp/go-plugin v0.0.0-20181030172320-54b6ff97d818
	github.com/hashicorp/yamux v0.0.0-20180826203732-cc6d2ea263b2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.0-20180912035003-be2c049b30cc // indirect
	github.com/pierrec/lz4 v0.0.0-20180906185208-bb6bfd13c6a2 // indirect
	github.com/pkg/errors v0.8.0
	github.com/rcrowley/go-metrics v0.0.0-20180503174638-e2704e165165 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.2.0
	github.com/stretchr/testify v1.2.2
	golang.org/x/crypto v0.0.0-20180910181607-0e37d006457b
	golang.org/x/net v0.0.0-20181113165502-88d92db4c548
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e // indirect
	google.golang.org/genproto v0.0.0-20180912233945-5a2fd4cab2d6 // indirect
	google.golang.org/grpc v1.16.0
)

replace (
	github.com/dghubble/sling => github.com/codyaray/sling v0.0.0-20180507231946-0b86fc2ffcc6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20171026124306-e509bb64fe11
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20170925234155-019ae5ada31d
)
