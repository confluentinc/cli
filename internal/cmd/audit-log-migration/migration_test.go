package audit_log_migration

import (
  "testing"
  "encoding/json"

	"github.com/stretchr/testify/require"

  "github.com/confluentinc/mds-sdk-go"
)

func TestAuditLogConfigTranslation(t *testing.T) {
	testCases := []struct {
		clusterConfigs  map[string]string
		bootstrapServers string
		crnAuthority  string
    wantSpecAsString string
	}{
		{
			map[string]string{
        "cluster123": "{\n    \"destinations\": {\n        \"bootstrap_servers\": [\n            \"audit.example.com:9092\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 50\n            },\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://mds1.example.com/kafka=*/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\"\n    ]\n}",

        "clusterABC": "{\n  \"destinations\": {\n      \"bootstrap_servers\": [\n          \"some-server\"\n      ],\n      \"topics\": {\n          \"confluent-audit-log-events_payroll\": {\n              \"retention_ms\": 2592000000\n          },\n          \"confluent-audit-log-events_billing\": {\n              \"retention_ms\": 2592000000\n          },\n          \"DIFFERENT-DEFAULT-TOPIC\": {\n              \"retention_ms\": 100\n          }\n      }\n  },\n  \"default_topics\": {\n      \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n      \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n  },\n  \"routes\": {\n      \"crn://mds1.example.com/kafka=*/topic=billing-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          }\n      },\n      \"crn://mds1.example.com/kafka=*/topic=payroll-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n      }\n  },\n  \"excluded_principals\": [\n      \"User:Bob\"\n  ]\n}",
      },
      "new_bootstrap",
      "NEW.CRN.AUTHORITY.COM",
      "{\n    \"destinations\": {\n        \"bootstrap_servers\": [\n            \"new_bootstrap\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            },\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 2592000000\n            },\n            \"confluent-audit-log-events_billing\": {\n                \"retention_ms\": 2592000000\n            },\n            \"DIFFERENT-DEFAULT-TOPIC\": {\n                \"retention_ms\": 100\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=billing-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            },\n            \"other\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=cluster123/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/connect=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/ksql=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\",\n        \"User:Bob\"\n    ]\n}",
		},
	}
	for _, c := range testCases {
    // billing_topic := "confluent-audit-log-events_billing"
    // payroll_topic := "confluent-audit-log-events_payroll"
    // different_default_topic := "DIFFERENT-DEFAULT-TOPIC"

    // want := mds.AuditLogConfigSpec{
    //   Destinations: mds.AuditLogConfigDestinations{
    //     BootstrapServers: []string{"new_bootstrap"},
    //     Topics: map[string]mds.AuditLogConfigDestinationConfig {
    //       "confluent-audit-log-events": mds.AuditLogConfigDestinationConfig{
    //         RetentionMs: 7776000000,
    //       },
    //       "confluent-audit-log-events_payroll": mds.AuditLogConfigDestinationConfig{
    //         RetentionMs: 2592000000,
    //       },
    //       "confluent-audit-log-events_billing": mds.AuditLogConfigDestinationConfig{
    //         RetentionMs: 2592000000,
    //       },
    //     },
    //   },
    //   DefaultTopics: mds.AuditLogConfigDefaultTopics{
    //     Allowed: "confluent-audit-log-events",
    //     Denied: "confluent-audit-log-events",
    //   },
    //   Routes: map[string]mds.AuditLogConfigRouteCategories{
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=billing-*": mds.AuditLogConfigRouteCategories {
    //       Produce: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &billing_topic,
    //         Denied: &billing_topic,
    //       },
    //       Consume: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &billing_topic,
    //         Denied: &billing_topic,
    //       },
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &billing_topic,
    //         Denied: &billing_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=cluster123/topic=payroll-*": mds.AuditLogConfigRouteCategories {
    //       Produce: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &payroll_topic,
    //         Denied: &payroll_topic,
    //       },
    //       Consume: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &payroll_topic,
    //         Denied: &payroll_topic,
    //       },
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &payroll_topic,
    //         Denied: &payroll_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=payroll-*": mds.AuditLogConfigRouteCategories {
    //       Produce: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &payroll_topic,
    //         Denied: &payroll_topic,
    //       },
    //       Consume: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &payroll_topic,
    //         Denied: &payroll_topic,
    //       },
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &different_default_topic,
    //         Denied: &different_default_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC": mds.AuditLogConfigRouteCategories {
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &different_default_topic,
    //         Denied: &different_default_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=*": mds.AuditLogConfigRouteCategories {
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &different_default_topic,
    //         Denied: &different_default_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/connect=*": mds.AuditLogConfigRouteCategories {
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &different_default_topic,
    //         Denied: &different_default_topic,
    //       },
    //     },
    //     "crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/ksql=*": mds.AuditLogConfigRouteCategories {
    //       Other: &mds.AuditLogConfigRouteCategoryTopics{
    //         Allowed: &different_default_topic,
    //         Denied: &different_default_topic,
    //       },
    //     },
    //   },
    //   ExcludedPrincipals: []string{"User:Alice","User:Bob"},
    // }
    var want mds.AuditLogConfigSpec
    json.Unmarshal([]byte(c.wantSpecAsString), &want)


    got := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)
    require.Equal(t, want, got)
    require.NotNil(t, want)
    require.NotNil(t, got)
	}
}
