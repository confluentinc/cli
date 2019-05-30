package secret

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/magiconair/properties"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestPasswordProtectionSuite_CreateMasterKey(t *testing.T) {
	type args struct {
		masterKeyPassphrase   string
		localSecureConfigPath string
		validateDiffKey bool
		secureDir             string
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: valid create master key",
			args: &args{
				secureDir: "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey: false,
			},
			wantErr: false,
		},
		{
			name: "ValidTestCase: verify for same passphrase it generates a different master key",
			args: &args{
				secureDir: "/tmp/securePass987/create",
				masterKeyPassphrase:   "abc123",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey: true,
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: empty passphrase",
			args: &args{
				secureDir: "/tmp/securePass987/create",
				masterKeyPassphrase:   "",
				localSecureConfigPath: "/tmp/securePass987/create/secureConfig.properties",
				validateDiffKey: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			logger := log.New()
			err := os.MkdirAll(tt.args.secureDir, os.ModePerm)
			req.NoError(err)

			plugin := NewPasswordProtectionPlugin(logger)
			key, err := plugin.CreateMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath)
			checkError(err, tt.wantErr, req)
			if !tt.wantErr {
				req.NotEqual(key, "")
			}

			if tt.args.validateDiffKey {
				newKey, err := plugin.CreateMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath)
				checkError(err, tt.wantErr, req)
				req.NotEqual(key, newKey)
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
		configVal              string
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
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
				configVal:              "",
				setMEK:                 false,
				createConfig:           true,
			},
			wantErr: true,
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
				configVal:              "",
				setMEK:                 true,
				createConfig:           false,
			},
			wantErr: true,
		},
		{
			name: "ValidTestCase: encrypt config file with no config param, create new dek",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password\ntestKey=key",
				configFilePath:         "/tmp/securePass987/encrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/encrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/encrypt",
				remoteSecureConfigPath: "/tmp/securePass987/encrypt/secureConfig.properties",
				config:                 "",
				configVal:              "",
				setMEK:                 true,
				createConfig:           true,
			},
			wantErr: false,
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
				config:                 "ssl.keystore.password,ssl.keystore.location",
				configVal:              "ssl.keystore.password=password\nssl.keystore.location=/usr/ssl",
				setMEK:                 true,
				createConfig:           true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			req := require.New(t)
			err := os.MkdirAll(tt.args.secureDir, os.ModePerm)
			req.NoError(err)
			plugin := NewPasswordProtectionPlugin(logger)
			if tt.args.setMEK {
				err := createMasterKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath, plugin)
				req.NoError(err)
			}
			if tt.args.createConfig {
				err := createNewConfigFile(tt.args.configFilePath, tt.args.contents)
				req.NoError(err)
			}

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.config)

			checkError(err, tt.wantErr, req)

			// Validate file contents for valid test cases
			if !tt.wantErr {
				err := validateFileContents(tt.args.contents, tt.args.configFilePath, tt.args.remoteSecureConfigPath, tt.args.localSecureConfigPath, plugin, tt.args.configVal)
				req.NoError(err)
			}

			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func TestPasswordProtectionSuite_DecryptConfigFileSecrets(t *testing.T) {
	type args struct {
		contents               string
		masterKeyPassphrase    string
		configFilePath         string
		outputConfigPath       string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		secureDir              string
		newMasterKey           string
		setNewMEK              bool
		corruptData            bool
		corruptDEK             bool
		corruptFewBytes        bool

	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "InvalidTestCase: Different master key for decryption",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              true,
				corruptData:            false,
				corruptDEK:             false,
				newMasterKey:           "xyz233",
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Corrupted encrypted data",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              false,
				corruptData:            true,
				corruptDEK:             false,
				newMasterKey:           "xyz233",
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Corrupted DEK",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt/",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              false,
				corruptData:            false,
				corruptDEK:             true,
				newMasterKey:           "xyz233",
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Corrupted Data few characters interchanged",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/decrypt/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/decrypt/secureConfig.properties",
				secureDir:              "/tmp/securePass987/decrypt/",
				remoteSecureConfigPath: "/tmp/securePass987/decrypt/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/decrypt/output.properties",
				setNewMEK:              false,
				corruptData:            false,
				corruptDEK:             true,
				newMasterKey:           "xyz233",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, "")
			req.NoError(err)

			if tt.args.corruptData {
				err := corruptEncryptedData(tt.args.localSecureConfigPath)
				req.NoError(err)
			}

			if tt.args.corruptDEK {
				err := corruptEncryptedDEK(tt.args.localSecureConfigPath)
				req.NoError(err)
			}

			if tt.args.corruptFewBytes {
				err := corruptEncryptedDEK(tt.args.localSecureConfigPath)
				req.NoError(err)
			}

			if tt.args.setNewMEK {
				os.Setenv(CONFLUENT_KEY_ENVVAR, tt.args.newMasterKey)
			}

			err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)
			checkError(err, tt.wantErr, req)

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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
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
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// SetUp
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.AddEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.newConfigs)
			checkError(err, tt.wantErr, req)

			if !tt.wantErr {
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.newConfigs, plugin)
				req.NoError(err)
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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
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
				updateConfigs:          "ssl.keystore.password = newSslPass\ntruststore.keystore.password = newKeystorePass\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.UpdateEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.updateConfigs)
			checkError(err, tt.wantErr, req)

			if !tt.wantErr {
				// Verify passwords are added
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.updateConfigs, plugin)
				req.NoError(err)
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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: Remove existing configs",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/remove/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/remove/secureConfig.properties",
				secureDir:              "/tmp/securePass987/remove",
				remoteSecureConfigPath: "/tmp/securePass987/remove/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/remove/output.properties",
				removeConfigs:          "testPassword",
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
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// SetUp
			plugin, err := setUpDir(tt.args.masterKeyPassphrase, tt.args.secureDir, tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.contents)
			req.NoError(err)

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, "")
			req.NoError(err)

			err = plugin.RemoveEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.removeConfigs)
			checkError(err, tt.wantErr, req)

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
		name    string
		args    *args
		wantErr bool
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
			wantErr: true,
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
			wantErr: true,
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
			checkError(err, tt.wantErr, req)

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
		name    string
		args    *args
		wantErr bool
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
				invalidMEK:              false,
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
				invalidMEK:              false,
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Incorrect old master key passphrase",
			args: &args{
				masterKeyPassphrase:    "abc123",
				invalidKeyPassphrase:   "xyz456",
				newMasterKeyPassphrase: "",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/rotateMek/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/rotateMek/secureConfig.properties",
				secureDir:              "/tmp/securePass987/rotateMek",
				remoteSecureConfigPath: "/tmp/securePass987/rotateMek/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/rotateMek/output.properties",
				invalidMEK:              false,
			},
			wantErr: true,
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
				invalidMEK:              false,
			},
			wantErr: true,
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
			checkError(err, tt.wantErr, req)

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
		fmt.Println( err)
		return err
	}
	os.Setenv(CONFLUENT_KEY_ENVVAR, key)
	return nil
}

