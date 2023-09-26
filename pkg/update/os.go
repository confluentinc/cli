package update

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

func GetOs() string {
	if runtime.GOOS == "linux" {
		stderr := new(bytes.Buffer)

		cmd := exec.Command("ldd", "--version")
		cmd.Stderr = stderr
		_ = cmd.Run()

		if strings.Contains(stderr.String(), "musl") {
			return "alpine"
		}
	}

	return runtime.GOOS
}
