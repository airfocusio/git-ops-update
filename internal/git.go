package internal

import (
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Git struct {
	Provider GitProvider
}

type GitProvider interface {
	Push(dir string, changes Changes, callbacks ...func() error) error
	Request(dir string, changes Changes, callbacks ...func() error) error
	AlreadyRequested(dir string, changes Changes) bool
}

type GitAuthor struct {
	Name    string
	Email   string
	SignKey *openpgp.Entity
}

const branchPrefix = "git-ops-update"

func applyChangesAsCommit(worktree git.Worktree, dir string, changes Changes, message string, author GitAuthor, callbacks ...func() error) (*plumbing.Hash, error) {
	err := changes.Push(dir)
	if err != nil {
		return nil, err
	}
	err = runCallbacks(callbacks)
	if err != nil {
		return nil, err
	}
	excludes, err := gitignore.ReadPatterns(osfs.New(dir), []string{})
	if err != nil {
		return nil, err
	}
	worktree.Excludes = excludes

	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}
	for path, status := range status {
		if status.Worktree != git.Unmodified || status.Staging != git.Unmodified {
			_, err := worktree.Add(path)
			if err != nil {
				return nil, err
			}
		}
	}

	signature := object.Signature{Name: author.Name, Email: author.Email, When: time.Now()}
	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author:  &signature,
		SignKey: author.SignKey,
	})
	if err != nil {
		return nil, err
	}

	return &commit, nil
}
