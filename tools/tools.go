// +build tools
package tools 

// This version controls our third-party tools, as per https://github.com/golang/go/issues/25922
//
// If you try doing this with "go get" in Makefile, it updates your go.mod/go.sum, creating dirty
// state that causes goreleaser to fail.

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
)
