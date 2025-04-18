package secret

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func TestPasswordProtectionSuite_CreateMasterKey(t *testing.T) {
	type args struct {
		masterKeyPassphrase   string
		localSecureConfigPath string
		validateDiffKey       bool
		secureDir             string
	}
	tests := []struct {
		name       string
		args       *args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "ValidTestCase: valid create master key",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: valid create master key with space at the end",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123 ",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: valid create master key with tab at the end",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123\t",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: valid create master key with new line at the end",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123\n",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: verify for same passphrase it generates a different master key",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       true,
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: empty passphrase",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
			},
			wantErr:    true,
			wantErrMsg: errors.EmptyPassphraseErrorMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			err := os.MkdirAll(test.args.secureDir, os.ModePerm)
			defer os.RemoveAll(test.args.secureDir)
			req.NoError(err)

			plugin := NewPasswordProtectionPlugin()

			key, err := plugin.CreateMasterKey(test.args.masterKeyPassphrase, test.args.localSecureConfigPath)
			checkError(err, test.wantErr, test.wantErrMsg, req)
			if !test.wantErr {
				req.Len(key, 44)
			}

			if test.args.validateDiffKey {
				newKey, err := plugin.CreateMasterKey(test.args.masterKeyPassphrase, test.args.localSecureConfigPath)
				checkError(err, test.wantErr, test.wantErrMsg, req)
				req.Len(newKey, 44)
				req.NotEqual(key, newKey)
			}
		})
	}
}

