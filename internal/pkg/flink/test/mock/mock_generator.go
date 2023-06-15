package mock

// Generating mocks using reflect mode: https://github.com/golang/mock#reflect-mode

// controllers
//go:generate mockgen -destination application_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types ApplicationControllerInterface
//go:generate mockgen -destination input_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types InputControllerInterface
//go:generate mockgen -destination table_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types TableControllerInterface
//go:generate mockgen -destination store_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/internal/store StoreInterface
//go:generate mockgen -destination gateway_client_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/ccloudv2 GatewayClientInterface
//go:generate mockgen -destination prompt_mock.go -package=mock github.com/confluentinc/go-prompt IPrompt
//go:generate mockgen -destination console_parser_mock.go -package=mock github.com/confluentinc/go-prompt ConsoleParser
