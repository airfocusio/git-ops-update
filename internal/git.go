package internal

type Git struct {
	Provider GitProvider
}

type GitProvider interface {
	Push(dir string, changes Changes, callbacks ...func() error) error
	Request(dir string, changes Changes, callbacks ...func() error) error
	AlreadyRequested(dir string, changes Changes) bool
}

type GitAuthor struct {
	Name  string
	Email string
}

const branchPrefix = "git-ops-update"
