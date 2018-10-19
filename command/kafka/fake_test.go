package kafka

import (
	"reflect"
	"github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
	"testing"
)

var conf *shared.Config

var resource1 = &kafka.ResourcePatternConfig{
	Name:         "test",
	ResourceType: kafka.ResourcePatternConfig_TOPIC.String(),
	PatternType:  kafka.ResourcePatternConfig_LITERAL.String(),
}
var resource2 = &kafka.ResourcePatternConfig{
	Name:         "test",
	ResourceType: kafka.ResourcePatternConfig_TOPIC.String(),
	PatternType:  kafka.ResourcePatternConfig_PREFIXED.String(),
}
var ace1 = &kafka.AccessControlEntryConfig{
	Principal:      "user:david bowie",
	Operation:      kafka.AccessControlEntryConfig_READ.String(),
	Host:           "*",
	PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
}
var ace2 = &kafka.AccessControlEntryConfig{
	Principal:      "user:david bowie",
	Operation:      kafka.AccessControlEntryConfig_WRITE.String(),
	Host:           "*",
	PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
}

func TestKafkaTopicSuccess(t *testing.T) {
	tests := []struct {
		name string
		args    []string
		expect *ACLConfiguration
	}{
		{
			name: "literal",
			args:    []string{"acl", "create", "--allow", "--principal", "david bowie", "--operation", "read", "--topic", "test"},
			expect: &ACLConfiguration{KafkaAPIACLRequest: &kafka.KafkaAPIACLRequest{Pattern: resource1, Entry: ace1}},
		},
		{
			name: "prefixed",
			args:    []string{"acl", "create", "--deny", "--principal", "david bowie", "--operation", "write", "--topic", "test*"},
			expect: &ACLConfiguration{KafkaAPIACLRequest: &kafka.KafkaAPIACLRequest{Pattern: resource2, Entry: ace2}},
		},
	}

	var current int
	for _, test := range tests {
		cmd, _ := New(conf)
		cmd2, args, _ := cmd.Find(test.args)
		cmd2.ParseFlags(args)
		actual := validateAddDelete(ParseCMD(cmd2))
		if len(actual.errors) > 0 || !reflect.DeepEqual(actual, test.expect) {
			goto fail
		}
		current++
	}
	return
	fail:
	t.Errorf("%s failed. expected %+v", tests[current].name, tests[current].expect)
	t.FailNow()
}

func TestACLFailure(t *testing.T) {
	tests := []struct {
		name string
		args    []string
		expect int
	}{
		{
			name: "missing_permissionType",
			args:    []string{"acl", "create", "--principal", "david bowie", "--operation", "read", "--topic", "test"},
			expect: 1,
		},
		{
			name: "multiple_resource",
			args:    []string{"acl", "create", "--deny", "--principal", "david bowie", "--operation", "write", "--topic", "test*",
			"--consumer_group", "labyrinth"},
			expect: 1,
		},
		{
			name: "invalid_operation",
			args:    []string{"acl", "create", "--deny", "--principal", "david bowie", "--operation", "party", "--topic", "test*"},
			expect: 2,
		},
	}

	var current int
	for _, test := range tests {
		cmd, _ := New(conf)
		cmd2, args, _ := cmd.Find(test.args)
		cmd2.ParseFlags(args)
		actual := validateAddDelete(ParseCMD(cmd2))
		if len(actual.errors) != test.expect {
			goto fail
		}
		current++
	}
	return
fail:
	t.Errorf("%s failed. expected %+v", tests[current].name, tests[current].expect)
	t.FailNow()
}


func init() {
	conf := shared.NewConfig()
	conf.Auth = &shared.AuthConfig{
		User:    new(v1.User),
		Account: &v1.Account{Id: "test"},
	}
}