func TestPasswordProtectionSuite_EncryptConfigFileSecrets(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		setMEK                 bool
		createConfig           bool
		config                 string
		validateUsingDecrypt   bool
		outputConfigPath       string
		originalConfigs        string
	}
	tests := []struct {
		name            string
		args            *args
		wantErr         bool
		wantErrMsg      string
		wantSuggestions string
		wantConfigFile  string
		wantSecretsFile string
	}{
		{
			name: "InvalidTestCase: master key not set",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				setMEK:                 false,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr:         true,
			wantErrMsg:      fmt.Sprintf(errors.MasterKeyNotExportedErrorMsg, ConfluentKeyEnvVar),
			wantSuggestions: fmt.Sprintf(errors.MasterKeyNotExportedSuggestions, ConfluentKeyEnvVar),
		},
		{
			name: "InvalidTestCase: invalid config file path",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/encrypt/random.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				setMEK:                 true,
				createConfig:           false,
				validateUsingDecrypt:   false,
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.InvalidConfigFilePathErrorMsg, "/tmp/securePass987/encrypt/random.properties"),
		},
		{
			name: "ValidTestCase: encrypt config file with no config param, create new dek",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr: false,
			wantConfigFile: `testPassword = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `config.properties/testPassword = ENC[AES/GCM/NoPadding,data:`,
		},
		{
			name: "ValidTestCase: encrypt config file with last line as Comment, create new dek",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password\n# LAST LINE SHOULD NOT BE DELETED",
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr: false,
			wantConfigFile: `testPassword = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
# LAST LINE SHOULD NOT BE DELETED
`,
			wantSecretsFile: `config.properties/testPassword = ENC[AES/GCM/NoPadding,data:`,
		},
		{
			name: "ValidTestCase: encrypt config file with config param",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "ssl.keystore.password=password\nssl.keystore.location=/usr/ssl\nssl.keystore.key=ssl",
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "ssl.keystore.password",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr: false,
			wantConfigFile: `ssl.keystore.password = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/ssl.keystore.password}
ssl.keystore.location = /usr/ssl
ssl.keystore.key = ssl
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `config.properties/ssl.keystore.password = ENC[AES/GCM/NoPadding,data:`,
		},
		{
			name: "ValidTestCase: encrypt properties file with jaas entry",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `ssl.keystore.location=/usr/ssl
		ssl.keystore.key=ssl
		listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required \
          username="admin" \
          password="admin-secret";`,
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr: false,
			wantConfigFile: `ssl.keystore.location = /usr/ssl
ssl.keystore.key = ssl
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config = org.apache.kafka.common.security.scram.ScramLoginModule required username="admin" password=${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password};
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `config.properties/listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password = ENC[AES/GCM/NoPadding,data:`,
		},
		{
			name: "ValidTestCase: encrypt properties file with multiple jaas entry",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `ssl.keystore.location=/usr/ssl
		ssl.keystore.key=ssl
		listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required \
          username="admin" \
          password="admin-secret";`,
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/username, listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   true,
				outputConfigPath:       "/tmp/securePass987/encrypt/output.properties",
				originalConfigs:        "listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/username=\"admin\"\nlistener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password=\"admin-secret\"",
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: encrypt configuration in a JSON file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
"name": "security configuration",
"credentials": {
        "ssl.keystore.password": "password",
        "ssl.keystore.location": "/usr/ssl"
   }
}`,
				configFilePath:         "/tmp/securePass987/encrypt/config.json",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "credentials.ssl\\.keystore\\.password",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr: false,
			wantConfigFile: `{
  "config.providers.securepass.class": "io.confluent.kafka.security.config.provider.SecurePassConfigProvider",
  "config.providers": "securepass",
  "name": "security configuration",
  "credentials": {
    "ssl.keystore.password": "${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.json/credentials.ssl\\.keystore\\.password}",
    "ssl.keystore.location": "/usr/ssl"
  }
}
`,
			wantSecretsFile: `config.json/credentials.ssl\.keystore\.password = ENC[AES/GCM/NoPadding,data:`,
		},
		{
			name: "InvalidTestCase: encrypt invalid configuration in a JSON file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
"name": "security configuration",
"credentials": {
        "ssl.keystore.password": "password",
        "ssl.keystore.location": "/usr/ssl"
   }
}`,
				configFilePath:         "/tmp/securePass987/encrypt/config.json",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "credentials.ssl\\.trustore.\\location",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.ConfigKeyNotInJsonErrorMsg, "credentials.ssl\\.trustore.\\location"),
		},
		{
			name: "InvalidTestCase: encrypt configuration in invalid a JSON file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
"name": "security configuration",
"credentials": {
        "ssl.keystore.password": "password",
        "ssl.keystore.location": "/usr/ssl"
}`,
				configFilePath:         "/tmp/securePass987/encrypt/config.json",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "credentials.ssl\\.trustore.\\location",
				setMEK:                 true,
				createConfig:           true,
				validateUsingDecrypt:   false,
			},
			wantErr:    true,
			wantErrMsg: errors.InvalidJsonFileFormatErrorMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			err := os.MkdirAll(test.args.secureDir, os.ModePerm)
			defer os.RemoveAll(test.args.secureDir)
			req.NoError(err)

			plugin := NewPasswordProtectionPlugin()
			plugin.Clock = clockwork.NewFakeClock()
			if test.args.setMEK {
				err := createMasterKey(test.args.masterKeyPassphrase, test.args.localSecureConfigPath, plugin)
				defer os.Unsetenv(ConfluentKeyEnvVar)
				req.NoError(err)
			}
			if test.args.createConfig {
				err := createNewConfigFile(test.args.configFilePath, test.args.contents)
				req.NoError(err)
			}

			err = plugin.EncryptConfigFileSecrets(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, test.args.config)

			checkErrorAndSuggestions(err, test.wantErr, test.wantErrMsg, test.wantSuggestions, req)

			// Validate file contents for valid test cases
			if !test.wantErr {
				if test.args.validateUsingDecrypt {
					err = validateUsingDecryption(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, test.args.originalConfigs, plugin)
					req.NoError(err)
				} else {
					if strings.HasSuffix(test.args.configFilePath, ".json") {
						validateJSONFileContents(test.args.configFilePath, test.wantConfigFile, req)
					} else {
						validateTextFileContents(test.args.configFilePath, test.wantConfigFile, req)
					}
					validateTextFileContains(test.args.localSecureConfigPath, test.wantSecretsFile, req)
				}
			}
		})
	}
}

