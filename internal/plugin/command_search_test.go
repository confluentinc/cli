package plugin

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestClonePluginRepo(t *testing.T) {
	// Set up dummy repo to clone from
	sourceDir, err := os.MkdirTemp("", "source")
	assert.NoError(t, err)
	defer os.RemoveAll(sourceDir)

	r, err := git.PlainInit(sourceDir, false)
	assert.NoError(t, err)
	w, err := r.Worktree()
	assert.NoError(t, err)

	file, err := os.Create(fmt.Sprintf("%s/file.txt", sourceDir))
	assert.NoError(t, err)
	file.Close()

	_, err = w.Add("file.txt")
	assert.NoError(t, err)
	commitOptions := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@confluent.io",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	}
	_, err = w.Commit("test commit", commitOptions)
	assert.NoError(t, err)
	_, err = w.Commit("test commit 2", commitOptions)
	assert.NoError(t, err)

	// Clone repo
	dir, err := os.MkdirTemp("", "clone")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	localRepoPath, _ := filepath.Abs(sourceDir)
	repoUrl, _ := url.Parse(localRepoPath)
	repoUrl.Scheme = "file"
	r, err = clonePluginRepo(dir, repoUrl.String())
	assert.NoError(t, err)

	// Check that the number of commits is 1 under shallow clone
	commits, err := r.Log(&git.LogOptions{})
	if !assert.NoError(t, err) {
		// ForEach below will panic if this assert is false, so exit early
		return
	}

	commitCount := 0
	_ = commits.ForEach(func(commit *object.Commit) error {
		commitCount++
		return nil
	})
	assert.Equal(t, 1, commitCount)
}

func TestGetPluginManifests(t *testing.T) {
	dir, _ := filepath.Abs("../../../test/fixtures/input/plugin")
	manifests, err := getPluginManifests(dir)
	assert.NoError(t, err)

	referenceManifests := []*ManifestOut{
		{
			Name:         "confluent-test_plugin",
			Description:  "Does nothing",
			Dependencies: "Python 3",
		},
	}
	assert.True(t, reflect.DeepEqual(referenceManifests, manifests))
}
