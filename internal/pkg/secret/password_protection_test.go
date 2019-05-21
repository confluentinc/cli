package secret

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/log"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"github.com/magiconair/properties"
)

func TestPasswordProtectionSuite_CreateMasterKey(t *testing.T) {
	type args struct {
		masterKeyPassphrase string
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: valid create master key",
			args: &args{
				masterKeyPassphrase: "abc123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			plugin := NewPasswordProtectionPlugin(logger)
			key, err := plugin.CreateMasterKey(tt.args.masterKeyPassphrase)
			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			if !tt.wantErr && key == "" {
				t.Fail()
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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "InvalidTestCase: Master Key not set",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:                 false,
				createConfig:           true,
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Invalid Config File Path",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/random.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:                 true,
				createConfig:           false,
			},
			wantErr: true,
		},
		{
			name: "ValidTestCase: Encrypt Config File, Create New DEK",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword=password",
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:                 true,
				createConfig:           true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := NewPasswordProtectionPlugin(logger)
			if tt.args.setMEK {
				err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
				if err != nil {
					t.Fail()
				}
			}
			if tt.args.createConfig {
				err := createNewConfigFile(tt.args.configFilePath, tt.args.contents)
				if err != nil {
					t.Fail()
				}
			}

			err := plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath)

			if (err != nil) != tt.wantErr {
				t.Fail()
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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: Decrypt Config File",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/config.properties",
				outputConfigPath:       "/tmp/securePass987/output.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := NewPasswordProtectionPlugin(logger)
			err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
			if err != nil {
				t.Fail()
			}

			err = createNewConfigFile(tt.args.configFilePath, tt.args.contents)
			if err != nil {
				t.Fail()
			}

			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath)

			if err != nil {
				t.Fail()
			}

			err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)

			if err != nil {
				t.Fail()
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SetUp
			err := os.MkdirAll(tt.args.secureDir, os.ModePerm)
			if err != nil {
				t.Fail()
			}
			logger := log.New()
			plugin := NewPasswordProtectionPlugin(logger)

			// Set master key
			err = createMasterKey(tt.args.masterKeyPassphrase, plugin)
			if err != nil {
				t.Fail()
			}

			err = createNewConfigFile(tt.args.configFilePath, tt.args.contents)
			if err != nil {
				t.Fail()
			}

			err = plugin.AddEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.newConfigs)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.newConfigs, plugin)

			if err != nil {
				t.Fail()
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
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/output.properties",
				updateConfigs:          "testPassword = newPassword\n",
			},
			wantErr: false,
		},
		{
			name: "InvalidTestCase: Key not present in config file",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/output.properties",
				updateConfigs:          "ssl.keystore.password = newSslPass\ntruststore.keystore.password = newKeystorePass\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SetUp
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			logger := log.New()
			plugin := NewPasswordProtectionPlugin(logger)

			// Set master key
			err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
			if err != nil {
				t.Fail()
			}

			err = createNewConfigFile(tt.args.configFilePath, tt.args.contents)
			if err != nil {
				t.Fail()
			}

			err = plugin.UpdateEncryptedPasswords(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath, tt.args.updateConfigs)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			if !tt.wantErr {
				// Verify passwords are added
				err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.updateConfigs, plugin)

				if err != nil {
					t.Fail()
				}
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
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: Rotate DEK",
			args: &args{
				masterKeyPassphrase:    "abc123",
				contents:               "testPassword = password\n",
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/output.properties",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := NewPasswordProtectionPlugin(logger)
			err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
			if err != nil {
				t.Fail()
			}

			err = createNewConfigFile(tt.args.configFilePath, tt.args.contents)
			if err != nil {
				t.Fail()
			}


			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}
			originalProps := properties.MustLoadFile(tt.args.localSecureConfigPath, properties.ISO_8859_1)

			err = plugin.RotateDataKey(tt.args.masterKeyPassphrase, tt.args.localSecureConfigPath)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			// Verify the encrypted values are different
			rotatedProps := properties.MustLoadFile(tt.args.localSecureConfigPath, properties.ISO_8859_1)
			for key, value := range originalProps.Map() {

				if !strings.HasPrefix(key, "metadata") {
					cipher := rotatedProps.GetString(key, "")
					if cipher == value {
						t.Fail()
					}
				}

			}

			err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)

			if err != nil {
				t.Fail()
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
		configFilePath         string
		localSecureConfigPath  string
		remoteSecureConfigPath string
		outputConfigPath       string
		secureDir              string
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
				configFilePath:         "/tmp/securePass987/config.properties",
				localSecureConfigPath:  "/tmp/securePass987/secureConfig.properties",
				secureDir:              "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				outputConfigPath:       "/tmp/securePass987/output.properties",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := NewPasswordProtectionPlugin(logger)
			err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
			if err != nil {
				t.Fail()
			}

			err = createNewConfigFile(tt.args.configFilePath, tt.args.contents)
			if err != nil {
				t.Fail()
			}


			err = plugin.EncryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.remoteSecureConfigPath)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			newKey, err := plugin.RotateMasterKey(tt.args.masterKeyPassphrase, tt.args.newMasterKeyPassphrase, tt.args.localSecureConfigPath)

			if (err != nil) != tt.wantErr {
				t.Fail()
			}

			os.Setenv(CONFLUENT_KEY_ENVVAR, newKey)

			err = validateUsingDecryption(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath, tt.args.contents, plugin)

			if err != nil {
				t.Fail()
			}

			// Clean Up
			os.Unsetenv(CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}


func createMasterKey(passphrase string, plugin *PasswordProtectionSuite) error {
	key, err := plugin.CreateMasterKey(passphrase)
	if err != nil {
		return err
	}

	os.Setenv(CONFLUENT_KEY_ENVVAR, key)

	return nil
}

func createNewConfigFile(path string, contents string) error {
	err := ioutil.WriteFile(path, []byte(contents), 0644)
	return err
}

func validateUsingDecryption(configFilePath string, localSecureConfigPath string, outputConfigPath string, origConfigs string, plugin *PasswordProtectionSuite) error {
	err := plugin.DecryptConfigFileSecrets(configFilePath, localSecureConfigPath, outputConfigPath)

	if err != nil {
		return fmt.Errorf("failed DecryptConfigFileSecrets mster master ")
	}

	decryptContent, _ := ioutil.ReadFile(outputConfigPath)
	decryptContentStr := string(decryptContent)
	decryptConfigProps := properties.MustLoadString(decryptContentStr)
	originalConfigProps := properties.MustLoadString(origConfigs)
	for key, value := range decryptConfigProps.Map() {
		originalVal, _ := originalConfigProps.Get(key)
		if value != originalVal {
			return fmt.Errorf("Configs file is empty !!!")
		}

	}

	return nil
}