func TestPasswordProtectionSuite_DecryptConfigFileSecrets(t *testing.T) {
	type args struct {
		configFileContent      string
		secretFileContent      string
		masterKeyPassphrase    string
		configFilePath         string
		outputConfigPath       string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		newMasterKey           string
		setNewMEK              bool
	}
	tests := []struct {
		name           string
		args           *args
		wantErr        bool
		wantErrMsg     string
		wantOutputFile string
	}{
		{
			name: "InvalidTestCase: Different master key for decryption",
			args: &args{
				masterKeyPassphrase: "xyz233",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 2019-05-30 19:34:58.190796 -0700 PDT m=+13.357260342
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
			},
			wantErr:    true,
			wantErrMsg: errors.UnwrapDataKeyErrorMsg,
		},
		{
			name: "InvalidTestCase: Corrupted encrypted data",
			args: &args{
				masterKeyPassphrase: "123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = a3dxASgtO0kVRyAjajx/Hqs8xDgnBXZwoZtzrfceuCc=
_metadata.symmetric_key.0.created_at = 2022-02-14 12:00:36.756515 -0800 PST m=+0.042773556
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 7rpBEn1HaaqYu90AT/Kx2FSzYpw9fOOdfftwhy0rJrg=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:AYNIhDBhdWi0a1t6NQFUEKg2Y4GL7agPBjAmNCmY9qoedy7k0fYOqBWfOX0Rsf3Pya1+tY3vTy7HqUDR,iv:vZPrsbCbfk1iqZBT,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:asdsdsssddsoooofVXowRlNy9wP3Weq03Yrye8aPi8/5TZVb,iv:reOcPTUnq73SmAFA,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
				newMasterKey:           "SlWcsfDAvm5amIcVc65pfYGcNkD3wZpM7DDiFByhfh8=",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.DecryptConfigErrorMsg, "testPassword"),
		},
		{
			name: "InvalidTestCase: Corrupted DEK",
			args: &args{
				masterKeyPassphrase: "abc123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 2019-05-30 19:34:58.190796 -0700 PDT m=+13.357260342
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:svYxySZYksI8oDkF36ZYRze3q1CiqJQLwp+9jrfb0w1znLXOKgDlw/PKQMtvrCkCd,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt/",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              false,
				newMasterKey:           "xyz233",
			},
			wantErr:    true,
			wantErrMsg: errors.UnwrapDataKeyErrorMsg,
		},
		{
			name: "InvalidTestCase: Corrupted Data few characters interchanged",
			args: &args{
				masterKeyPassphrase: "123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = a3dxASgtO0kVRyAjajx/Hqs8xDgnBXZwoZtzrfceuCc=
_metadata.symmetric_key.0.created_at = 2022-02-14 12:00:36.756515 -0800 PST m=+0.042773556
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 7rpBEn1HaaqYu90AT/Kx2FSzYpw9fOOdfftwhy0rJrg=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:AYNIhDBhdWi0a1t6NQFUEKg2Y4GL7agPBjAmNCmY9qoedy7k0fYOqBWfOX0Rsf3Pya1+tY3vTy7HqUDR,iv:vZPrsbCbfk1iqZBT,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:lNyVXowR9wP3Weq03Yrye8aPi8/5TZVb,iv:reOcPTUnq73SmAFA,type:str
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt/",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
				newMasterKey:           "SlWcsfDAvm5amIcVc65pfYGcNkD3wZpM7DDiFByhfh8=",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.DecryptConfigErrorMsg, "testPassword"),
		},
		{
			name: "InvalidTestCase: Corrupted Data few characters removed",
			args: &args{
				masterKeyPassphrase: "123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = a3dxASgtO0kVRyAjajx/Hqs8xDgnBXZwoZtzrfceuCc=
_metadata.symmetric_key.0.created_at = 2022-02-14 12:00:36.756515 -0800 PST m=+0.042773556
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 7rpBEn1HaaqYu90AT/Kx2FSzYpw9fOOdfftwhy0rJrg=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:AYNIhDBhdWi0a1t6NQFUEKg2Y4GL7agPBjAmNCmY9qoedy7k0fYOqBWfOX0Rsf3Pya1+tY3vTy7HqUDR,iv:vZPrsbCbfk1iqZBT,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:lNy9wP3Weq03Yrye8aPi8/5TZVb,iv:reOcPTUnq73SmAFA,type:str
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt/",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
				newMasterKey:           "SlWcsfDAvm5amIcVc65pfYGcNkD3wZpM7DDiFByhfh8=",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.DecryptConfigErrorMsg, "testPassword"),
		},
		{
			name: "ValidTestCase: Decrypt Config File (AES CBC)",
			args: &args{
				masterKeyPassphrase: "123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 2019-05-30 19:34:58.190796 -0700 PDT m=+13.357260342
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:svYxySZYkI8oDkF36ZYRze3q1CiqJQLwp+9jrfb0w1znLXOKgDlw/PKQMtvrCkCd,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:zzjj9G+MeJ6XgsoIUFOVog==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				setNewMEK:              true,
				newMasterKey:           "YC7IvcB0J60YBytDhGLP+GlAQ2j7igE0kXIZ+VphUKA=",
			},
			wantErr:        false,
			wantOutputFile: "testPassword = password\n",
		},
		{
			name: "ValidTestCase: Decrypt Config File (AES GCM)",
			args: &args{
				masterKeyPassphrase: "123",
				configFileContent: `testPassword = ${securepass:/tmp/securePass987/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
				secretFileContent: `_metadata.master_key.0.salt = a3dxASgtO0kVRyAjajx/Hqs8xDgnBXZwoZtzrfceuCc=
_metadata.symmetric_key.0.created_at = 2022-02-14 12:00:36.756515 -0800 PST m=+0.042773556
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 10000
_metadata.symmetric_key.0.salt = 7rpBEn1HaaqYu90AT/Kx2FSzYpw9fOOdfftwhy0rJrg=
_metadata.symmetric_key.0.enc = ENC[AES/GCM/NoPadding,data:AYNIhDBhdWi0a1t6NQFUEKg2Y4GL7agPBjAmNCmY9qoedy7k0fYOqBWfOX0Rsf3Pya1+tY3vTy7HqUDR,iv:vZPrsbCbfk1iqZBT,type:str]
config.properties/testPassword = ENC[AES/GCM/NoPadding,data:VXowRlNy9wP3Weq03Yrye8aPi8/5TZVb,iv:reOcPTUnq73SmAFA,type:str
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				setNewMEK:              true,
				newMasterKey:           "SlWcsfDAvm5amIcVc65pfYGcNkD3wZpM7DDiFByhfh8=",
			},
			wantErr:        false,
			wantOutputFile: "testPassword = password\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, "")
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			// Create config file
			err = os.WriteFile(test.args.configFilePath, []byte(test.args.configFileContent), 0644)
			req.NoError(err)

			err = os.WriteFile(test.args.localSecureConfigPath, []byte(test.args.secretFileContent), 0644)
			req.NoError(err)

			if test.args.setNewMEK {
				os.Setenv(ConfluentKeyEnvVar, test.args.newMasterKey)
			}

			err = plugin.DecryptConfigFileSecrets(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, "")
			checkError(err, test.wantErr, test.wantErrMsg, req)

			if !test.wantErr {
				validateTextFileContents(test.args.outputConfigPath, test.wantOutputFile, req)
			}
		})
	}
}

