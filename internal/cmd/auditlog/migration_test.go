package auditlog

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	warn "github.com/confluentinc/cli/internal/pkg/errors"
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
				"cluster123": `
{
    "destinations": {
        "bootstrap_servers": [
            "audit.example.com:9092"
        ],
        "topics": {
            "confluent-audit-log-events_payroll": {
                "retention_ms": 50
            },
            "confluent-audit-log-events": {
                "retention_ms": 500
            }
        }
    },
    "default_topics": {
        "allowed": "confluent-audit-log-events",
        "denied": "confluent-audit-log-events"
    },
    "routes": {
        "crn://mds1.example.com/kafka=*/topic=payroll-*": {
            "produce": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            },
            "consume": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            }
        },
        "crn://some-authority/kafka=clusterX": {
          "other": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          }
        }
    },
    "excluded_principals": [
        "User:Alice"
    ]
}`,

				"clusterABC": `
{
  "destinations": {
      "bootstrap_servers": [
          "some-server"
      ],
      "topics": {
          "confluent-audit-log-events_payroll": {
              "retention_ms": 2592000000
          },
          "confluent-audit-log-events_billing": {
              "retention_ms": 2592000000
          },
          "DIFFERENT-DEFAULT-TOPIC": {
              "retention_ms": 100
          }
      }
  },
  "default_topics": {
      "allowed": "DIFFERENT-DEFAULT-TOPIC",
      "denied": "DIFFERENT-DEFAULT-TOPIC"
  },
  "routes": {
      "crn://mds1.example.com/kafka=*/topic=billing-*": {
          "produce": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          },
          "consume": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          },
          "other": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          }
      },
      "crn://diff-authority/kafka=different-cluster-id/topic=payroll-*": {
          "produce": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          },
          "consume": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          }
      },
      "crn://some-authority/kafka=clusterX": {
        "other": {
            "allowed": "confluent-audit-log-events_payroll",
            "denied": "confluent-audit-log-events_payroll"
        }
      }
  },
  "excluded_principals": [
      "User:Bob"
  ]
}`,
			},
			[]string{"new_bootstrap_2", "new_bootstrap_1"},
			"NEW.CRN.AUTHORITY.COM",
			test.LoadFixture(t, "auditlog/migration-result.golden"),
			[]string{
				fmt.Sprintf(warn.MismatchedKafkaClusterWarning,"cluster123", "crn://some-authority/kafka=clusterX"),
				fmt.Sprintf(warn.MismatchedKafkaClusterWarning,"clusterABC", "crn://diff-authority/kafka=different-cluster-id/topic=payroll-*"),
				fmt.Sprintf(warn.MismatchedKafkaClusterWarning,"clusterABC", "crn://some-authority/kafka=clusterX"),
				fmt.Sprintf(warn.MultipleCRNWarning,"cluster123", "[crn://mds1.example.com/ crn://some-authority/]"),
				fmt.Sprintf(warn.MultipleCRNWarning,"clusterABC", "[crn://diff-authority/ crn://mds1.example.com/ crn://some-authority/]"),
				fmt.Sprintf(warn.NewBootstrapWarning,"cluster123", "[audit.example.com:9092]", "[new_bootstrap_1 new_bootstrap_2]"),
				fmt.Sprintf(warn.NewBootstrapWarning,"clusterABC", "[some-server]", "[new_bootstrap_1 new_bootstrap_2]"),
				fmt.Sprintf(warn.NewExcludedPrincipalsWarning,"cluster123", "[User:Bob]"),
				fmt.Sprintf(warn.NewExcludedPrincipalsWarning,"clusterABC", "[User:Alice]"),
				fmt.Sprintf(warn.RepeatedRouteWarning,"crn://some-authority/kafka=clusterX"),
				fmt.Sprintf(warn.RetentionTimeDiscrepancyWarning,"confluent-audit-log-events_payroll", 2592000000),
			},
		},
		// This case has only one cluster, and it also has a route topic=* which has some existing routes,
		// we expect the script to leave those alone and add the other routes for topic=* (e.g authorize, describe, etc)
		{
			map[string]string{
				"cluster123": `
{
    "destinations": {
        "bootstrap_servers": [
            "audit.example.com:9092"
        ],
        "topics": {
            "confluent-audit-log-events_payroll": {
                "retention_ms": 50
            },
            "confluent-audit-log-events": {
                "retention_ms": 500
            }
        }
    },
    "default_topics": {
        "allowed": "confluent-audit-log-events",
        "denied": "confluent-audit-log-events_different_default_denied"
    },
    "routes": {
        "crn://mds1.example.com/kafka=*/topic=*": {
            "produce": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            },
            "consume": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            }
        },
        "crn://some-authority/kafka=clusterX": {
          "other": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          }
        },
        "crn://some-authority/kafka=clusterY": {
          "other": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          },
          "management": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events"
          }
        }
    },
    "excluded_principals": [
        "User:Alice"
    ]
}`,
			},
			[]string{"new_bootstrap_2", "new_bootstrap_1"},
			"NEW.CRN.AUTHORITY.COM",
			test.LoadFixture(t, "auditlog/migration-result-merge-topics.golden"),
			[]string{
				fmt.Sprintf(warn.MismatchedKafkaClusterWarning,"cluster123", "crn://some-authority/kafka=clusterX"),
				fmt.Sprintf(warn.MismatchedKafkaClusterWarning,"cluster123", "crn://some-authority/kafka=clusterY"),
				fmt.Sprintf(warn.MultipleCRNWarning,"cluster123", "[crn://mds1.example.com/ crn://some-authority/]"),
				fmt.Sprintf(warn.NewBootstrapWarning,"cluster123", "[audit.example.com:9092]", "[new_bootstrap_1 new_bootstrap_2]"),
				fmt.Sprintf(warn.OtherCategoryWarning,"crn://some-authority/kafka=clusterY" , "cluster123"),
			},
		},
	}

	for i, c := range testCases {
		var want mds.AuditLogConfigSpec
		err := json.Unmarshal([]byte(c.wantSpecAsString), &want)
		require.Nil(t, err)

		got, gotWarnings, err := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)

		require.Nil(t, err)
		require.Equal(t, want, got, "testCase: %d", i)
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
				"cluster123": `
{malformed string            "audit.example.com:9092"
        ],
        "topics": {
            "confluent-audit-log-events_payroll": {
                "retention_ms": 50
            },
            "confluent-audit-log-events": {
                "retention_ms": 500
            }
        }
    },
    "default_topics": {
        "allowed": "confluent-audit-log-events",
        "denied": "confluent-audit-log-events"
    },
    "routes": {
        "crn://mds1.example.com/kafka=*/topic=payroll-*": {
            "produce": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            },
            "consume": {
                "allowed": "confluent-audit-log-events_payroll",
                "denied": "confluent-audit-log-events_payroll"
            }
        },
        "crn://some-authority/kafka=clusterX": {
          "other": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          }
        }
    },
    "excluded_principals": [
        "User:Alice"
    ]
}`,

				"clusterABC": `
{
  "destinations": {
      "bootstrap_servers": [
          "some-server"
      ],
      "topics": {
          "confluent-audit-log-events_payroll": {
              "retention_ms": 2592000000
          },
          "confluent-audit-log-events_billing": {
              "retention_ms": 2592000000
          },
          "DIFFERENT-DEFAULT-TOPIC": {
              "retention_ms": 100
          }
      }
  },
  "default_topics": {
      "allowed": "DIFFERENT-DEFAULT-TOPIC",
      "denied": "DIFFERENT-DEFAULT-TOPIC"
  },
  "routes": {
      "crn://mds1.example.com/kafka=*/topic=billing-*": {
          "produce": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          },
          "consume": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          },
          "other": {
              "allowed": "confluent-audit-log-events_billing",
              "denied": "confluent-audit-log-events_billing"
          }
      },
      "crn://diff-authority/kafka=different-cluster-id/topic=payroll-*": {
          "produce": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          },
          "consume": {
              "allowed": "confluent-audit-log-events_payroll",
              "denied": "confluent-audit-log-events_payroll"
          }
      },
      "crn://some-authority/kafka=clusterX": {
        "other": {
            "allowed": "confluent-audit-log-events_payroll",
            "denied": "confluent-audit-log-events_payroll"
        }
      }
  },
  "excluded_principals": [
      "User:Bob"
  ]
}`,
			},
			[]string{"new_bootstrap_2", "new_bootstrap_1"},
			"NEW.CRN.AUTHORITY.COM",
		},
	}
	for _, c := range testCases {
		_, _, err := AuditLogConfigTranslation(c.clusterConfigs, c.bootstrapServers, c.crnAuthority)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "cluster123")
	}
}

func TestAuditLogConfigTranslationNilCase(t *testing.T) {
	var null mds.AuditLogConfigSpec
	val, _ := json.Marshal(null)
	clusterConfig := map[string]string{"abc": string(val)}
	var bootstrapServers []string
	var crnAuthority string

	_, _, err := AuditLogConfigTranslation(clusterConfig, bootstrapServers, crnAuthority)
	require.Nil(t, err)
}
