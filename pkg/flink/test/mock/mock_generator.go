package mock

// Generating mocks using reflect mode: https://github.com/uber-go/mock

// controllers
//go:generate go run go.uber.org/mock/mockgen -destination application_controller_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types ApplicationControllerInterface
//go:generate go run go.uber.org/mock/mockgen -destination input_controller_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types InputControllerInterface
//go:generate go run go.uber.org/mock/mockgen -destination result_fetcher_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types ResultFetcherInterface
//go:generate go run go.uber.org/mock/mockgen -destination statement_controller_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types StatementControllerInterface
//go:generate go run go.uber.org/mock/mockgen -destination output_controller_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types OutputControllerInterface
//go:generate go run go.uber.org/mock/mockgen -destination store_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types StoreInterface
//go:generate go run go.uber.org/mock/mockgen -destination reverse_i_search_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/internal/reverseisearch ReverseISearch
//go:generate go run go.uber.org/mock/mockgen -destination gateway_client_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/ccloudv2 GatewayClientInterface
//go:generate go run go.uber.org/mock/mockgen -destination table_view_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/components TableViewInterface
//go:generate go run go.uber.org/mock/mockgen -destination prompt_mock.go -package=mock github.com/confluentinc/go-prompt IPrompt
//go:generate go run go.uber.org/mock/mockgen -destination console_parser_mock.go -package=mock github.com/confluentinc/go-prompt ConsoleParser
//go:generate go run go.uber.org/mock/mockgen -destination json_rpc2_conn.go -package=mock github.com/confluentinc/cli/v4/pkg/flink/types JSONRpcConn
//go:generate go run go.uber.org/mock/mockgen -destination cmf_client_mock.go -package=mock github.com/confluentinc/cli/v4/pkg/flink CmfClientInterface