func TestPasswordProtectionSuite_AddConfigFileSecrets(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		newConfigs             string
		outputConfigPath       string
		validateUsingDecrypt   bool
	}
	tests := []struct {
		name            string
		args            *args
		wantErr         bool
		wantErrMsg      string
		wantConfigFile  string
		wantSecretsFile string
	}{
		{
			name: "ValidTestCase: Add new configs",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/add/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/add/secureConfig.properties",
				secureDir:              "/tmp/securePass987/add",
				remoteSecureConfigPath: "/tmp/securePass987/add/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/add/output.properties",
				validateUsingDecrypt:   true,
				newConfigs:             "ssl.keystore.password = sslPass\ntruststore.keystore.password = keystorePass\n",
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Empty new configs",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/add/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/add/secureConfig.properties",
				secureDir:              "/tmp/securePass987/add",
				remoteSecureConfigPath: "/tmp/securePass987/add/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/add/output.properties",
				newConfigs:             "",
			},
			wantErr:    true,
			wantErrMsg: "add failed: empty list of new configs",
		},
		{
			name: "ValidTestCase: Add new config to JAAS config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
   useKeyTab=false \
   useTicketCache=true \
   doNotPrompt=true;`,
				configFilePath:         "/tmp/securePass987/add/embeddedjaas.properties",
				localSecureConfigPath:  "/tmp/securePass987/add/secureConfig.properties",
				secureDir:              "/tmp/securePass987/add",
				remoteSecureConfigPath: "/tmp/securePass987/add/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/add/output.properties",
				newConfigs:             "test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/password = testpassword\n",
				validateUsingDecrypt:   true,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: Add new config to JSON file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
"name": "security configuration",
"credentials": {
        "ssl.keystore.location": "/usr/ssl"
   }
}`,
				configFilePath:         "/tmp/securePass987/encrypt/config.json",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				newConfigs:             "credentials.password = password",
			},
			wantErr: false,
			wantConfigFile: `{
  "config.providers.securepass.class": "io.confluent.kafka.security.config.provider.SecurePassConfigProvider",
  "config.providers": "securepass",
  "name": "security configuration",
  "credentials": {
    "password": "${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.json/credentials.password}",
    "ssl.keystore.location": "/usr/ssl"
  }
}
`,
			wantSecretsFile: `config.json/credentials.password = ENC[AES/GCM/NoPadding,data:`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, test.args.contents)
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			err = plugin.AddEncryptedPasswords(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, test.args.newConfigs)
			checkError(err, test.wantErr, test.wantErrMsg, req)

			if !test.wantErr && test.args.validateUsingDecrypt {
				err = validateUsingDecryption(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, test.args.newConfigs, plugin)
				req.NoError(err)
			}

			if !test.wantErr && !test.args.validateUsingDecrypt {
				validateJSONFileContents(test.args.configFilePath, test.wantConfigFile, req)
				validateTextFileContains(test.args.localSecureConfigPath, test.wantSecretsFile, req)
			}
		})
	}
}

