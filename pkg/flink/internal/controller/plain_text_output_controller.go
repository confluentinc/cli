package controller

import (
	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

func NewPlainTextOutputController(resultFetcher types.ResultFetcherInterface, getWindowWidth func() int) types.OutputControllerInterface {
	return &BaseOutputController{
		resultFetcher: resultFetcher,
		getWindowSize: getWindowWidth,
		outputFormat:  config.OutputFormatPlainText,
	}
}
