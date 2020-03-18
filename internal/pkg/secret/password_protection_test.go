package secret

import (
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/confluentinc/properties"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func TestPasswordProtectionSuite_CreateMasterKey(t *testing.T) {
	type args struct {
		masterKeyPassphrase          string
		localSecureConfigPath        string
		passphraseWithoutSpecialChar string
		validateSpecialChar          bool
		validateDiffKey              bool
		secureDir                    string
		newSeed                      int64
		seed                         int64
	}
	tests := []struct {
		name                      string
		args                      *args
		wantErr                   bool
		wantErrMsg                string
		wantMasterKey             string
		wantMEKNewSeed            string
		wantMEKWithoutSpecialChar string
		wantEqual                 bool
	}{
		{
			name: "ValidTestCase: valid create master key",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
				validateSpecialChar:   false,
				seed:                  99,
			},
			wantErr:       false,
			wantMasterKey: "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
		},
		{
			name: "ValidTestCase: valid create master key with space at the end",
			args: &args{
				secureDir:                    "/tmp/securePass987/create",
				masterKeyPassphrase:          "abc123 ",
				passphraseWithoutSpecialChar: "abc123",
				localSecureConfigPath:        "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:              false,
				validateSpecialChar:          true,
				seed:                         99,
			},
			wantErr:                   false,
			wantEqual:                 false,
			wantMasterKey:             "G0WWpceOnaCwbQSpfrHt94SRymEAt01dpTN9IRW4fxw=",
			wantMEKWithoutSpecialChar: "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
		},
		{
			name: "ValidTestCase: valid create master key with tab at the end",
			args: &args{
				secureDir:                    "/tmp/securePass987/create",
				masterKeyPassphrase:          "abc123\t",
				passphraseWithoutSpecialChar: "abc123",
				localSecureConfigPath:        "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:              false,
				validateSpecialChar:          true,
				seed:                         99,
			},
			wantErr:                   false,
			wantEqual:                 false,
			wantMasterKey:             "vmWub/JptUEihqjgzC+5x8Y0NSeqcVqraNRDV7opmLI=",
			wantMEKWithoutSpecialChar: "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
		},
		{
			name: "ValidTestCase: valid create master key with new line at the end",
			args: &args{
				secureDir:                    "/tmp/securePass987/create",
				masterKeyPassphrase:          "abc123\n",
				passphraseWithoutSpecialChar: "abc123",
				localSecureConfigPath:        "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:              false,
				validateSpecialChar:          true,
				seed:                         99,
			},
			wantErr:                   false,
			wantEqual:                 true,
			wantMasterKey:             "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
			wantMEKWithoutSpecialChar: "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
		},
		{
			name: "ValidTestCase: verify for same passphrase it generates a different master key",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       true,
				validateSpecialChar:   false,
				seed:                  99,
				newSeed:               10,
			},
			wantErr:        false,
			wantMasterKey:  "XWiYpuA2A6fG/gaweaHlr4So/ZHz2swjgV1QT2mf/sM=",
			wantMEKNewSeed: "jeBtSH5mR4AZc9KiBEt/uVyrx9vJ2WPKdD3g1YhW2kI=",
		},
		{
			name: "InvalidTestCase: empty passphrase",
			args: &args{
				secureDir:             "/tmp/securePass987/create",
				masterKeyPassphrase:   "",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey:       false,
				validateSpecialChar:   false,
				seed:                  99,
			},
			wantErr:    true,
			wantErrMsg: "master key passphrase cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			logger := log.New()
			err := os.MkdirAll(tt.args.secureDir, os.ModePerm)
			req.NoError(err)

			plugin := NewPasswordProtectionPlugin(logger)
			plugin.RandSource = rand.NewSource(tt.args.seed)

			key, err := plugin.CreateMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)
			if !tt.wantErr {
				req.Equal(key, tt.wantMasterKey)
			}

			if tt.args.validateDiffKey {
				plugin.RandSource = rand.NewSource(tt.args.newSeed)
				newKey, err := plugin.CreateMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath)
				req.Equal(newKey, tt.wantMEKNewSeed)
				checkError(err, tt.wantErr, tt.wantErrMsg, req)
				req.NotEqual(key, newKey)
			}

			if tt.args.validateSpecialChar {
				plugin.RandSource = rand.NewSource(tt.args.seed)
				newKey, err := plugin.CreateMasterKey(tt.args.passphraseWithoutSpecialChar, tt.args.localSecureConfigPath)
				req.Equal(newKey, tt.wantMEKWithoutSpecialChar)
				checkError(err, tt.wantErr, tt.wantErrMsg, req)
				if tt.wantEqual {
					req.Equal(key, newKey)
				} else {
					req.NotEqual(key, newKey)
				}
			}

			os.RemoveAll(tt.args.secureDir)
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
			},
			wantErr:    true,
			wantErrMsg: "master key is not exported in CONFLUENT_SECURITY_MASTER_KEY environment variable; export the key and execute this command again",
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
			},
			wantErr:    true,
			wantErrMsg: "invalid config file path: /tmp/securePass987/encrypt/random.properties",
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
			},
			wantErr: false,
			wantConfigFile: `testPassword = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
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
			},
			wantErr: false,
			wantConfigFile: `testPassword = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/testPassword}
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
# LAST LINE SHOULD NOT BE DELETED
`,
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
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
			},
			wantErr: false,
			wantConfigFile: `ssl.keystore.password = ${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/ssl.keystore.password}
