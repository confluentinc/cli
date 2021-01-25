package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
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
	err := d.ParseFlagsIntoConfig(cmd)
	return err
}

// Parse "--context" flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (d *DynamicConfig) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	ctxName, err := d.Resolver.ResolveContextFlag(cmd)
	if err != nil {
		return err
	}
	if ctxName != "" {
		_, err := d.FindContext(ctxName)
		if err != nil {
			return err
		}
		d.Config.SetOverwrittenCurrContext(d.Config.CurrentContext)
		d.Config.CurrentContext = ctxName
	}
	//called to initiate DynamicContext so that flags are parsed into context
	ctx, err := d.Context(cmd)
	if err != nil {
		return err
	}
	if ctx == nil {
		return nil
	}
	err = ctx.ParseFlagsIntoContext(cmd)
	return err
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}

//Returns active Context wrapped as a new DynamicContext instance
func (d *DynamicConfig) Context(cmd *cobra.Command) (*DynamicContext, error) {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil, nil
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}
