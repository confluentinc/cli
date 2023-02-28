package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	req := require.New(t)

	v := NewVersion("1.2.3", "abc1234", "Fri Feb 22 20:55:53 UTC 2019")

	req.Regexp(`Version: *1.2.3`, v.String())
	req.Regexp(`Git Ref: *abc1234`, v.String())
	req.Regexp(`Build Date: *Fri Feb 22 20:55:53 UTC 2019`, v.String())
	req.Regexp(`Development: *false`, v.String())
}

func TestNewVersion_v0(t *testing.T) {
	req := require.New(t)

	v := NewVersion("0.0.0", "abc1234", "01/23/45")

	req.Regexp(`Version: *0.0.0`, v.String())
	req.Regexp(`Git Ref: *abc1234`, v.String())
	req.Regexp(`Development: *true`, v.String())
}

func TestNewVersion_Dirty(t *testing.T) {
	req := require.New(t)

	v := NewVersion("1.2.3-dirty-timmy", "abc1234", "01/23/45")

	req.Regexp(`Version: *1.2.3-dirty-timmy`, v.String())
	req.Regexp(`Git Ref: *abc1234`, v.String())
	req.Regexp(`Development: *true`, v.String())
}

func TestNewVersion_Unmerged(t *testing.T) {
	req := require.New(t)

	v := NewVersion("1.2.3-g16dd476", "abc1234", "01/23/45")

	req.Regexp(`Version: *1.2.3-g16dd476`, v.String())
	req.Regexp(`Git Ref: *abc1234`, v.String())
	req.Regexp(`Development: *true`, v.String())
}