func TestPasswordProtectionSuite_UpdateConfigFileSecrets(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		outputConfigPath       string
		updateConfigs          string
		validateUsingDecrypt   bool
	}
	tests := []struct {
		name            string
		args            *args
		wantErr         bool
		wantErrMsg      string
		wantConfigFile  string
		wantSecretsFile string
	}{
		{
			name: "ValidTestCase: Update existing configs",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/update/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/update/secureConfig.properties",
				secureDir:              "/tmp/securePass987/update",
				remoteSecureConfigPath: "/tmp/securePass987/update/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/update/output.properties",
				updateConfigs:          "testPassword = newPassword\n",
				validateUsingDecrypt:   true,
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Key not present in config file",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/update/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/update/secureConfig.properties",
				secureDir:              "/tmp/securePass987/update",
				remoteSecureConfigPath: "/tmp/securePass987/update/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/update/output.properties",
				updateConfigs:          "ssl.keystore.password = newSslPass\ntestPassword = newPassword\n",
				validateUsingDecrypt:   true,
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.ConfigKeyNotPresentErrorMsg, "ssl.keystore.password"),
		},
		{
			name: "ValidTestCase: Update existing config in jaas config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
    useKeyTab=false \
    password=pass234 \
    useTicketCache=true \
    doNotPrompt=true;`,
				configFilePath:         "/tmp/securePass987/update/embeddedJaas.properties",
				localSecureConfigPath:  "/tmp/securePass987/update/secureConfig.properties",
				secureDir:              "/tmp/securePass987/update",
				remoteSecureConfigPath: "/tmp/securePass987/update/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/update/output.properties",
				updateConfigs:          "test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/password = newPassword\n",
				validateUsingDecrypt:   true,
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, test.args.contents)
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			err = plugin.UpdateEncryptedPasswords(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, test.args.updateConfigs)
			checkError(err, test.wantErr, test.wantErrMsg, req)

			if !test.wantErr && test.args.validateUsingDecrypt {
				err = validateUsingDecryption(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, test.args.updateConfigs, plugin)
				req.NoError(err)
			}

			if !test.wantErr && !test.args.validateUsingDecrypt {
				validateJSONFileContents(test.args.configFilePath, test.wantConfigFile, req)
				validateTextFileContains(test.args.localSecureConfigPath, test.wantSecretsFile, req)
			}
		})
	}
}

func TestPasswordProtectionSuite_RemoveConfigFileSecrets(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		outputConfigPath       string
		removeConfigs          string
		config                 string
	}
	tests := []struct {
		name       string
		args       *args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "ValidTestCase: Remove existing configs from properties file",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/remove/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/remove/output.properties",
				removeConfigs:          "testPassword",
				config:                 "",
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Key not present in config file",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/remove/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove/",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/remove/output.properties",
				removeConfigs:          "ssl.keystore.password",
				config:                 "",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.ConfigKeyNotEncryptedErrorMsg, "ssl.keystore.password"),
		},
		{
			name: "ValidTestCase:Remove existing configs from jaas config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
    useKeyTab=false \
    password=pass234 \
    useTicketCache=true \
    doNotPrompt=true;`,
				configFilePath:         "/tmp/securePass987/remove/embeddedJaas.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				removeConfigs:          "test.config.jaas",
				config:                 "test.config.jaas",
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase:Nested Key in jaas config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
    useKeyTab=false \
    password=pass234 \
    useTicketCache=true \
    doNotPrompt=true;`,
				configFilePath:         "/tmp/securePass987/remove/embeddedJaas.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				removeConfigs:          "test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/password",
				config:                 "",
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase:Key not present in jaas config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
    useKeyTab=false \
    password=pass234 \
    useTicketCache=true \
    doNotPrompt=true;`,
				configFilePath:         "/tmp/securePass987/remove/embeddedJaas.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				removeConfigs:          "test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/location",
				config:                 "",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.ConfigKeyNotEncryptedErrorMsg, "test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/location"),
		},
		{
			name: "ValidTestCase:Remove existing configs from json config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
			"name": "security configuration",
			"credentials": {
			"ssl.keystore.location": "/usr/ssl"
		}
		}`,
				configFilePath:         "/tmp/securePass987/remove/configuration.json",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				removeConfigs:          "credentials.ssl\\.keystore\\.location",
				config:                 "credentials.ssl\\.keystore\\.location",
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase:Key not present in json config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `{
			"name": "security configuration",
			"credentials": {
			"ssl.keystore.location": "/usr/ssl"
		}
		}`,
				configFilePath:         "/tmp/securePass987/remove/configuration.json",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				removeConfigs:          "credentials/location",
				config:                 "",
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf(errors.ConfigKeyNotEncryptedErrorMsg, "credentials/location"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, test.args.contents)
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, test.args.config)
			req.NoError(err)

			err = plugin.RemoveEncryptedPasswords(test.args.configFilePath, test.args.localSecureConfigPath, test.args.removeConfigs)
			checkError(err, test.wantErr, test.wantErrMsg, req)

			if !test.wantErr {
				// Verify passwords are removed
				err = verifyConfigsRemoved(test.args.configFilePath, test.args.localSecureConfigPath, test.args.removeConfigs)
				req.NoError(err)
			}
		})
	}
}

func TestPasswordProtectionSuite_RotateDataKey(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		outputConfigPath       string
		secureDir              string
		invalidPassphrase      string
		corruptDEK             bool
		invalidMEK             bool
	}
	tests := []struct {
		name       string
		args       *args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "ValidTestCase: Rotate dek",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotate/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotate/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotate",
				remoteSecureConfigPath: "/tmp/securePass987/rotate/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotate/output.properties",
				corruptDEK:             false,
				invalidMEK:             false,
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Rotate corrupted dek",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotate/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotate/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotate/",
				remoteSecureConfigPath: "/tmp/securePass987/rotate/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotate/output.properties",
				corruptDEK:             true,
				invalidMEK:             false,
			},
			wantErr:    true,
			wantErrMsg: errors.UnwrapDataKeyErrorMsg,
		},
		{
			name: "InvalidTestCase: Invalid master key",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotate/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotate/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotate/",
				remoteSecureConfigPath: "/tmp/securePass987/rotate/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotate/output.properties",
				corruptDEK:             false,
				invalidMEK:             true,
				invalidPassphrase:      "random",
			},
			wantErr:    true,
			wantErrMsg: errors.IncorrectPassphraseErrorMsg,
		},
		{
			name: "InvalidTestCase: Invalid master key special character space",
			args: &args{
				masterKeyPassphrase:    "abc123 ",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotate/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotate/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotate/",
				remoteSecureConfigPath: "/tmp/securePass987/rotate/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotate/output.properties",
				corruptDEK:             false,
				invalidMEK:             true,
				invalidPassphrase:      "abc123",
			},
			wantErr:    true,
			wantErrMsg: errors.IncorrectPassphraseErrorMsg,
		},
		{
			name: "InvalidTestCase: Invalid master key special character tab",
			args: &args{
				masterKeyPassphrase:    "abc123\t",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotate/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotate/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotate/",
				remoteSecureConfigPath: "/tmp/securePass987/rotate/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotate/output.properties",
				corruptDEK:             false,
				invalidMEK:             true,
				invalidPassphrase:      "abc123",
			},
			wantErr:    true,
			wantErrMsg: errors.IncorrectPassphraseErrorMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, test.args.contents)
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, "")

			req.NoError(err)
			originalProps, err := properties.LoadFile(test.args.localSecureConfigPath, properties.UTF8)
			req.NoError(err)
			if test.args.corruptDEK {
				err := corruptEncryptedDEK(test.args.localSecureConfigPath)
				req.NoError(err)
			}

			masterKey := test.args.masterKeyPassphrase
			if test.args.invalidMEK {
				masterKey = test.args.invalidPassphrase
			}
			err = plugin.RotateDataKey(masterKey, test.args.localSecureConfigPath)
			checkError(err, test.wantErr, test.wantErrMsg, req)

			// Verify the encrypted values are different
			if !test.wantErr {
				rotatedProps, err := properties.LoadFile(test.args.localSecureConfigPath, properties.UTF8)
				req.NoError(err)
				for key, value := range originalProps.Map() {
					if !strings.HasPrefix(key, MetadataPrefix) {
						cipher := rotatedProps.GetString(key, "")
						req.NotEqual(cipher, value)
					}
				}
				err = validateUsingDecryption(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, test.args.contents, plugin)
				req.NoError(err)
			}
		})
	}
}

func TestPasswordProtectionSuite_RotateMasterKey(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		newMasterKeyPassphrase string
		invalidKeyPassphrase   string
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		outputConfigPath       string
		secureDir              string
		invalidMEK             bool
	}
	tests := []struct {
		name       string
		args       *args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "ValidTestCase: Rotate MEK",
			args: &args{
				masterKeyPassphrase:    "abc123",
				newMasterKeyPassphrase: "xyz987",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: Rotate MEK with special character master key",
			args: &args{
				masterKeyPassphrase:    "abc123 ",
				newMasterKeyPassphrase: "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             false,
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Empty master key passphrase",
			args: &args{
				masterKeyPassphrase:    "abc123",
				newMasterKeyPassphrase: "",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             false,
			},
			wantErr:    true,
			wantErrMsg: errors.EmptyPassphraseErrorMsg,
		},
		{
			name: "InvalidTestCase: Incorrect old master key passphrase",
			args: &args{
				masterKeyPassphrase:    "abc123",
				invalidKeyPassphrase:   "xyz456",
				newMasterKeyPassphrase: "mnt456",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             true,
			},
			wantErr:    true,
			wantErrMsg: errors.IncorrectPassphraseErrorMsg,
		},
		{
			name: "InvalidTestCase: Incorrect old master key passphrase with special char space",
			args: &args{
				masterKeyPassphrase:    "abc123 ",
				invalidKeyPassphrase:   "abc123",
				newMasterKeyPassphrase: "mnt456",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             true,
			},
			wantErr:    true,
			wantErrMsg: errors.IncorrectPassphraseErrorMsg,
		},
		{
			name: "InvalidTestCase: New master key passphrase same as old master key passphrase",
			args: &args{
				masterKeyPassphrase:    "abc123",
				newMasterKeyPassphrase: "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek/",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:             false,
			},
			wantErr:    true,
			wantErrMsg: errors.SamePassphraseErrorMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			plugin, err := setUpDir(test.args.masterKeyPassphrase, test.args.secureDir, test.args.configFilePath, test.args.localSecureConfigPath, test.args.contents)
			defer os.RemoveAll(test.args.secureDir)
			defer os.Unsetenv(ConfluentKeyEnvVar)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(test.args.configFilePath, test.args.localSecureConfigPath, test.args.remoteSecureConfigPath, "")
			req.NoError(err)

			masterKey := test.args.masterKeyPassphrase
			if test.args.invalidMEK {
				masterKey = test.args.invalidKeyPassphrase
			}
			newKey, err := plugin.RotateMasterKey(masterKey, test.args.newMasterKeyPassphrase, test.args.localSecureConfigPath)
			checkError(err, test.wantErr, test.wantErrMsg, req)

			if !test.wantErr {
				os.Setenv(ConfluentKeyEnvVar, newKey)
				err = validateUsingDecryption(test.args.configFilePath, test.args.localSecureConfigPath, test.args.outputConfigPath, test.args.contents, plugin)
				req.NoError(err)
			}
		})
	}
}

func createMasterKey(passphrase, localSecretsFile string, plugin *PasswordProtectionSuite) error {
	key, err := plugin.CreateMasterKey(passphrase, localSecretsFile)
	if err != nil {
		return err
	}
	os.Setenv(ConfluentKeyEnvVar, key)
	return nil
}

func createNewConfigFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0644)
}

func validateTextFileContents(path, expectedFileContent string, req *require.Assertions) {
	readContent, err := os.ReadFile(path)
	req.NoError(err)
	req.Equal(expectedFileContent, string(readContent))
}

func validateTextFileContains(path, expectedFileContent string, req *require.Assertions) {
	readContent, err := os.ReadFile(path)
	req.NoError(err)
	req.Contains(string(readContent), expectedFileContent)
}

func validateJSONFileContents(path, expectedFileContent string, req *require.Assertions) {
	readContent, err := os.ReadFile(path)
	req.NoError(err)
	req.JSONEq(expectedFileContent, string(readContent))
}

func generateCorruptedData(cipher string) (string, error) {
	data, _, _ := ParseCipherValue(cipher)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	corruptedData := base32.StdEncoding.EncodeToString(randomBytes)[:32]
	return strings.Replace(cipher, data, corruptedData, 1), nil
}

func corruptEncryptedDEK(localSecureConfigPath string) error {
	secretsProps, err := utils.LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	value := secretsProps.GetString(MetadataDataKey, "")
	corruptedCipher, err := generateCorruptedData(value)
	if err != nil {
		return err
	}
	if _, _, err := secretsProps.Set(MetadataDataKey, corruptedCipher); err != nil {
		return err
	}

	return WritePropertiesFile(localSecureConfigPath, secretsProps, true)
}

func verifyConfigsRemoved(configFilePath, localSecureConfigPath, removedConfigs string) error {
	secretsProps, err := utils.LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	configs := strings.Split(removedConfigs, ",")
	_, err = LoadConfiguration(configFilePath, configs, true)
	// Check if config is removed from configs files
	if err == nil {
		return fmt.Errorf("failed to remove config from config file")
	}
	for _, key := range configs {
		pathKey := GenerateConfigKey(configFilePath, key)

		// Check if config is removed from secrets files
		if _, ok := secretsProps.Get(pathKey); ok {
			return fmt.Errorf("failed to remove config from secrets file")
		}
	}

	return nil
}

func validateUsingDecryption(configFilePath, localSecureConfigPath, outputConfigPath, origConfigs string, plugin *PasswordProtectionSuite) error {
	if err := plugin.DecryptConfigFileSecrets(configFilePath, localSecureConfigPath, outputConfigPath, ""); err != nil {
		return fmt.Errorf("failed to decrypt config file")
	}

	decryptContent, err := os.ReadFile(outputConfigPath)
	if err != nil {
		return err
	}
	decryptContentStr := string(decryptContent)
	decryptConfigProps, err := properties.LoadString(decryptContentStr)
	if err != nil {
		return err
	}
	originalConfigProps, err := properties.LoadString(origConfigs)
	if err != nil {
		return err
	}
	originalConfigProps.DisableExpansion = true
	for key, value := range decryptConfigProps.Map() {
		originalVal, _ := originalConfigProps.Get(key)
		if value != originalVal {
			return fmt.Errorf("config file is empty")
		}
	}

	return nil
}

func setUpDir(masterKeyPassphrase, secureDir, configFile, localSecureConfigPath, contents string) (*PasswordProtectionSuite, error) {
	if err := os.MkdirAll(secureDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create password protection directory")
	}
	plugin := NewPasswordProtectionPlugin()
	plugin.Clock = clockwork.NewFakeClock()

	// Set master key
	if err := createMasterKey(masterKeyPassphrase, localSecureConfigPath, plugin); err != nil {
		return nil, fmt.Errorf("failed to create master key")
	}

	if err := createNewConfigFile(configFile, contents); err != nil {
		return nil, fmt.Errorf("failed to create config file")
	}

	return plugin, nil
}

func checkError(err error, wantErr bool, wantErrMsg string, req *require.Assertions) {
	if wantErr {
		req.Error(err)
		req.Contains(err.Error(), wantErrMsg)
	} else {
		req.NoError(err)
	}
}

func checkErrorAndSuggestions(err error, wantErr bool, wantErrMsg, wantSuggestions string, req *require.Assertions) {
	if wantErr {
		req.Error(err)
		req.Contains(err.Error(), wantErrMsg)
		if wantSuggestions != "" {
			errors.VerifyErrorAndSuggestions(req, err, wantErrMsg, wantSuggestions)
		}
	} else {
		req.NoError(err)
	}
}
