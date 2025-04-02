package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
)

var (
	IsSsoEnabled = false
)

// Handler for: "/security/1.0/registry/clusters"
func handleRegistryClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
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

		if r.Method == http.MethodDelete {
			clusterName := r.URL.Query().Get("clusterName")
			require.NotEmpty(t, clusterName)
		}

		if r.Method == http.MethodPost {
			var clusterInfos []*mdsv1.ClusterInfo
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
func handleAuthenticate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := &mdsv1.AuthenticationResponse{
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

// Handler for: "/security/1.0/features"
func handleFeatures(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := &mdsv1.FeaturesInfo{
			Features: map[string]bool{"oidc.login.device.1.enabled": IsSsoEnabled},
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/security/1.0/oidc/device/authenticate"
func handleDeviceAuthenticate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := &mdsv1.InitDeviceAuthResponse{
			VerificationUri: "https://example.com",
			Interval:        5,
			ExpiresIn:       30,
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/security/1.0/oidc/device/check-auth"
func handleDeviceCheckAuth(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := &mdsv1.CheckDeviceAuthResponse{
			Complete:  true,
			AuthToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			ExpiresIn: 10,
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/security/1.0/oidc/device/extend-auth"
func handleDeviceExtendAuth(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// Handler for: "/security/1.0/audit/config"
func handleAuditConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			config := mdsv1.AuditLogConfigSpec{
				Destinations: mdsv1.AuditLogConfigDestinations{
					Topics: map[string]mdsv1.AuditLogConfigDestinationConfig{
						"confluent-audit-log-events_general_allowed_events": {RetentionMs: 2592000000},
						"confluent-audit-log-events_general_denied_events":  {RetentionMs: 7776000000},
					},
				},
				ExcludedPrincipals: &[]string{"User:Alice", "User:service_account_id"},
				DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
					Allowed: "confluent-audit-log-events_general_allowed_events",
					Denied:  "confluent-audit-log-events_general_denied_events",
				},
				Routes: &map[string]mdsv1.AuditLogConfigRouteCategories{
					"crn://mds1.example.com/kafka=*/topic=*": {
						Authorize: &mdsv1.AuditLogConfigRouteCategoryTopics{
							Allowed: ptrString("confluent-audit-log-events_general_allowed_events"),
							Denied:  ptrString("confluent-audit-log-events_general_denied_events"),
						},
					},
				},
				Metadata: &mdsv1.AuditLogConfigMetadata{
					ResourceVersion: "ASNFZ4mrze8BI0VniavN7w",
				},
			}
			err := json.NewEncoder(w).Encode(config)
			require.NoError(t, err)
		}
		if r.Method == http.MethodPut {
			var configSpec mdsv1.AuditLogConfigSpec
			err := json.NewDecoder(r.Body).Decode(&configSpec)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(configSpec)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/security/1.0/audit/lookup"
func handleAuditLookup(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lookup := mdsv1.AuditLogConfigResolveResourceRouteResponse{
				Route: "crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test",
				Categories: mdsv1.AuditLogConfigRouteCategories{
					Authorize: &mdsv1.AuditLogConfigRouteCategoryTopics{
						Allowed: ptrString("confluent-audit-log-events_general_allowed_events"),
						Denied:  ptrString("confluent-audit-log-events_general_denied_events"),
					},
					Consume: &mdsv1.AuditLogConfigRouteCategoryTopics{
						Denied: ptrString("confluent-audit-log-events_finance_denied"),
					},
					Management: &mdsv1.AuditLogConfigRouteCategoryTopics{
						Allowed: ptrString("confluent-audit-log-events_general_allowed_events"),
						Denied:  ptrString("confluent-audit-log-events_general_denied_events"),
					},
					Produce: &mdsv1.AuditLogConfigRouteCategoryTopics{
						Allowed: ptrString("confluent-audit-log-events_finance_produce_allowed"),
						Denied:  ptrString("confluent-audit-log-events_finance_denied"),
					},
				},
			}
			err := json.NewEncoder(w).Encode(lookup)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/security/1.0/audit/routes"
func handleAuditRoutes(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			routes := mdsv1.AuditLogConfigListRoutesResponse{
				DefaultTopics: mdsv1.AuditLogConfigDefaultTopics{
					Allowed: "confluent-audit-log-events_general_allowed_events",
					Denied:  "confluent-audit-log-events_general_denied_events",
				},
				Routes: &map[string]mdsv1.AuditLogConfigRouteCategories{
					"crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test/connector=from-db4": {
						Management: &mdsv1.AuditLogConfigRouteCategoryTopics{Allowed: ptrString(""), Denied: ptrString("")},
					},
				},
			}
			err := json.NewEncoder(w).Encode(routes)
			require.NoError(t, err)
		}
	}
}