func createNewConfigFile(path string, contents string) error {
	err := ioutil.WriteFile(path, []byte(contents), 0644)
	return err
}

func validateFileContents(contents string, configFile string, remoteSecretsFile string, localSecretsFile string, plugin *PasswordProtectionSuite, config string) error {
	originalConfigs, err := properties.LoadString(contents)
	if err != nil {
		return err
	}
	originalConfigs.DisableExpansion = true
	encryptedConfigs, err := properties.LoadString(config)
	if err != nil {
		return err
	}
	encryptedConfigs.DisableExpansion = true
	// Load the configs.
	configProps, err := LoadPropertiesFile(configFile)
	if err != nil {
		return err
	}

	secretsProps, err := LoadPropertiesFile(localSecretsFile)
	if err != nil {
		return err
	}

	for key, value := range configProps.Map() {
		_, ok := encryptedConfigs.Get(key)
		if (config == "" && strings.Contains(strings.ToLower(key), "password")) || ok {
			// Validate the config value in config file
			pathKey := GenerateConfigKey(configFile, key)
			expectedVal := GenerateConfigValue(pathKey, remoteSecretsFile)
			if strings.Compare(expectedVal, value) != 0 {
				return fmt.Errorf("failed to encrypt a secret config")
			}

			// Validate the secrets value in secret file
			secretsVal, ok := secretsProps.Get(pathKey)
			if !ok {
				return fmt.Errorf("secrets config does not contain encrypted secret for key " + key)
			}

			data, iv, algo := ParseCipherValue(secretsVal)

			if len(strings.TrimSpace(data)) == 0 || len(strings.TrimSpace(iv)) == 0 || len(strings.TrimSpace(algo)) == 0 {
				return fmt.Errorf("secrets config value not in correct format.")
			}

		} else if !strings.Contains(strings.ToLower(key), "config.provider") {
			// Validate non secret configs are not modified
			originalVal, _ := originalConfigs.Get(key)
			if strings.Compare(originalVal, value) != 0 {
				return fmt.Errorf("illegal operation: non secret config modified!!!")
			}
		}
	}
	return nil
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

func corruptEncryptedData(localSecureConfigPath string) error {
	secretsProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	for key, value := range secretsProps.Map() {
		if !strings.HasPrefix(key, METADATA_PREFIX) {
			corruptedCipher, err := generateCorruptedData(value)
			if err != nil {
				return err
			}
			_, _, err = secretsProps.Set(key, corruptedCipher)
			if err != nil {
				return err
			}
		}
	}

	err = WritePropertiesFile(localSecureConfigPath, secretsProps, true)
	return err
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
	configProps, err := LoadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}
	configs := strings.Split(removedConfigs, ",")
	for _, key := range configs {
		pathKey := GenerateConfigKey(configFilePath, key)

		// Check if config is removed from configs files
		_, ok := configProps.Get(key)
		if ok {
			return fmt.Errorf("failed to remove config from config file !!!")
		}

		// Check if config is removed from secrets files
		_, ok = secretsProps.Get(pathKey)
		if ok {
			return fmt.Errorf("failed to remove config from secrets file !!!")
		}
	}

	return nil
}

func validateUsingDecryption(configFilePath string, localSecureConfigPath string, outputConfigPath string, origConfigs string, plugin *PasswordProtectionSuite) error {
	err := plugin.DecryptConfigFileSecrets(configFilePath, localSecureConfigPath, outputConfigPath)
	if err != nil {
		return fmt.Errorf("failed to decrypt config file !!!")
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
			return fmt.Errorf("Configs file is empty !!!")
		}

	}

	return nil
}

func setUpDir(masterKeyPassphrase string, secureDir string, configFile string, localSecureConfigPath string, contents string) (*PasswordProtectionSuite, error) {
	err := os.MkdirAll(secureDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Failed to create password protection directory")
	}
	logger := log.New()
	plugin := NewPasswordProtectionPlugin(logger)

	// Set master key
	err = createMasterKey(masterKeyPassphrase, localSecureConfigPath, plugin)
	if err != nil {
		return nil, fmt.Errorf("Failed to create master key")
	}

	err = createNewConfigFile(configFile, contents)
	if err != nil {
		return nil, fmt.Errorf("Failed to create config file")
	}

	return plugin, nil
}

func checkError(err error, wantErr bool, req *require.Assertions) {
	if wantErr {
		req.Error(err)
	} else {
		req.NoError(err)
	}
}
