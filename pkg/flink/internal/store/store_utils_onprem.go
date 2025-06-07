package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

func (s *StoreOnPrem) processSetStatement(statement string) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	configKey, configVal, err := parseSetStatement(statement)
	if err != nil {
		return nil, err.(*types.StatementError)
	}
	if configKey == "" {
		return &types.ProcessedStatementOnPrem{
			Kind:             config.OpSet,
			Status:           types.COMPLETED,
			StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
			IsLocalStatement: true,
		}, nil
	}
	if configKey == config.KeyDatabase || configKey == config.KeyCatalog {
		return nil, &types.StatementError{
			Message:    "cannot set a catalog or a database with SET command",
			Suggestion: `please set a catalog with "USE CATALOG catalog-name" and a database with "USE db-name"`,
		}
	}
	if configKey == config.KeyStatementName && strings.TrimSpace(configVal) == "" {
		return nil, &types.StatementError{
			Message:    "cannot set an empty statement name",
			Suggestion: `please provide a non-empty statement name with "SET 'client.statement-name'='non-empty-name'"`,
		}
	}

	if configKey == config.KeyOutputFormat {
		outputFormat := config.OutputFormat(configVal)
		if outputFormat != config.OutputFormatStandard && outputFormat != config.OutputFormatPlainText {
			return nil, &types.StatementError{
				Message:    fmt.Sprintf(`invalid output format for "%s"`, config.KeyOutputFormat),
				Suggestion: fmt.Sprintf(`please provide a valid output format: "%s" or "%s"`, config.OutputFormatStandard, config.OutputFormatPlainText),
			}
		}
	}

	s.Properties.Set(configKey, configVal)
	return &types.ProcessedStatementOnPrem{
		Kind:                 config.OpSet,
		StatusDetail:         "configuration updated successfully",
		Status:               types.COMPLETED,
		StatementResults:     createStatementResults([]string{"Key", "Value"}, [][]string{{configKey, configVal}}),
		IsLocalStatement:     true,
		IsSensitiveStatement: hasSensitiveKey(configKey),
	}, nil
}

func (s *StoreOnPrem) processResetStatement(statement string) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	configKey, err := parseResetStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	if configKey == "" {
		s.Properties.Clear()
		return &types.ProcessedStatementOnPrem{
			Kind:             config.OpReset,
			StatusDetail:     "configuration has been reset successfully",
			Status:           types.COMPLETED,
			StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
			IsLocalStatement: true,
		}, nil
	}
	if !s.Properties.HasKey(configKey) {
		return nil, &types.StatementError{Message: fmt.Sprintf(`configuration key "%s" is not set`, configKey)}
	}
	// if catalog is reset, also reset the database
	if configKey == config.KeyCatalog {
		s.Properties.Delete(config.KeyDatabase)
	}

	s.Properties.Delete(configKey)
	return &types.ProcessedStatementOnPrem{
		Kind:             config.OpReset,
		StatusDetail:     fmt.Sprintf(`configuration key "%s" has been reset successfully`, configKey),
		Status:           types.COMPLETED,
		StatementResults: createStatementResults([]string{"Key", "Value"}, s.Properties.ToSortedSlice(true)),
		IsLocalStatement: true,
	}, nil
}

func (s *StoreOnPrem) processUseStatement(statement string) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	catalog, database, err := parseUseStatement(statement)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	addedConfig := [][]string{}

	// "USE CATALOG catalog_name" statement
	if catalog != "" && database == "" {
		// USE CATALOG <catalog> will remove the current database
		s.Properties.Delete(config.KeyDatabase)

		s.Properties.Set(config.KeyCatalog, catalog)
		addedConfig = append(addedConfig, []string{config.KeyCatalog, catalog})

		// "USE database" statement
	} else if catalog == "" && database != "" {
		// require catalog to be set before running USE <database>
		if !s.Properties.HasKey(config.KeyCatalog) {
			return nil, &types.StatementError{
				Message:    "no catalog was set",
				Suggestion: `please set a catalog first with "USE CATALOG catalog-name" or  before setting a database`,
			}
		}

		s.Properties.Set(config.KeyDatabase, database)
		addedConfig = append(addedConfig, []string{config.KeyDatabase, database})

		// "USE `catalog_name`.`database_name`" statement
	} else if catalog != "" && database != "" {
		s.Properties.Set(config.KeyCatalog, catalog)
		s.Properties.Set(config.KeyDatabase, database)
		addedConfig = append(addedConfig, []string{config.KeyCatalog, catalog})
		addedConfig = append(addedConfig, []string{config.KeyDatabase, database})
	} else {
		return nil, useError()
	}

	return &types.ProcessedStatementOnPrem{
		Kind:             config.OpUse,
		StatusDetail:     "configuration updated successfully",
		Status:           types.COMPLETED,
		StatementResults: createStatementResults([]string{"Key", "Value"}, addedConfig),
		IsLocalStatement: true,
	}, nil
}

func (s *StoreOnPrem) getTimeout() time.Duration {
	if s.Properties.HasKey(config.KeyResultsTimeout) {
		timeoutInMilliseconds, err := strconv.Atoi(s.Properties.Get(config.KeyResultsTimeout))
		if err == nil {
			// TODO - check for error when setting the property so user knows he hasn't set the results-timeout property properly
			return time.Duration(timeoutInMilliseconds) * time.Millisecond
		} else {
			return config.DefaultTimeoutDuration
		}
	} else {
		return config.DefaultTimeoutDuration
	}
}
