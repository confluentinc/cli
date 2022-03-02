package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type DynamicConfig struct {
	*v1.Config
	Resolver FlagResolver
	Client   *ccloud.Client
	V2Client *V2Client
}

func NewDynamicConfig(config *v1.Config, resolver FlagResolver, client *ccloud.Client, cmkClient *cmkv2.APIClient, orgClient *orgv2.APIClient) *DynamicConfig {
	return &DynamicConfig{
		Config:   config,
		Resolver: resolver,
		Client:   client,
		V2Client: &V2Client{CmkClient: cmkClient, OrgClient: orgClient},
	}
}

// Set DynamicConfig values for command with config and resolver from prerunner
// Calls ParseFlagsIntoConfig so that state flags are parsed ino config struct
func (d *DynamicConfig) InitDynamicConfig(cmd *cobra.Command, cfg *v1.Config, resolver FlagResolver) error {
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
	if d.V2Client == nil {
		return NewDynamicContext(ctx, d.Resolver, d.Client, nil, nil), nil
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client, d.V2Client.CmkClient, d.V2Client.OrgClient), nil
}

// Context returns the active context as a DynamicContext object.
func (d *DynamicConfig) Context() *DynamicContext {
	ctx := d.Config.Context()
	if ctx == nil {
		return nil
	}
	if d.V2Client == nil {
		return NewDynamicContext(ctx, d.Resolver, d.Client, nil, nil)
	}
	return NewDynamicContext(ctx, d.Resolver, d.Client, d.V2Client.CmkClient, d.V2Client.OrgClient)
}
