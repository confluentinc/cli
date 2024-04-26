package types

import "github.com/confluentinc/cli/v3/pkg/flink/config"

type UserPropertiesInterface interface {
	Clear()
	Delete(key string)
	Get(key string) string
	GetNonLocalProperties() map[string]string
	GetOrDefault(key string, defaultValue string) string
	GetOutputFormat() config.OutputFormat
	GetProperties() map[string]string
	HasKey(key string) bool
	Set(key string, value string)
	ToSortedSlice(annotateDefaultValues bool) [][]string
}
