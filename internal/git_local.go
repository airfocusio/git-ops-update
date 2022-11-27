package internal

var _ GitProvider = (*LocalGitProvider)(nil)

type LocalGitProvider struct {
	Author GitAuthor
}

var localGitProviderWarned = false

func warnLocalGitProvider() {
	if !localGitProviderWarned {
		localGitProviderWarned = true
		LogWarning("Local git provider does not support push mode. Will apply changes to worktree")
	}
}

func (p LocalGitProvider) Push(dir string, changes Changes, callbacks ...func() error) error {
	warnLocalGitProvider()
	err := changes.Push(dir)
	if err != nil {
		return err
	}
	return runCallbacks(callbacks)
}

func (p LocalGitProvider) Request(dir string, changes Changes, callbacks ...func() error) error {
	warnLocalGitProvider()
	err := p.Push(dir, changes)
	if err != nil {
		return err
	}
	return runCallbacks(callbacks)
}

func (p LocalGitProvider) AlreadyRequested(dir string, changes Changes) bool {
	return false
}
