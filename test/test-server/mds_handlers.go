package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/stretchr/testify/require"
)

// Handler for: "/security/1.0/registry/clusters"
func (m MdsRouter) HandleRegistryClusters(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/json")
			clusterType := r.URL.Query().Get("clusterType")
			response := `[ {
		"clusterName": "theMdsConnectCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "connect-cluster": "connect-name" } },
		"hosts": [ { "host": "10.5.5.5", "port": 9005 } ],
        "protocol": "HTTPS"
	  },{
		"clusterName": "theMdsKSQLCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		"hosts": [ { "host": "10.4.4.4", "port": 9004 } ],
        "protocol": "HTTPS"
	  },{
		"clusterName": "theMdsKafkaCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID" } },
		"hosts": [ { "host": "10.10.10.10", "port": 8090 },{ "host": "mds.example.com", "port": 8090 } ],
        "protocol": "SASL_PLAINTEXT"
	  },{
		"clusterName": "theMdsSchemaRegistryCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "schema-registry-cluster": "schema-registry-name" } },
		"hosts": [ { "host": "10.3.3.3", "port": 9003 } ],
        "protocol": "HTTPS"
	} ]`
			if clusterType == "ksql-cluster" {
				response = `[ {
		    "clusterName": "theMdsKSQLCluster",
		    "scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		    "hosts": [ { "host": "10.4.4.4", "port": 9004 } ],
            "protocol": "HTTPS"
			} ]`
			}
			if clusterType == "kafka-cluster" {
				response = `[ {
			"clusterName": "theMdsKafkaCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID" } },
			"hosts": [ { "host": "10.10.10.10", "port": 8090 },{ "host": "mds.example.com", "port": 8090 } ],
        	"protocol": "SASL_PLAINTEXT"
			} ]`
			}
			if clusterType == "connect-cluster" {
				response = `[ {
			"clusterName": "theMdsConnectCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "connect-cluster": "connect-name" } },
			"hosts": [ { "host": "10.5.5.5", "port": 9005 } ],
        	"protocol": "HTTPS"
			} ]`
			}
			if clusterType == "schema-registry-cluster" {
				response = `[ {
			"clusterName": "theMdsSchemaRegistryCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "schema-registry-cluster": "schema-registry-name" } },
			"hosts": [ { "host": "10.3.3.3", "port": 9003 } ],
        	"protocol": "HTTPS"
			} ]`
			}
			_, err := io.WriteString(w, response)
			require.NoError(t, err)
		}

		if r.Method == "DELETE" {
			clusterName := r.URL.Query().Get("clusterName")
			require.NotEmpty(t, clusterName)
		}

		if r.Method == "POST" {
			var clusterInfos []*mds.ClusterInfo
			err := json.NewDecoder(r.Body).Decode(&clusterInfos)
			require.NoError(t, err)
			require.NotEmpty(t, clusterInfos)
			for _, clusterInfo := range clusterInfos {
				require.NotEmpty(t, clusterInfo.ClusterName)
				require.NotEmpty(t, clusterInfo.Hosts)
				require.NotEmpty(t, clusterInfo.Scope)
				require.NotEmpty(t, clusterInfo.Protocol)
			}
		}
	}
}

// Handler for: "/security/1.0/authenticate"
func (m MdsRouter) HandleAuthenticate(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		reply := &mds.AuthenticationResponse{
			AuthToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			TokenType: "dunno",
			ExpiresIn: 9999999999,
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
