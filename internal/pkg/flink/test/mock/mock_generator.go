package mock

// Generating mocks using reflect mode: https://github.com/golang/mock#reflect-mode

// controllers
//go:generate mockgen -destination application_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types ApplicationControllerInterface
//go:generate mockgen -destination input_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types InputControllerInterface
//go:generate mockgen -destination result_fetcher_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types ResultFetcherInterface
//go:generate mockgen -destination statement_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types StatementControllerInterface
//go:generate mockgen -destination output_controller_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types OutputControllerInterface
//go:generate mockgen -destination store_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types StoreInterface
//go:generate mockgen -destination reverse_i_search_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/internal/reverseisearch ReverseISearch
//go:generate mockgen -destination gateway_client_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/ccloudv2 GatewayClientInterface
//go:generate mockgen -destination table_view_mock.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/components TableViewInterface
//go:generate mockgen -destination prompt_mock.go -package=mock github.com/confluentinc/go-prompt IPrompt
//go:generate mockgen -destination console_parser_mock.go -package=mock github.com/confluentinc/go-prompt ConsoleParser
//go:generate mockgen -destination json_rpc2_conn.go -package=mock github.com/confluentinc/cli/internal/pkg/flink/types Conn
