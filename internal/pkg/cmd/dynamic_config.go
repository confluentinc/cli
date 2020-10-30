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

func (d *DynamicConfig) InitDynamicConfig(cmd *cobra.Command, cfg *v3.Config, resolver FlagResolver) error{
	d.Config = cfg
	d.Resolver = resolver
	err := d.ParseFlagsIntoConfig(cmd)
	return err
}

func (d *DynamicConfig) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	ctxName, err := d.Resolver.ResolveContextFlag(cmd)
	if err != nil {
		return err
	}
	if ctxName != "" {
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
	ctx.ParseFlagsIntoContext(cmd)
	return nil
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}

func (d *DynamicConfig) Context(cmd *cobra.Command) (*DynamicContext, error) {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil, nil
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}
