package audit_log_migration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go"
)

func TestAuditLogConfigTranslation(t *testing.T) {
	testCases := []struct {
		clusterConfigs   map[string]string
		bootstrapServers string
		crnAuthority     string
		wantSpecAsString string
	}{
		{
			map[string]string{
				"cluster123": "{\n    \"destinations\": {\n        \"bootstrap_servers\": [\n            \"audit.example.com:9092\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 50\n            },\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://mds1.example.com/kafka=*/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\"\n    ]\n}",

				"clusterABC": "{\n  \"destinations\": {\n      \"bootstrap_servers\": [\n          \"some-server\"\n      ],\n      \"topics\": {\n          \"confluent-audit-log-events_payroll\": {\n              \"retention_ms\": 2592000000\n          },\n          \"confluent-audit-log-events_billing\": {\n              \"retention_ms\": 2592000000\n          },\n          \"DIFFERENT-DEFAULT-TOPIC\": {\n              \"retention_ms\": 100\n          }\n      }\n  },\n  \"default_topics\": {\n      \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n      \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n  },\n  \"routes\": {\n      \"crn://mds1.example.com/kafka=*/topic=billing-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          },\n          \"other\": {\n              \"allowed\": \"confluent-audit-log-events_billing\",\n              \"denied\": \"confluent-audit-log-events_billing\"\n          }\n      },\n      \"crn://diff-authority/kafka=*/topic=payroll-*\": {\n          \"produce\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          },\n          \"consume\": {\n              \"allowed\": \"confluent-audit-log-events_payroll\",\n              \"denied\": \"confluent-audit-log-events_payroll\"\n          }\n      }\n  },\n  \"excluded_principals\": [\n      \"User:Bob\"\n  ]\n}",
			},
			"new_bootstrap",
			"NEW.CRN.AUTHORITY.COM",
			"{\n    \"destinations\": {\n        \"bootstrap_servers\": [\n            \"new_bootstrap\"\n        ],\n        \"topics\": {\n            \"confluent-audit-log-events\": {\n                \"retention_ms\": 500\n            },\n            \"confluent-audit-log-events_payroll\": {\n                \"retention_ms\": 2592000000\n            },\n            \"confluent-audit-log-events_billing\": {\n                \"retention_ms\": 2592000000\n            },\n            \"DIFFERENT-DEFAULT-TOPIC\": {\n                \"retention_ms\": 100\n            }\n        }\n    },\n    \"default_topics\": {\n        \"allowed\": \"confluent-audit-log-events\",\n        \"denied\": \"confluent-audit-log-events\"\n    },\n    \"routes\": {\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=billing-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            },\n            \"other\": {\n                \"allowed\": \"confluent-audit-log-events_billing\",\n                \"denied\": \"confluent-audit-log-events_billing\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=cluster123/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=payroll-*\": {\n            \"produce\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"consume\": {\n                \"allowed\": \"confluent-audit-log-events_payroll\",\n                \"denied\": \"confluent-audit-log-events_payroll\"\n            },\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/topic=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/connect=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        },\n        \"crn://NEW.CRN.AUTHORITY.COM/kafka=clusterABC/ksql=*\": {\n            \"other\": {\n                \"allowed\": \"DIFFERENT-DEFAULT-TOPIC\",\n                \"denied\": \"DIFFERENT-DEFAULT-TOPIC\"\n            }\n        }\n    },\n    \"excluded_principals\": [\n        \"User:Alice\",\n        \"User:Bob\"\n    ]\n}",
		},
	}
	for _, c := range testCases {
		var want mds.AuditLogConfigSpec
		json.Unmarshal([]byte(c.wantSpecAsString), &want)

		got, err := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)
		require.Nil(t, err)
		require.Equal(t, want, got)
	}
}
