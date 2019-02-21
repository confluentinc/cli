package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// Injected from linker flag like `go build -ldflags "-X 'github.com/confluentinc/cli/version.Version=$(VERSION)' -X ..."
	Version   string = "0.0.0"
	Ref       string
	BuildDate string = "Unknown"
	Host      string
	UserAgent = fmt.Sprintf("Confluent/1.0 ccloud/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
)

func IsReleased() bool {
	return Version != "0.0.0" && !strings.Contains(Version, "dirty") && !strings.Contains(Version, "-g")
}
