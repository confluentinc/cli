package panic_recovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStack(t *testing.T) {
	rawTrace := `runtime/debug.Stack()
				~/.goenv/versions/1.20.4/src/runtime/debug/stack.go:24 +0x65
				github.com/confluentinc/cli/internal/cmd.Execute.func1()
				~/cli/internal/cmd/command.go:157 +0x198
				panic-recovery({0x1019356e0, 0xc0004ba990})
				~/.goenv/versions/1.20.4/src/runtime/panic-recovery.go:884 +0x213
				github.com/confluentinc/cli/internal/cmd.Execute(0xc000004600, {0xc000040050?, 0x101d1b4c0?, 0x28?}, 0xc0002c62c0)
				~/cli/internal/cmd/command.go:172 +0x17c
				main.main()
				~/cli/cmd/confluent/main.go:40 +0x3f4`
	formattedTrace := []string{
		"~/.goenv/versions/1.20.4/src/runtime/panic-recovery.go:884 +0x213",
		"~/cli/internal/cmd/command.go:172 +0x17c",
		"~/cli/cmd/confluent/main.go:40 +0x3f4",
	}
	assert.Equal(t, formattedTrace, parseStack(rawTrace))

	rawTrace = `github.com/confluentinc/cli/internal/pkg/dynamic-config.(*DynamicContext).AuthenticatedEnvId(0xc0001254c0)
				/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/dynamic_context.go:241 +0x64
				github.com/confluentinc/cli/internal/pkg/dynamic-config.(*DynamicContext).FetchCluster(0xc0001254c0, {0x7ffccb841541, 0xa})
				/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/client.go:17 +0x45
				github.com/confluentinc/cli/internal/pkg/dynamic-config.(*DynamicContext).FindKafkaCluster(0xc0001254c0, {0x7ffccb841541, 0xa})
				/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/dynamic_context.go:140 +0x167
				github.com/confluentinc/cli/internal/cmd/api-key.(*command).resolveResourceId(0xc0001c3b30, 0x0?, 0xc0001a5180) 
				/go/src/github.com/confluentinc/cli/internal/cmd/api-key/command.go:136 +0x269
				github.com/confluentinc/cli/internal/cmd/api-key.(*command).create(0xc0001c3b30, 0x20?, {0xc000100000?, 0xc000125740?, 0x0?}) 
				/go/src/github.com/confluentinc/cli/internal/cmd/api-key/command_create.go:67 +0xb6 
				github.com/confluentinc/cli/internal/pkg/cmd.Chain.func1(0xc00011a0c0?, {0xc0004a0080, 0x0, 0x8}) 
				/go/src/github.com/confluentinc/cli/internal/pkg/cmd/cobra.go:24 +0x83 
				github.com/confluentinc/cli/internal/pkg/cmd.CatchErrors.func1(0xc0001ef800, {0xc0004a0080, 0x0, 0x8})
				/go/src/github.com/confluentinc/cli/internal/pkg/cmd/cobra.go:12 +0x69
				github.com/spf13/cobra.(*Command).execute(0xc0001ef800, {0xc0004a0000, 0x8, 0x8})
				/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:916 +0x862
				github.com/spf13/cobra.(*Command).ExecuteC(0xc000479b00)
				/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:1040 +0x3b4
				github.com/spf13/cobra.(*Command).Execute(...)
				/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:968
				github.com/confluentinc/cli/internal/cmd.Execute(0xc000479b00, {0xc00004e0c0?, 0x2c28e58?, 0x8?}, 0xc0004d1ef0)
				/go/src/github.com/confluentinc/cli/internal/cmd/command.go:150 +0x1e5
				main.main()
				/go/src/github.com/confluentinc/cli/cmd/confluent/main.go:36 +0x26f`
	formattedTrace = []string{
		"/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/dynamic_context.go:241 +0x64",
		"/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/client.go:17 +0x45",
		"/go/src/github.com/confluentinc/cli/internal/pkg/dynamic-config/dynamic_context.go:140 +0x167",
		"/go/src/github.com/confluentinc/cli/internal/cmd/api-key/command.go:136 +0x269",
		"/go/src/github.com/confluentinc/cli/internal/cmd/api-key/command_create.go:67 +0xb6",
		"/go/src/github.com/confluentinc/cli/internal/pkg/cmd/cobra.go:24 +0x83",
		"/go/src/github.com/confluentinc/cli/internal/pkg/cmd/cobra.go:12 +0x69",
		"/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:916 +0x862",
		"/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:1040 +0x3b4",
		"/go/src/github.com/confluentinc/cli/vendor/github.com/spf13/cobra/command.go:968",
		"/go/src/github.com/confluentinc/cli/internal/cmd/command.go:150 +0x1e5",
		"/go/src/github.com/confluentinc/cli/cmd/confluent/main.go:36 +0x26f",
	}
	assert.Equal(t, formattedTrace, parseStack(rawTrace))
}
