package utils

import (
	"os"
	"strings"
)

var wslKernelReleaseFiles = []string{"/proc/sys/kernel/osrelease", "/proc/version"}

// IsWSL reports whether the current process is running under Windows Subsystem for Linux.
func IsWSL() bool {
	for _, path := range wslKernelReleaseFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if isWSLKernelRelease(string(data)) {
			return true
		}
	}
	return false
}

func isWSLKernelRelease(release string) bool {
	return strings.Contains(strings.ToLower(release), "microsoft")
}
