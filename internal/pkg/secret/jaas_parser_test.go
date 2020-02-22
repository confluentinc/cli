package secret

import (
	"fmt"
	"github.com/confluentinc/properties"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestJAASParser_Load(t *testing.T) {
	type args struct {
		contents        string
		path            string
		secureDir       string
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
			name: "Valid: JAAS File",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=false
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
				expectedContent: `Client/com.sun.security.auth.module.Krb5LoginModule/useKeyTab = false
Client/com.sun.security.auth.module.Krb5LoginModule/useTicketCache = true
Client/com.sun.security.auth.module.Krb5LoginModule/doNotPrompt = true
Client/com.sun.security.auth.module.PascalLoginModule/useKeyTab = false
Client/com.sun.security.auth.module.PascalLoginModule/doNotPrompt = true
`,
			},
			wantErr: false,
		},
		{
			name: "Invalid: login module control flag missing in JAAS file",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
			},
			wantErr:    true,
			wantErrMsg: "invalid jaas file: login module control flag not specified",
		},
		{
			name: "Invalid: value field missing in JAAS file",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab
  useTicketCache=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
			},
			wantErr:    true,
			wantErrMsg: "invalid jaas file: value not specified for the key",
		},
		{
			name: "Invalid: ; field missing in JAAS file",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
			},
			wantErr:    true,
			wantErrMsg: "invalid jaas file: expected a configuration name received }",
		},
		{
			name: "Invalid: } field missing in JAAS file",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=false
  useTicketCache=true
  doNotPrompt=true;
`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
			},
			wantErr:    true,
			wantErrMsg: "invalid jaas file: not terminated with '}'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(tt.args.secureDir)
			req := require.New(t)
			err := createDir(tt.args.secureDir, tt.args.path, tt.args.contents)
			req.NoError(err)

			parser := NewJAASParser()
			props, err := parser.Load(tt.args.path)
			if tt.wantErr {
				req.Error(err)
				req.Contains(err.Error(), tt.wantErrMsg)
			} else {
				req.NoError(err)
				parsedString := props.String()
				req.Equal(tt.args.expectedContent, parsedString)
			}

			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func TestJAASParser_Write(t *testing.T) {
	type args struct {
		contents        string
		path            string
		secureDir       string
		expectedContent string
		updatedConfig   string
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
			name: "Valid: Update config in JAAS File",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  password=pass%25#
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
				operation: "update",
				updatedConfig: `Client/com.sun.security.auth.module.Krb5LoginModule/password = #######
Client/com.sun.security.auth.module.PascalLoginModule/useKeyTab = false
`,
				expectedContent: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  password=#######
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=false
  doNotPrompt=true;
};`,
			},
			wantErr: false,
		},
		{
			name: "Valid: Delete config from JAAS File",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  password=${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.json/credentials.ssl\\.keystore\\.password}
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
				operation: "delete",
				updatedConfig: `Client/com.sun.security.auth.module.Krb5LoginModule/password
`,
				expectedContent: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
			},
			wantErr: false,
		},
		{
			name: "Valid: Add config from JAAS File",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  doNotPrompt=true;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
				operation: "update",
				updatedConfig: `Client/com.sun.security.auth.module.Krb5LoginModule/password = pass%25#
`,
				expectedContent: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  doNotPrompt=true
 password=pass%25#;
          
  com.sun.security.auth.module.PascalLoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
			},
			wantErr: false,
		},
		{
			name: "Valid: Add config from JAAS File",
			args: &args{
				contents: `Client {
  com.sun.security.auth.module.Krb5LoginModule required
  useKeyTab=true
  doNotPrompt=true;
};`,
				path:      "/tmp/securePass987/parser/config.conf",
				secureDir: "/tmp/securePass987/parser",
				operation: "add",
				updatedConfig: `Client/com.sun.security.auth.module.PascalLoginModule/location = /var/log
`,
			},
			wantErr:    true,
			wantErrMsg: "config Client/com.sun.security.auth.module.PascalLoginModule/location not present in the file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(tt.args.secureDir)
			req := require.New(t)
			err := createDir(tt.args.secureDir, tt.args.path, tt.args.contents)
			req.NoError(err)
			props := properties.MustLoadString(tt.args.updatedConfig)
			parser := NewJAASParser()
			err = parser.Write(tt.args.path, props, tt.args.operation, false)
			if tt.wantErr {
				req.Error(err)
				req.Contains(err.Error(), tt.wantErrMsg)
			} else {
				req.NoError(err)
				jaasFile, err := ioutil.ReadFile(tt.args.path)
				req.NoError(err)
				req.Equal(tt.args.expectedContent, string(jaasFile))
			}

			os.RemoveAll(tt.args.secureDir)
		})
	}
}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			parser := NewJAASParser()
			props, err := parser.ParseJAASConfigurationEntry(tt.args.contents, tt.args.key)
			if tt.wantErr {
				req.Error(err)
				req.Contains(err.Error(), tt.wantErrMsg)
			} else {
				req.NoError(err)
				parsedString := props.String()
				req.Equal(tt.args.expectedContent, parsedString)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			parser := NewJAASParser()
			_, err := parser.ParseJAASConfigurationEntry(tt.args.originalContent, tt.args.key)
			req.NoError(err)
			updatedProps := properties.MustLoadString(tt.args.contents)
			jaasConfig, err := parser.ConvertPropertiesToJAAS(updatedProps, tt.args.operation, false)
			if tt.wantErr {
				req.Error(err)
				req.Contains(err.Error(), tt.wantErrMsg)
			} else {
				req.NoError(err)
				req.Equal(tt.args.expectedContent, jaasConfig.String())
			}
		})
	}
}

func createDir(secureDir string, path string, contents string) error {
	err := os.MkdirAll(secureDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create secrets directory")
	}
	return ioutil.WriteFile(path, []byte(contents), 0644)
}
