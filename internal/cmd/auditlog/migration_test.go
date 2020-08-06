package auditlog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/test"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

func TestAuditLogConfigTranslation(t *testing.T) {
	testCases := []struct {
		clusterConfigs   map[string]string
		bootstrapServers []string
		crnAuthority     string
		wantSpecAsString string
		wantWarnings     []string
	}{
		{
			map[string]string{
				"cluster123": "{\n    \"destinations\": {\n        \"bootstrap_servers\": [\n            \"audit.example.com:9092\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 50\n            },\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://mds1.example.com/kafka=*/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        },\n        \"crn://some-authority/kafka=clusterX\": {\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\"\n    ]\n}",

				"clusterABC": "{\n  \"destinations\": {\n      \"bootstrap_servers\": [\n          \"some-server\"\n      ],\n      \"topics\": {\n          \"confluent-audit-log-events_payroll\": {\n              \"retention_ms\": 2592000000\n          },\n          \"confluent-audit-log-events_billing\": {\n              \"retention_ms\": 2592000000\n          },\n          \"DIFFERENT-DEFAULT-TOPIC\": {\n              \"retention_ms\": 100\n          }\n      }\n  },\n  \"default_topics\": {\n      \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n      \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n  },\n  \"routes\": {\n      \"crn://mds1.example.com/kafka=*/topic=billing-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          }\n      },\n      \"crn://diff-authority/kafka=different-cluster-id/topic=payroll-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n      },\n      \"crn://some-authority/kafka=clusterX\": {\n        \"other\": {\n            \"allowed\": \"confluent-audit-log-events_payroll\",\n            \"denied\": \"confluent-audit-log-events_payroll\"\n        }\n      }\n  },\n  \"excluded_principals\": [\n      \"User:Bob\"\n  ]\n}",
			},
			[]string{"new_bootstrap_2", "new_bootstrap_1"},
			"NEW.CRN.AUTHORITY.COM",
			test.LoadFixture(t, "auditlog/migration-result.golden"),
			[]string{
				"Mismatched Kafka Cluster Warning: Cluster \"cluster123\" has a route with a different clusterId. Route: \"crn://some-authority/kafka=clusterX\".",
				"Mismatched Kafka Cluster Warning: Cluster \"clusterABC\" has a route with a different clusterId. Route: \"crn://diff-authority/kafka=different-cluster-id/topic=payroll-*\".",
				"Mismatched Kafka Cluster Warning: Cluster \"clusterABC\" has a route with a different clusterId. Route: \"crn://some-authority/kafka=clusterX\".",
				"Multiple CRN Authorities Warning: Cluster \"cluster123\" had multiple CRN Authorities in its routes: [crn://mds1.example.com/ crn://some-authority/].",
				"Multiple CRN Authorities Warning: Cluster \"clusterABC\" had multiple CRN Authorities in its routes: [crn://diff-authority/ crn://mds1.example.com/ crn://some-authority/].",
				"New Bootstrap Servers Warning: Cluster \"cluster123\" currently has bootstrap servers = [audit.example.com:9092]. Replacing with [new_bootstrap_1 new_bootstrap_2].",
				"New Bootstrap Servers Warning: Cluster \"clusterABC\" currently has bootstrap servers = [some-server]. Replacing with [new_bootstrap_1 new_bootstrap_2].",
				"New Excluded Principals Warning: Cluster \"cluster123\" will now also exclude the following principals: [User:Bob].", "New Excluded Principals Warning: Cluster \"clusterABC\" will now also exclude the following principals: [User:Alice].",
				"Repeated Route Warning: Route Name : \"crn://some-authority/kafka=clusterX\".",
				"Retention Time Discrepancy Warning: Topic \"confluent-audit-log-events_payroll\" had discrepancies with retention time. Using max: 2592000000.",
			},
		},
	}

	for _, c := range testCases {
		var want mds.AuditLogConfigSpec
		_ = json.Unmarshal([]byte(c.wantSpecAsString), &want)

		got, gotWarnings, err := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)
		require.Nil(t, err)
		require.Equal(t, want, got)
		require.Equal(t, c.wantWarnings, gotWarnings)
	}
}

func TestAuditLogConfigTranslationMalformedProperties(t *testing.T) {
	testCases := []struct {
		clusterConfigs   map[string]string
		bootstrapServers []string
		crnAuthority     string
	}{
		{
			map[string]string{
				"cluster123": "{malformed string            \"audit.example.com:9092\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 50\n            },\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://mds1.example.com/kafka=*/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        },\n        \"crn://some-authority/kafka=clusterX\": {\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\"\n    ]\n}",

				"clusterABC": "{\n  \"destinations\": {\n      \"bootstrap_servers\": [\n          \"some-server\"\n      ],\n      \"topics\": {\n          \"confluent-audit-log-events_payroll\": {\n              \"retention_ms\": 2592000000\n          },\n          \"confluent-audit-log-events_billing\": {\n              \"retention_ms\": 2592000000\n          },\n          \"DIFFERENT-DEFAULT-TOPIC\": {\n              \"retention_ms\": 100\n          }\n      }\n  },\n  \"default_topics\": {\n      \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n      \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n  },\n  \"routes\": {\n      \"crn://mds1.example.com/kafka=*/topic=billing-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          }\n      },\n      \"crn://diff-authority/kafka=different-cluster-id/topic=payroll-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n      },\n      \"crn://some-authority/kafka=clusterX\": {\n        \"other\": {\n            \"allowed\": \"confluent-audit-log-events_payroll\",\n            \"denied\": \"confluent-audit-log-events_payroll\"\n        }\n      }\n  },\n  \"excluded_principals\": [\n      \"User:Bob\"\n  ]\n}",
			},
			[]string{"new_bootstrap_2", "new_bootstrap_1"},
			"NEW.CRN.AUTHORITY.COM",
		},
	}
	for _, c := range testCases {
		_, _, err := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "cluster123");
	}
}

func TestAuditLogConfigTranslationNilCase(t *testing.T) {
	var null mds.AuditLogConfigSpec
	val, _ := json.Marshal(null);
	clusterConfig := map[string]string{"abc": string(val)}
	var bootstrapServers []string
	var crnAuthority string

	_, _, err := AuditLogConfigTranslation(clusterConfig, bootstrapServers, crnAuthority)
	require.Nil(t, err)
}