ssl.keystore.location = /usr/ssl
ssl.keystore.key = ssl
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/ssl.keystore.password = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
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
			},
			wantErr: false,
			wantConfigFile: `ssl.keystore.location = /usr/ssl
ssl.keystore.key = ssl
listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config = org.apache.kafka.common.security.scram.ScramLoginModule required username="admin" password=${securepass:/tmp/securePass987/encrypt/secureConfig.properties:config.properties/listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password};
config.providers = securepass
config.providers.securepass.class = io.confluent.kafka.security.config.provider.SecurePassConfigProvider
`,
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/listener.name.sasl_ssl.scram-sha-256.sasl.jaas.config/org.apache.kafka.common.security.scram.ScramLoginModule/password = ENC[AES/CBC/PKCS5Padding,data:6etDBw0weeD4UQF664szSQ==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
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
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.json/credentials.ssl\.keystore\.password = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
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
			},
			wantErr:    true,
			wantErrMsg: "Configuration key credentials.ssl\\.trustore.\\location is not present in JSON configuration file.",
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
			},
			wantErr:    true,
			wantErrMsg: "Invalid json file format.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
			logger := log.New()
			req := require.New(t)
			err := os.MkdirAll(tt.args.secureDir, os.ModePerm)
			req.NoError(err)
			plugin := NewPasswordProtectionPlugin(logger)
			plugin.RandSource = rand.NewSource(99)
			plugin.Clock = clockwork.NewFakeClock()
			if tt.args.setMEK {
				err := createMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath, plugin)
				req.NoError(err)
			}
			if tt.args.createConfig {
				err := createNewConfigFile(tt.args.configFilePath, tt.args.contents)
				req.NoError(err)
			}

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.config)

			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			// Validate file contents for valid test cases
			if !tt.wantErr {
				validateFileContents(tt.args.configFilePath, tt.wantConfigFile, req)
				validateFileContents(tt.args.localSecureConfigPath, tt.wantSecretsFile, req)
			}

			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
			},
			wantErr:    true,
			wantErrMsg: "failed to unwrap the data key due to invalid master key or corrupted data key.",
		},
		{
			name: "InvalidTestCase: Corrupted encrypted data",
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:asdsdsssddsoooofsccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              false,
				newMasterKey:           "xyz233",
			},
			wantErr:    true,
			wantErrMsg: "failed to decrypt config testPassword due to corrupted data.",
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkdddswwsassddccaaaQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
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
			wantErrMsg: "failed to unwrap the data key due to invalid master key or corrupted data key.",
		},
		{
			name: "InvalidTestCase: Corrupted Data few characters interchanged",
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:lcSgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
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
			wantErrMsg: "failed to decrypt config testPassword due to corrupted data.",
		},
		{
			name: "InvalidTestCase: Corrupted Data few characters removed",
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
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
			wantErrMsg: "failed to decrypt config testPassword due to corrupted data.",
		},
		{
			name: "ValidTestCase: Decrypt Config File",
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
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.properties/testPassword = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
			},
			wantErr:        false,
			wantOutputFile: "testPassword = password\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, "")
			req.NoError(err)

			// Create config file
			err = ioutil.WriteFile(tt.args.configFilePath, []byte(tt.args.configFileContent), 0644)
			req.NoError(err)

			err = ioutil.WriteFile(tt.args.localSecureConfigPath, []byte(tt.args.secretFileContent), 0644)
			req.NoError(err)

			if tt.args.setNewMEK {
				os.Setenv(CONFLUENT_KEY_ENVVAR, tt.args.newMasterKey)
			}

			err = plugin.DecryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, "")
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			if !tt.wantErr {
				validateFileContents(tt.args.outputConfigPath, tt.wantOutputFile, req)
			}

			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
			wantSecretsFile: `_metadata.master_key.0.salt = de0YQknpvBlnXk0fdmIT2nG2Qnj+0srV8YokdhkgXjA=
_metadata.symmetric_key.0.created_at = 1984-04-04 00:00:00 +0000 UTC
_metadata.symmetric_key.0.envvar = CONFLUENT_SECURITY_MASTER_KEY
_metadata.symmetric_key.0.length = 32
_metadata.symmetric_key.0.iterations = 1000
_metadata.symmetric_key.0.salt = 2BEkhLYyr0iZ2wI5xxsbTJHKWul75JcuQu3BnIO4Eyw=
_metadata.symmetric_key.0.enc = ENC[AES/CBC/PKCS5Padding,data:SlpCTPDO/uyWDOS59hkcS9vTKm2MQ284YQhBM2iFSUXgsDGPBIlYBs4BMeWFt1yn,iv:qDtNy+skN3DKhtHE/XD6yQ==,type:str]
config.json/credentials.password = ENC[AES/CBC/PKCS5Padding,data:SclgTBDDeLwccqtsaEmDlA==,iv:3IhIyRrhQpYzp4vhVdcqqw==,type:str]
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
			req := require.New(t)
			// SetUp

			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.AddEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.newConfigs)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			if !tt.wantErr && tt.args.validateUsingDecrypt {
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.newConfigs, plugin)
				req.NoError(err)
			}

			if !tt.wantErr && !tt.args.validateUsingDecrypt {
				validateFileContents(tt.args.configFilePath, tt.wantConfigFile, req)
				validateFileContents(tt.args.localSecureConfigPath, tt.wantSecretsFile, req)
			}

			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
			wantErrMsg: "Configuration key ssl.keystore.password is not present in the configuration file.",
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.UpdateEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.updateConfigs)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			if !tt.wantErr && tt.args.validateUsingDecrypt {
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.updateConfigs, plugin)
				req.NoError(err)
			}

			if !tt.wantErr && !tt.args.validateUsingDecrypt {
				validateFileContents(tt.args.configFilePath, tt.wantConfigFile, req)
				validateFileContents(tt.args.localSecureConfigPath, tt.wantSecretsFile, req)
			}
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
			wantErrMsg: "Configuration key ssl.keystore.password is not present in the configuration file.",
		},
		{
			name: "ValidTestCase:Remove existing configs from jaas config file",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: `test.config.jaas = com.sun.security.auth.module.Krb5LoginModule required \
    useKeyTab=false \
    password=pass234 \
    useTicketCache=true \
    password=testPass \
    doNotPrompt=true;
};`,
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
			wantErrMsg: "Configuration key test.config.jaas/com.sun.security.auth.module.Krb5LoginModule/location is not present in the configuration file.",
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
			wantErrMsg: "Configuration key credentials/location is not present in JSON configuration file.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
			// SetUp
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.config)
			req.NoError(err)

			err = plugin.RemoveEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.removeConfigs)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			if !tt.wantErr {
				// Verify passwords are removed
				err = verifyConfigsRemoved(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.removeConfigs)
				req.NoError(err)
			}
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
			wantErrMsg: "failed to unwrap the data key due to invalid master key or corrupted data key.",
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
			wantErrMsg: "authentication failure: incorrect master key passphrase.",
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
			wantErrMsg: "authentication failure: incorrect master key passphrase.",
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
			wantErrMsg: "authentication failure: incorrect master key passphrase.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, "")

			req.NoError(err)
			originalProps, err := properties.LoadFile(tt.args.localSecureConfigPath, properties.UTF8)
			req.NoError(err)
			if tt.args.corruptDEK {
				err := corruptEncryptedDEK(tt.args.localSecureConfigPath)
				req.NoError(err)
			}

			masterKey := tt.args.masterKeyPassphrase
			if tt.args.invalidMEK {
				masterKey = tt.args.invalidPassphrase
			}
			err = plugin.RotateDataKey(masterKey, tt.args.localSecureConfigPath)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			// Verify the encrypted values are different
			if !tt.wantErr {
				rotatedProps, err := properties.LoadFile(tt.args.localSecureConfigPath, properties.UTF8)
				req.NoError(err)
				for key, value := range originalProps.Map() {
					if !strings.HasPrefix(key, METADATA_PREFIX) {
						cipher := rotatedProps.GetString(key, "")
						req.NotEqual(cipher, value)
					}
				}
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)
				req.NoError(err)
			}
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
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
			wantErrMsg: "master key passphrase cannot be empty.",
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
			wantErrMsg: "authentication failure: incorrect master key passphrase.",
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
			wantErrMsg: "authentication failure: incorrect master key passphrase.",
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
			wantErrMsg: "new master key passphrase may not be the same as the previous passphrase.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, "")
			req.NoError(err)

			masterKey := tt.args.masterKeyPassphrase
			if tt.args.invalidMEK {
				masterKey = tt.args.invalidKeyPassphrase
			}
			newKey, err := plugin.RotateMasterKey(masterKey, tt.args.newMasterKeyPassphrase, tt.args.localSecureConfigPath)
			checkError(err, tt.wantErr, tt.wantErrMsg, req)

			if !tt.wantErr {
				os.Setenv(CONFLUENT_KEY_ENVVAR, newKey)
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)
				req.NoError(err)
			}
			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func createMasterKey(passphrase string, localSecretsFile string, plugin *PasswordProtectionSuite) error {
	key, err := plugin.CreateMasterKey(passphrase, localSecretsFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	os.Setenv(CONFLUENT_KEY_ENVVAR, key)
	return nil
}

func createNewConfigFile(path string, contents string) error {
	err := ioutil.WriteFile(path, []byte(contents), 0644)
	return err
}

func validateFileContents(path string, expectedFileContent string, req *require.Assertions) {
	readContent, err := ioutil.ReadFile(path)
	req.NoError(err)
	req.Equal(expectedFileContent, string(readContent))
}

func generateCorruptedData(cipher string) (string, error) {
	data, _, _ := ParseCipherValue(cipher)
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	corruptedData := base32.StdEncoding.EncodeToString(randomBytes)[:32]
	result := strings.Replace(cipher, data, corruptedData, 1)
	return result, nil
}

func corruptEncryptedDEK(localSecureConfigPath string) error {
	secretsProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	value := secretsProps.GetString(METADATA_DATA_KEY, "")
	corruptedCipher, err := generateCorruptedData(value)
	if err != nil {
		return err
	}
	_, _, err = secretsProps.Set(METADATA_DATA_KEY, corruptedCipher)
	if err != nil {
		return err
	}

	err = WritePropertiesFile(localSecureConfigPath, secretsProps, true)
	return err
}

func verifyConfigsRemoved(configFilePath string, localSecureConfigPath string, removedConfigs string) error {
	secretsProps, err := LoadPropertiesFile(localSecureConfigPath)
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
		_, ok := secretsProps.Get(pathKey)
		if ok {
			return fmt.Errorf("failed to remove config from secrets file")
		}
	}

	return nil
}

func validateUsingDecryption(configFilePath string, localSecureConfigPath string, outputConfigPath string, origConfigs string, plugin *PasswordProtectionSuite) error {
	err := plugin.DecryptConfigFileSecrets(configFilePath, localSecureConfigPath, outputConfigPath, "")
	if err != nil {
		return fmt.Errorf("failed to decrypt config file")
	}

	decryptContent, err := ioutil.ReadFile(outputConfigPath)
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

func setUpDir(masterKeyPassphrase string, secureDir string, configFile string, localSecureConfigPath string, contents string) (*PasswordProtectionSuite, error) {
	err := os.MkdirAll(secureDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create password protection directory")
	}
	logger := log.New()
	plugin := NewPasswordProtectionPlugin(logger)
	plugin.RandSource = rand.NewSource(99)
	plugin.Clock = clockwork.NewFakeClock()

	// Set master key
	err = createMasterKey(masterKeyPassphrase, localSecureConfigPath, plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key")
	}

	err = createNewConfigFile(configFile, contents)
	if err != nil {
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
