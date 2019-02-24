// +build tools

package tools

// This version controls our third-party tools, as per
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
//
// If you don't pin the version, "go get" updates your go.mod/go.sum, creating dirty state
// that causes goreleaser to fail.

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
)
