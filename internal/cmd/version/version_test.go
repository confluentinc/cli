package version

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/terminal"
	"github.com/confluentinc/cli/internal/pkg/version"
	cliMock "github.com/confluentinc/cli/mock"
)

func TestVersion(t *testing.T) {
	req := require.New(t)

	root, prompt := terminal.BuildRootCommand()
	v := version.NewVersion("1.2.3", "abc1234", "Fri Feb 22 20:55:53 UTC 2019", "CI")
	cmd := NewVersionCmd(&cliMock.Commander{Prompt: prompt}, v, prompt)
	root.AddCommand(cmd)

	output, err := terminal.ExecuteCommand(root, "version")
	req.NoError(err)
	req.Regexp(`Version: *1.2.3`, output)
	req.Regexp(`Git Ref: *abc1234`, output)
	req.Regexp(`Build Date: *Fri Feb 22 20:55:53 UTC 2019`, output)
	req.Regexp(`Build Host: *CI`, output)
	req.Regexp(`Development: *false`, output)
}

func TestDevelopmentVersion_v0(t *testing.T) {
	req := require.New(t)

	root, prompt := terminal.BuildRootCommand()
	v := version.NewVersion("0.0.0", "abc1234", "01/23/45", "CI")
	cmd := NewVersionCmd(&cliMock.Commander{Prompt: prompt}, v, prompt)
	root.AddCommand(cmd)

	output, err := terminal.ExecuteCommand(root, "version")
	req.NoError(err)
	req.Regexp(`Version: *0.0.0`, output)
	req.Regexp(`Git Ref: *abc1234`, output)
	req.Regexp(`Development: *true`, output)
}

func TestDevelopmentVersion_Dirty(t *testing.T) {
	req := require.New(t)

	root, prompt := terminal.BuildRootCommand()
	v := version.NewVersion("1.2.3-dirty-timmy", "abc1234", "01/23/45", "CI")
	cmd := NewVersionCmd(&cliMock.Commander{Prompt: prompt}, v, prompt)
	root.AddCommand(cmd)

	output, err := terminal.ExecuteCommand(root, "version")
	req.NoError(err)
	req.Regexp(`Version: *1.2.3-dirty-timmy`, output)
	req.Regexp(`Git Ref: *abc1234`, output)
	req.Regexp(`Development: *true`, output)
}

func TestDevelopmentVersion_Unmerged(t *testing.T) {
	req := require.New(t)

	root, prompt := terminal.BuildRootCommand()
	v := version.NewVersion("1.2.3-g16dd476", "abc1234", "01/23/45", "CI")
	cmd := NewVersionCmd(&cliMock.Commander{Prompt: prompt}, v, prompt)
	root.AddCommand(cmd)

	output, err := terminal.ExecuteCommand(root, "version")
	req.NoError(err)
	req.Regexp(`Version: *1.2.3-g16dd476`, output)
	req.Regexp(`Git Ref: *abc1234`, output)
	req.Regexp(`Development: *true`, output)
}
