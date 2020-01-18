package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
)

type DynamicConfig struct {
	*config.Config
	Resolver FlagResolver
	Client   *ccloud.Client
}

func NewDynamicConfig(config *config.Config, resolver FlagResolver, client *ccloud.Client) *DynamicConfig {
	return &DynamicConfig{
		Config:   config,
		Resolver: resolver,
		Client:   client,
	}
}

func (d *DynamicConfig) FindContext(name string) (*DynamicContext, error) {
	ctx, err := d.Config.FindContext(name)
	if err != nil {
		return nil, err
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}

func (d *DynamicConfig) Context(cmd *cobra.Command) (*DynamicContext, error) {
	ctxName, err := d.Resolver.ResolveContextFlag(cmd)
	if err != nil {
		return nil, err
	}
	if ctxName != "" {
		return d.FindContext(ctxName)
	}
	ctx := d.Config.Context()
	if ctx == nil {
		return nil, nil
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client), nil
}
