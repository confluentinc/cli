package mock

// Generating mocks using reflect mode: https://github.com/golang/mock#reflect-mode

// controllers
//go:generate mockgen -destination application_controller_mock.go -package=mock github.com/confluentinc/flink-sql-client/pkg/controller ApplicationControllerInterface
//go:generate mockgen -destination input_controller_mock.go -package=mock github.com/confluentinc/flink-sql-client/pkg/controller InputControllerInterface
//go:generate mockgen -destination table_controller_mock.go -package=mock github.com/confluentinc/flink-sql-client/pkg/controller TableControllerInterface
//go:generate mockgen -destination store_mock.go -package=mock github.com/confluentinc/flink-sql-client/pkg/controller StoreInterface
//go:generate mockgen -destination gateway_client_mock.go -package=mock github.com/confluentinc/flink-sql-client/pkg/controller GatewayClientInterface
