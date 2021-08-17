package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

type DynamicConfig struct {
	*v3.Config
	Resolver FlagResolver
	Client   *ccloud.Client
}

func NewDynamicConfig(config *v3.Config, resolver FlagResolver, client *ccloud.Client) *DynamicConfig {
	return &DynamicConfig{
		Config:   config,
		Resolver: resolver,
		Client:   client,
	}
}

// Set DynamicConfig values for command with config and resolver from prerunner
// Calls ParseFlagsIntoConfig so that state flags are parsed ino config struct
func (d *DynamicConfig) InitDynamicConfig(cmd *cobra.Command, cfg *v3.Config, resolver FlagResolver) error {
	d.Config = cfg
	d.Resolver = resolver
	return d.ParseFlagsIntoConfig(cmd)
}

// Parse "--context" flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (d *DynamicConfig) ParseFlagsIntoConfig(cmd *cobra.Command) error { //version *version.Version) error {
	ctxName, err := d.Resolver.ResolveContextFlag(cmd)
	if err != nil {
		return err
	}

	if ctxName != "" {
		if _, err := d.FindContext(ctxName); err != nil {
			return err
		}
		d.Config.SetOverwrittenCurrContext(d.Config.CurrentContext)
		d.Config.CurrentContext = ctxName
	}

	return nil
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}

// Context returns the active context as a DynamicContext object.
func (d *DynamicConfig) Context() *DynamicContext {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client)
}
