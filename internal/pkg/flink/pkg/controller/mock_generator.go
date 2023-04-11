package controller

// Generating mocks using reflect mode: https://github.com/golang/mock#reflect-mode

// controllers
//go:generate mockgen -destination application_controller_mock.go -package=controller github.com/confluentinc/flink-sql-client/pkg/controller ApplicationControllerInterface
//go:generate mockgen -destination input_controller_mock.go -package=controller github.com/confluentinc/flink-sql-client/pkg/controller InputControllerInterface
//go:generate mockgen -destination table_controller_mock.go -package=controller github.com/confluentinc/flink-sql-client/pkg/controller TableControllerInterface
