package local

import (
	"errors"
	"os"
	"testing"

	"github.com/atrox/homedir"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/cli/mock/local"
)

func TestLocal(t *testing.T) {
	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any())
	shellRunner.EXPECT().Run("main", gomock.Eq([]string{"local", "help"})).Return(0, nil)
	localCmd := New(&cliMock.Commander{}, shellRunner, &mock.FileSystem{})
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.NoError(err)
}

func TestLocalErrorDuringSource(t *testing.T) {
	req := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellRunner := mock_local.NewMockShellRunner(ctrl)
	shellRunner.EXPECT().Init(os.Stdout, os.Stderr)
	shellRunner.EXPECT().Export("CONFLUENT_HOME", "blah")
	shellRunner.EXPECT().Source("cp_cli/confluent.sh", gomock.Any()).Return(errors.New("oh no"))
	localCmd := New(&cliMock.Commander{}, shellRunner, &mock.FileSystem{})
	_, err := cmd.ExecuteCommand(localCmd, "local", "--path", "blah", "help")
	req.Error(err)
}

func TestDetermineConfluentInstallDir(t *testing.T) {
	tests := []struct {
		name      string
		dirExists map[string][]string
		wantDir   string
		wantFound bool
		wantErr   bool
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
			wantDir:   "/opt/confluent-5.2.2",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "multiple versioned directories found in /opt (reverse order)",
			dirExists: map[string][]string{"/opt/confluent*": {"/opt/confluent-4.1.0", "/opt/confluent-5.2.2"}},
			wantDir:   "/opt/confluent-5.2.2",
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
			}
			dir, found, err := determineConfluentInstallDir(fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineConfluentInstallDir() error: %v, wantErr: %v", err, tt.wantErr)
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
