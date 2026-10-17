//go:build eval

package eval

import (
	"path/filepath"
	"testing"
)

func TestSharedProvisionerReturnsOneDirForAllSessions(t *testing.T) {
	root := t.TempDir()
	p := NewSharedProvisioner(root)

	a := p.HomeDir(0)
	b := p.HomeDir(1)

	if a != b {
		t.Fatalf("shared provisioner returned different dirs: %q vs %q", a, b)
	}
	if a != root {
		t.Fatalf("shared provisioner returned %q, want root %q", a, root)
	}
}

func TestIsolatedProvisionerReturnsDistinctDirsPerSession(t *testing.T) {
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)

	a := p.HomeDir(0)
	b := p.HomeDir(1)

	if a == b {
		t.Fatalf("isolated provisioner returned same dir for two sessions: %q", a)
	}
	if want := filepath.Join(root, "session-0"); a != want {
		t.Fatalf("session 0 dir = %q, want %q", a, want)
	}
}
