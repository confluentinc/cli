package cmd

import (
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

type CmfREST struct {
	Client *cmfsdk.APIClient
}
