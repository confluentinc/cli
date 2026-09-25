package endpoint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEndpointFlag(t *testing.T) {
	const (
		networkCrn     = "crn://confluent.cloud/organization=org-123/environment=env-123456/network=n-111111"
		accessPointCrn = "crn://confluent.cloud/organization=org-123/environment=env-123456/gateway=platt-111111/access-point=plattc-111111"
	)

	t.Run("private with network", func(t *testing.T) {
		config, err := parseEndpointFlag("name=west,type=private,network=" + networkCrn)
		require.NoError(t, err)
		require.Equal(t, "west", config.Name)
		require.Equal(t, "private", config.EndpointFilter.Type)
		require.Equal(t, networkCrn, config.EndpointFilter.GetNetworkCrn())
		require.Nil(t, config.EndpointFilter.AccessPointCrn)
	})

	t.Run("private with access-point", func(t *testing.T) {
		config, err := parseEndpointFlag("name=west,type=private,access-point=" + accessPointCrn)
		require.NoError(t, err)
		require.Equal(t, accessPointCrn, config.EndpointFilter.GetAccessPointCrn())
		require.Nil(t, config.EndpointFilter.NetworkCrn)
	})

	t.Run("public", func(t *testing.T) {
		config, err := parseEndpointFlag("name=west,type=public")
		require.NoError(t, err)
		require.Equal(t, "public", config.EndpointFilter.Type)
	})

	// The CLI validates only what it needs to build the request: the name=<name>,type=<type>
	// shape. Which filter fields a private or public endpoint may carry is the Switchover API's
	// decision, so those combinations are forwarded as given and rejected (or not) server-side.
	t.Run("filter combinations are passed through, not validated client-side", func(t *testing.T) {
		for name, tc := range map[string]struct {
			raw                       string
			typ, network, accessPoint string
		}{
			"unknown type":                     {"name=west,type=internal", "internal", "", ""},
			"private without network or ap":    {"name=west,type=private", "private", "", ""},
			"private with both network and ap": {"name=west,type=private,network=" + networkCrn + ",access-point=" + accessPointCrn, "private", networkCrn, accessPointCrn},
			"public with network":              {"name=west,type=public,network=" + networkCrn, "public", networkCrn, ""},
			"public with access-point":         {"name=west,type=public,access-point=" + accessPointCrn, "public", "", accessPointCrn},
		} {
			t.Run(name, func(t *testing.T) {
				config, err := parseEndpointFlag(tc.raw)
				require.NoError(t, err)
				require.Equal(t, tc.typ, config.EndpointFilter.Type)
				require.Equal(t, tc.network, config.EndpointFilter.GetNetworkCrn())
				require.Equal(t, tc.accessPoint, config.EndpointFilter.GetAccessPointCrn())
			})
		}
	})

	t.Run("malformed values are rejected", func(t *testing.T) {
		for name, raw := range map[string]string{
			"missing name":  "type=private,network=" + networkCrn,
			"missing type":  "name=west,network=" + networkCrn,
			"unknown key":   "name=west,type=public,foo=bar",
			"not key=value": "name=west,private",
		} {
			t.Run(name, func(t *testing.T) {
				_, err := parseEndpointFlag(raw)
				require.Error(t, err)
			})
		}
	})
}
