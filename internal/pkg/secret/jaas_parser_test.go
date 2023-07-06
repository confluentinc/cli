package secret

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func TestJAASParser_String(t *testing.T) {
	type args struct {
		key             string
		contents        string
		expectedContent string
	}
	tests := []struct {
		name           string
		args           *args
		wantErr        bool
		wantErrMsg     string
		wantConfigFile string
	}{
		{
			name: "Valid: JAAS config entry",
			args: &args{
				key: "listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config",
				contents: `com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;`,
				expectedContent: `listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useKeyTab = false
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useTicketCache = true
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/doNotPrompt = true
`,
			},
			wantErr: false,
		},
		{
			name: "Valid: JAAS config entry with backslash",
			args: &args{
				key: "listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config",
				contents: `com.sun.security.auth.module.Krb5LoginModule required \
  useKeyTab=false \
  useTicketCache=true \
  doNotPrompt=true;`,
				expectedContent: `listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useKeyTab = false
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useTicketCache = true
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/doNotPrompt = true
`,
			},
			wantErr: false,
		},
		{
			name: "Invalid: login module control flag missing in JAAS file",
			args: &args{
				contents: ` com.sun.security.auth.module.Krb5LoginModule
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;
`,
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.InvalidJAASConfigErrorMsg, errors.LoginModuleControlFlagErrorMsg),
		},
		{
			name: "Invalid: ; field missing in JAAS file",
			args: &args{
				contents: `com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true
`,
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.InvalidJAASConfigErrorMsg, fmt.Sprintf(errors.ExpectedConfigNameErrorMsg, "")),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			parser := NewJAASParser()
			props, err := parser.ParseJAASConfigurationEntry(test.args.contents, test.args.key)
			if test.wantErr {
				req.Error(err)
				req.Contains(err.Error(), test.wantErrMsg)
			} else {
				req.NoError(err)
				parsedString := props.String()
				req.Equal(test.args.expectedContent, parsedString)
			}
		})
	}
}

func TestJAASParser_StringUpdate(t *testing.T) {
	type args struct {
		key             string
		contents        string
		expectedContent string
		originalContent string
		operation       string
	}
	tests := []struct {
		name           string
		args           *args
		wantErr        bool
		wantErrMsg     string
		wantConfigFile string
	}{
		{
			name: "Valid: JAAS config entry",
			args: &args{
				key: "listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config",
				originalContent: `com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;`,
				contents: `listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useKeyTab = true
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/useTicketCache = true
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/com.sun.security.auth.module.Krb5LoginModule/doNotPrompt = true
`,
				expectedContent: `listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config = com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  useTicketCache=true
  doNotPrompt=true;
`,
				operation: "update",
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			parser := NewJAASParser()
			_, err := parser.ParseJAASConfigurationEntry(test.args.originalContent, test.args.key)
			req.NoError(err)
			updatedProps := properties.MustLoadString(test.args.contents)
			jaasConfig, err := parser.ConvertPropertiesToJAAS(updatedProps, test.args.operation)
			if test.wantErr {
				req.Error(err)
				req.Contains(err.Error(), test.wantErrMsg)
			} else {
				req.NoError(err)
				req.Equal(test.args.expectedContent, jaasConfig.String())
			}
		})
	}
}
