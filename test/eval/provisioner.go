//go:build eval

package eval

import (
	"fmt"
	"os"
	"path/filepath"
)

// Provisioner decides the HOME directory a given concurrent session runs under. It is the isolation
// knob: shared points every session at one dir (state collides); isolated gives each its own.
type Provisioner interface {
	HomeDir(session int) string
}

type sharedProvisioner struct{ root string }

func NewSharedProvisioner(root string) Provisioner {
	mustMkdir(root)
	return &sharedProvisioner{root: root}
}

func (p *sharedProvisioner) HomeDir(int) string { return p.root }

type isolatedProvisioner struct{ root string }

func NewIsolatedProvisioner(root string) Provisioner {
	mustMkdir(root)
	return &isolatedProvisioner{root: root}
}

func (p *isolatedProvisioner) HomeDir(session int) string {
	dir := filepath.Join(p.root, fmt.Sprintf("session-%d", session))
	mustMkdir(dir)
	return dir
}

func mustMkdir(dir string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Sprintf("provisioner: mkdir %q: %v", dir, err))
	}
}
