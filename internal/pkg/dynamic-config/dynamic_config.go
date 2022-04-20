package dynamicconfig

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type DynamicConfig struct {
	*v1.Config
	Client   *ccloud.Client
	V2Client *ccloudv2.Client
}

func NewDynamicConfig(config *v1.Config, client *ccloud.Client, v2Client *ccloudv2.Client) *DynamicConfig {
	return &DynamicConfig{
		Config:   config,
		Client:   client,
		V2Client: v2Client,
	}
}

// Set DynamicConfig values for command with config and resolver from prerunner
// Calls ParseFlagsIntoConfig so that state flags are parsed ino config struct
func (d *DynamicConfig) InitDynamicConfig(cmd *cobra.Command, cfg *v1.Config) error {
	d.Config = cfg
	return d.ParseFlagsIntoConfig(cmd)
}

// Parse "--context" flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (d *DynamicConfig) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	if context, _ := cmd.Flags().GetString("context"); context != "" {
		if _, err := d.FindContext(context); err != nil {
			return err
		}
		d.Config.SetOverwrittenCurrContext(d.Config.CurrentContext)
		d.Config.CurrentContext = context
	}

	return nil
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return New(ctx, d.Client, d.V2Client), nil
}

// Context returns the active context as a DynamicContext object.
func (d *DynamicConfig) Context() *DynamicContext {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil
	}
	return New(ctx, d.Client, d.V2Client)
}
