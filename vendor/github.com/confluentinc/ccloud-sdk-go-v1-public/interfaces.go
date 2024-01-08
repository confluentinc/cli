//go:generate mocker --package mock --prefix "" --destination mock/account.go interfaces.go AccountInterface
//go:generate mocker --package mock --prefix "" --destination mock/auth.go interfaces.go Auth
//go:generate mocker --package mock --prefix "" --destination mock/billing.go interfaces.go Billing
//go:generate mocker --package mock --prefix "" --destination mock/environment_metadata.go interfaces.go EnvironmentMetadata
//go:generate mocker --package mock --prefix "" --destination mock/growth.go interfaces.go Growth
//go:generate mocker --package mock --prefix "" --destination mock/schema_registry.go interfaces.go SchemaRegistry
//go:generate mocker --package mock --prefix "" --destination mock/user.go interfaces.go UserInterface

package ccloud

// Account service allows managing accounts in Confluent Cloud
type AccountInterface interface {
	Create(*Account) (*Account, error)
	Get(*Account) (*Account, error)
	List(*Account) ([]*Account, error)
}

// Auth allows authenticating in Confluent Cloud
type Auth interface {
	Login(*AuthenticateRequest) (*AuthenticateReply, error)
	OktaLogin(*AuthenticateRequest) (*AuthenticateReply, error)
	User() (*GetMeReply, error)
}

// Billing service allows getting billing information for an org in Confluent Cloud
type Billing interface {
	GetPriceTable(*Organization, string) (*PriceTable, error)
	GetPaymentInfo(*Organization) (*Card, error)
	UpdatePaymentInfo(*Organization, string) error
	ClaimPromoCode(*Organization, string) (*PromoCodeClaim, error)
	GetClaimedPromoCodes(*Organization, bool) ([]*PromoCodeClaim, error)
}

// Environment metadata service allows getting information about available cloud regions data
type EnvironmentMetadata interface {
	Get() ([]*CloudMetadata, error)
}

// External Identity services allow managing external identities for Bring-Your-Own-Key in Confluent Cloud.
type ExternalIdentity interface {
	CreateExternalIdentity(string, string) (string, error)
}

type Growth interface {
	GetFreeTrialInfo(int32) ([]*GrowthPromoCodeClaim, error)
}

// Schema Registry service allows managing SR clusters in Confluent Cloud
type SchemaRegistry interface {
	CreateSchemaRegistryCluster(*SchemaRegistryClusterConfig) (*SchemaRegistryCluster, error)
	GetSchemaRegistryClusters(*SchemaRegistryCluster) ([]*SchemaRegistryCluster, error)
	GetSchemaRegistryCluster(*SchemaRegistryCluster) (*SchemaRegistryCluster, error)
	UpdateSchemaRegistryCluster(*SchemaRegistryCluster) (*SchemaRegistryCluster, error)
	DeleteSchemaRegistryCluster(*SchemaRegistryCluster) error
}

// Signup service allows managing signups in Confluent Cloud
type Signup interface {
	Create(*SignupRequest) (*SignupReply, error)
	SendVerificationEmail(*User) error
}

// User service allows managing users in Confluent Cloud
type UserInterface interface {
	List() ([]*User, error)
	GetServiceAccounts() ([]*User, error)
	GetServiceAccount(int32) (*User, error)
	LoginRealm(*GetLoginRealmRequest) (*GetLoginRealmReply, error)
}

// Logger provides an interface that will be used for all logging in this client. User provided
// logging implementations must conform to this interface. Popular loggers like zap and logrus
// already implement this interface.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}
