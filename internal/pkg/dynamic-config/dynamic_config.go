package dynamicconfig

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type DynamicConfig struct {
	*v1.Config
	V2Client *ccloudv2.Client
}

func New(config *v1.Config, v2Client *ccloudv2.Client) *DynamicConfig {
	return &DynamicConfig{
		Config:   config,
		V2Client: v2Client,
	}
}

// Parse "--context" flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (d *DynamicConfig) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	if context, _ := cmd.Flags().GetString("context"); context != "" {
		if _, err := d.FindContext(context); err != nil {
			return err
		}
		d.Config.SetOverwrittenCurrentContext(d.Config.CurrentContext)
		d.Config.CurrentContext = context
	}

	return nil
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return NewDynamicContext(ctx, d.V2Client), nil
}

// Context returns the active context as a DynamicContext object.
func (d *DynamicConfig) Context() *DynamicContext {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil
	}
	return NewDynamicContext(ctx, d.V2Client)
}
