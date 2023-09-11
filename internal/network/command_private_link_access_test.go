package network

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
)

func TesGetCloudWithAwsPrivateLinkAccess(t *testing.T) {
	cmd := new(cobra.Command)
	cfg := config.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &privateLinkAccessCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.Config = dynamicconfig.New(cfg, nil)

	cloud, err := c.getCloud(networkingv1.NetworkingV1PrivateLinkAccess{
		Id: networkingv1.PtrString("pla-123456"),
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Cloud: &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
				NetworkingV1AwsPrivateLinkAccess: &networkingv1.NetworkingV1AwsPrivateLinkAccess{
					Kind:    "AwsPrivateLinkAccess",
					Account: "012345678901",
				},
			},
			DisplayName: networkingv1.PtrString("aws-pla"),
			Environment: &networkingv1.ObjectReference{Id: "env-00000"},
			Network:     &networkingv1.ObjectReference{Id: "n-abcde1"},
		},
		Status: &networkingv1.NetworkingV1PrivateLinkAccessStatus{
			Phase: "READY",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, CloudAws, cloud)
}
