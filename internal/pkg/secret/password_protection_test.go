package secret

import (
	"github.com/confluentinc/cli/internal/pkg/log"
	"io/ioutil"
	"os"
	"strings"
	"testing"
     secrets "github.com/confluentinc/cli/internal/pkg/secret"
)


func TestPasswordProtectionSuite_CreateMasterKey(t *testing.T) {
	type args struct {
		masterKeyPassphrase   string
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
			plugin := secrets.NewPasswordProtectionPlugin(logger)
			key, err:= plugin.CreateMasterKey(tt.args.masterKeyPassphrase)
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
		contents string
		masterKeyPassphrase string
		configFilePath string
		localSecureConfigPath string
		remoteSecureConfigPath string
		secureDir string
		setMEK bool
		createConfig bool
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "InvalidTestCase: Master Key not set",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: "testPassword=password",
				configFilePath: "/tmp/securePass987/config.properties",
				localSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				secureDir: "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:false,
				createConfig:true,
			},
			wantErr: true,
		},
		{
			name: "InvalidTestCase: Invalid Config File Path",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: "testPassword=password",
				configFilePath: "/tmp/securePass987/random.properties",
				localSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				secureDir: "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:true,
				createConfig:false,
			},
			wantErr: true,
		},
		{
			name: "ValidTestCase: Encrypt Config File, Create New DEK",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: "testPassword=password",
				configFilePath: "/tmp/securePass987/config.properties",
				localSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				secureDir: "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				setMEK:true,
				createConfig:true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := secrets.NewPasswordProtectionPlugin(logger)
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
			os.Unsetenv(secrets.CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func TestPasswordProtectionSuite_DecryptConfigFileSecrets(t *testing.T) {
	type args struct {
		contents string
		masterKeyPassphrase string
		configFilePath string
		outputConfigPath string
		localSecureConfigPath string
		remoteSecureConfigPath string
		secureDir string
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: Decrypt Config File",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: "testPassword = password\n",
				configFilePath: "/tmp/securePass987/config.properties",
				outputConfigPath: "/tmp/securePass987/output.properties",
				localSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				secureDir: "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			plugin := secrets.NewPasswordProtectionPlugin(logger)
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

			err = plugin.DecryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath)

			if err != nil {
				t.Fail()
			}

			decryptContent, _ := ioutil.ReadFile(tt.args.outputConfigPath)
			decryptContentStr := string(decryptContent)
			if strings.Compare(decryptContentStr, tt.args.contents) != 0 {
				t.Fail()
			}

			// Clean Up
			os.Unsetenv(secrets.CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func TestPasswordProtectionSuite_AddConfigFileSecrets(t *testing.T) {
	type args struct {
		contents string
		masterKeyPassphrase string
		configFilePath string
		localSecureConfigPath string
		remoteSecureConfigPath string
		secureDir string
		newConfigs string
		outputConfigPath string
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			name: "ValidTestCase: Add new configs",
			args: &args{
				masterKeyPassphrase: "abc123",
				contents: "testPassword = password\n",
				configFilePath: "/tmp/securePass987/config.properties",
				localSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				secureDir: "/tmp/securePass987",
				remoteSecureConfigPath: "/tmp/securePass987/secureConfig.properties",
				outputConfigPath: "/tmp/securePass987/output.properties",
				newConfigs: "ssl.keystore.password = sslPass\ntruststore.keystore.password = keystorePass\n",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SetUp
			_ = os.MkdirAll(tt.args.secureDir, os.ModePerm)
			logger := log.New()
			plugin := secrets.NewPasswordProtectionPlugin(logger)

			// Set master key
			err := createMasterKey(tt.args.masterKeyPassphrase, plugin)
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

			// Verify passwords are added
			err = plugin.DecryptConfigFileSecrets(tt.args.configFilePath, tt.args.localSecureConfigPath, tt.args.outputConfigPath)

			if err != nil {
				t.Fail()
			}

			decryptContent, _ := ioutil.ReadFile(tt.args.outputConfigPath)
			decryptContentStr := string(decryptContent)
			if strings.Compare(decryptContentStr, tt.args.newConfigs) != 0 {
				t.Fail()
			}
			// Clean Up
			os.Unsetenv(secrets.CONFLUENT_KEY_ENVVAR)
			os.RemoveAll(tt.args.secureDir)
		})
	}
}

func createMasterKey(passphrase string, plugin *secrets.PasswordProtectionSuite) error {
	key, err := plugin.CreateMasterKey(passphrase)
	if err != nil {
		return err
	}

	os.Setenv(secrets.CONFLUENT_KEY_ENVVAR, key)

	return nil
}

func createNewConfigFile(path string, contents string) error {
	err := ioutil.WriteFile(path, []byte(contents), 0644)
	return err
}
