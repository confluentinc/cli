package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var example = &releaseNotes{
	date:    time.Date(2023, time.April, 25, 0, 0, 0, 0, time.UTC),
	version: "3.10.0",
	sections: map[string][]string{
		"New Features": {"Added `confluent a`"},
		"Bug Fixes":    {"Fixed a bug"},
	},
}

func TestDocs(t *testing.T) {
	expected := []string{
		"[4/25/2023] Confluent CLI v3.10.0 Release Notes",
		"===============================================",
		"",
		"New Features",
		"------------",
		"- Added ``confluent a``",
		"",
		"Bug Fixes",
		"---------",
		"- Fixed a bug",
		"",
		"",
	}
	require.Equal(t, strings.Join(expected, "\n"), docs(example))
}

func TestGithub(t *testing.T) {
	expected := []string{
		"New Features",
		"------------",
		"- Added `confluent a`",
		"",
		"Bug Fixes",
		"---------",
		"- Fixed a bug",
		"",
		"",
	}
	require.Equal(t, strings.Join(expected, "\n"), github(example))
}

func TestS3(t *testing.T) {
	expected := []string{
		"[4/25/2023] Confluent CLI v3.10.0 Release Notes",
		"===============================================",
		"",
		"New Features",
		"------------",
		"- Added `confluent a`",
		"",
		"Bug Fixes",
		"---------",
		"- Fixed a bug",
		"",
		"",
	}
	require.Equal(t, strings.Join(expected, "\n"), s3(example))
}
