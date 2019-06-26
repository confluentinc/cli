package local

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/cli/mock/local"
)

func TestLocal(t *testing.T) {
	oldCurrent := os.Getenv("CONFLUENT_CURRENT")
	_ = os.Setenv("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	defer func() { _ = os.Setenv("CONFLUENT_CURRENT", oldCurrent) }()

	oldTmp := os.Getenv("TMPDIR")
	_ = os.Setenv("TMPDIR", "/var/folders/some/junk")
	defer func() { _ = os.Setenv("TMPDIR", oldTmp) }()

	oldJavaHome := os.Getenv("JAVA_HOME")
	_ = os.Setenv("JAVA_HOME", "/path/to/java")
	defer func() { _ = os.Setenv("JAVA_HOME", oldJavaHome) }()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", "/path/to/home")
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("HOME", "/path/to/home")
	shellRunner.EXPECT().Export("JAVA_HOME", "/path/to/java")
	shellRunner.EXPECT().Export("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Export("TMPDIR", "/var/folders/some/junk")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any())
	shellRunner.EXPECT().Run("main", gomock.Eq([]string{"local", "help"})).Return(0, nil)
	localCmd := New(&cliMock.Commander{}, shellRunner, &mock.FileSystem{})
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.NoError(err)
}

func TestLocalErrorDuringSource(t *testing.T) {
	oldCurrent := os.Getenv("CONFLUENT_CURRENT")
	_ = os.Setenv("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	defer func() { _ = os.Setenv("CONFLUENT_CURRENT", oldCurrent) }()

	oldTmp := os.Getenv("TMPDIR")
	_ = os.Setenv("TMPDIR", "/var/folders/some/junk")
	defer func() { _ = os.Setenv("TMPDIR", oldTmp) }()

	oldJavaHome := os.Getenv("JAVA_HOME")
	_ = os.Setenv("JAVA_HOME", "/path/to/java")
	defer func() { _ = os.Setenv("JAVA_HOME", oldJavaHome) }()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", "/path/to/home")
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("HOME", "/path/to/home")
	shellRunner.EXPECT().Export("JAVA_HOME", "/path/to/java")
	shellRunner.EXPECT().Export("CONFLUENT_CURRENT", "/path/to/confluent/workdir")
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Export("TMPDIR", "/var/folders/some/junk")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any()).Return(errors.New("oh no"))
	localCmd := New(&cliMock.Commander{}, shellRunner, &mock.FileSystem{})
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.Error(err)
}

func TestDetermineConfluentInstallDir(t *testing.T) {
	tests := []struct {
		name string
		// glob -> matching directory names
		dirExists map[string][]string
		// files that exist in CP install dir (mock) for valid CP install dir canary/heuristics only
		fileExists map[string][]string
		wantDir    string
		wantFound  bool
		wantErr    bool
	}{
		{
			name:      "no directories found",
			dirExists: map[string][]string{},
			wantDir:   "",
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "unversioned directory found in /opt",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent"}},
			wantDir:   "/opt/confluent",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "versioned directory found in /opt",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-5.2.2"}},
			wantDir:   "/opt/confluent-5.2.2",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "unversioned directory found in /usr/local and versioned directory found in ~/Downloads",
			dirExists: map[string][]string{"/usr/local/confluent*": {"/usr/local/confluent"}, "~/Downloads/confluent*": {"~/Downloads/confluent-4.1.0"}},
			wantDir:   "/usr/local/confluent",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directories found in /opt",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-5.2.2", "/opt/confluent-4.1.0"}},
			fileExists: map[string][]string{
				"/opt/confluent-5.2.2/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
				"/opt/confluent-4.1.0/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
			},
			wantDir:   "/opt/confluent-5.2.2",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directories found in /opt (reverse order)",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-4.1.0", "/opt/confluent-5.2.2"}},
			wantDir:   "/opt/confluent-5.2.2",
			fileExists: map[string][]string{
				"/opt/confluent-5.2.2/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
				"/opt/confluent-4.1.0/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directory found in ~/confluent (special test because of the ~)",
			dirExists: map[string][]string{"~/confluent*": {"~/confluent-5.2.2"}},
			wantDir:   "~/confluent-5.2.2",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "not a valid CP install dir - isn't a directory",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-5.2.2.tar.gz"}},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "not a valid CP install dir - missing canary file",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent"}},
			fileExists: map[string][]string{
				"/opt/confluent/bin": {}, // The canary file isn't present
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "not a valid CP install dir - missing kafka-server-start",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent"}},
			fileExists: map[string][]string{
				"/opt/confluent/bin": {
					"connect-distributed",
					"ksql-server-start",
					"zookeeper-server-start",
				},
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directories found in /opt (validate dir is CP installation)",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-4.1.0", "/opt/confluent-5.2.2"}},
			wantDir:   "/opt/confluent-5.2.2",
			fileExists: map[string][]string{
				"/opt/confluent-4.1.0/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
				"/opt/confluent-5.2.2/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directories found in /opt but latest versioned dir is not valid CP install dir",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-5.2.2", "/opt/confluent-4.1.0"}},
			fileExists: map[string][]string{
				"/opt/confluent-4.1.0/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
				"/opt/confluent-5.2.2/bin": {},
			},
			wantDir:   "/opt/confluent-4.1.0",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "accept CONFLUENT_HOME/../etc as valid CP install dir",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-5.2.2"}},
			fileExists: map[string][]string{
				"/opt/confluent-5.2.2/bin": {
					"connect-distributed",
					"kafka-server-start",
					"ksql-server-start",
					"zookeeper-server-start",
					"schema-registry/connect-avro-distributed.properties",
				},
			},
			wantDir:   "/opt/confluent-5.2.2",
			wantFound: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &mock.FileSystem{
				GlobFunc: func(pattern string) ([]string, error) {
					var matches []string
					// we can't just do tt.dirExists[pattern]; pattern has expanded ~ but dirExists doesn't
					for p, dir := range tt.dirExists {
						abs, err := homedir.Expand(p)
						if err != nil {
							return nil, err
						}
						if pattern == abs {
							matches = dir
						}
					}
					// matches won't match what happens in the real world because it still has ~ in it
					// but we'll test for values including the ~ in them in our tests too to simplify things
					return matches, nil
				},
				StatFunc: func(name string) (os.FileInfo, error) {
					// Values for testing the CONFLUENT_HOME directory itself (to make sure it is, in fact, a directory)
					if filepath.Ext(name) == ".gz" || filepath.Ext(name) == ".zip" { // tar.gz or zip file
						return &mock.FileInfo{NameVal: name, IsDirVal: false}, nil
					}
					if filepath.Ext(name) != ".properties" {
						return &mock.FileInfo{NameVal: name, IsDirVal: true}, nil
					}

					// Values for testing the canary files inside the CONFLUENT_HOME directory
					if tt.fileExists[filepath.Dir(name)] == nil {
						// if fileExists isn't set, we assume the file exists since we're not testing these heuristics
						return &mock.FileInfo{NameVal: name, IsDirVal: true}, nil
					}
					for _, d := range tt.fileExists[filepath.Dir(name)] {
						if filepath.Clean(d) == name {
							return &mock.FileInfo{NameVal: name, IsDirVal: true}, nil
						}
					}
					return nil, os.ErrNotExist
				},
				ReadDirFunc: func(dirname string) ([]os.FileInfo, error) {
					if tt.fileExists == nil {
						tt.fileExists = map[string][]string{}
					}
					if tt.fileExists[dirname] == nil {
						tt.fileExists[dirname] = []string{}
						d := filepath.Dir(dirname)
						for _, canary := range validCPInstallBinCanaries {
							tt.fileExists[dirname] = append(tt.fileExists[dirname], canary)
						}
						tt.fileExists[dirname] = append(tt.fileExists[dirname], filepath.Join(d, validCPInstallEtcCanary))
						tt.fileExists[dirname] = append(tt.fileExists[dirname], validCPInstallEtcCanary)
					}
					infos := make([]os.FileInfo, 0, len(tt.fileExists[dirname]))
					for _, f := range tt.fileExists[dirname] {
						infos = append(infos, &mock.FileInfo{NameVal: f})
					}
					return infos, nil
				},
			}
			dir, found, err := determineConfluentInstallDir(fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineConfluentInstallDir() error: %v, wantErr: %v", err, tt.wantErr)
				return
			}
			tt.wantDir, err = homedir.Expand(tt.wantDir)
			if err != nil {
				t.Errorf("Error: %v", err)
				return
			}
			if dir != tt.wantDir {
				t.Errorf("determineConfluentInstallDir() dir = %#v, wantDir %#v", dir, tt.wantDir)
			}
			if found != tt.wantFound {
				t.Errorf("determineConfluentInstallDir() found = %v, wantFound %v", found, tt.wantFound)
			}
		})
	}
}
