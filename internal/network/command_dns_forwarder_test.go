package network

import (
	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-dnsforwarder/v1"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDomainFlagToMap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		map1 := map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings{
			"example": {Zone: networkingdnsforwarderv1.PtrString("zone1"), Project: networkingdnsforwarderv1.PtrString("project1")},
		}
		file, _ := os.CreateTemp(os.TempDir(), "test")
		_, _ = file.Write([]byte("example=zone1,project1"))
		defer func() {
			_ = os.Remove(file.Name())
		}()
		actual, _ := DomainFlagToMap(file.Name())
		assert.Equal(t, map1, actual)
	})
	t.Run("fail, wrong path", func(t *testing.T) {
		file, _ := os.CreateTemp(os.TempDir(), "test")
		_, _ = file.Write([]byte("example=zone1,project1"))
		defer func() {
			_ = os.Remove(file.Name())
		}()
		_, err := DomainFlagToMap("XYZ")
		assert.Error(t, err)
	})
	t.Run("fail, wrong values", func(t *testing.T) {
		map1 := map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings{
			"example": {Zone: networkingdnsforwarderv1.PtrString("zone1"), Project: networkingdnsforwarderv1.PtrString("project1")},
		}
		file, _ := os.CreateTemp(os.TempDir(), "test")
		_, _ = file.Write([]byte("example=zone1,projectxyz"))
		defer func() {
			_ = os.Remove(file.Name())
		}()
		actual, _ := DomainFlagToMap(file.Name())
		assert.NotEqual(t, map1, actual)
	})
}

func TestConvertToTypeMapString(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		map1 := map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings{
			"example": {Zone: networkingdnsforwarderv1.PtrString("zone1"), Project: networkingdnsforwarderv1.PtrString("project1")},
		}
		map1Expected := map[string]string{
			"example": "{zone1, project1}",
		}
		actual := convertToTypeMapString(map1)
		assert.Equal(t, map1Expected, actual)
	})

	t.Run("fail", func(t *testing.T) {
		map1 := map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings{
			"example": {Zone: networkingdnsforwarderv1.PtrString("zone1"), Project: networkingdnsforwarderv1.PtrString("project1")},
		}
		map1Expected := map[string]string{
			"example": "{zone1, projectxyz}",
		}
		actual := convertToTypeMapString(map1)
		assert.NotEqual(t, map1Expected, actual)
	})
}
