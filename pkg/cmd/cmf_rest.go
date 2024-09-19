package cmd

import (
	cmfsdk "github.com/confluentinc/cmf-sdk-go"
)

type CmfREST struct {
	Client *cmfsdk.APIClient
}
