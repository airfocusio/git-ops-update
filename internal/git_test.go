package internal

import (
	"os"
	"path"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func TestApplyChangesAsCommit(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "git-ops-update-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	file := path.Join(dir, "file")

	repo, err := git.PlainInit(dir, false)
	assert.NoError(t, err)

	worktree, err := repo.Worktree()
	assert.NoError(t, err)

	status, err := worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, true, status.IsClean())

	err = os.WriteFile(file, []byte("1"), 0o664)
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, false, status.IsClean())

	_, err = applyChangesAsCommit(*worktree, dir, Changes{}, "commit", GitAuthor{Name: "test", Email: "test"})
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, true, status.IsClean())

	err = os.WriteFile(file, []byte("2"), 0o664)
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, false, status.IsClean())

	_, err = applyChangesAsCommit(*worktree, dir, Changes{}, "commit", GitAuthor{Name: "test", Email: "test"})
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, true, status.IsClean())

	err = os.Remove(file)
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, false, status.IsClean())

	_, err = applyChangesAsCommit(*worktree, dir, Changes{}, "commit", GitAuthor{Name: "test", Email: "test"})
	assert.NoError(t, err)

	status, err = worktree.Status()
	assert.NoError(t, err)
	assert.Equal(t, true, status.IsClean())
}
