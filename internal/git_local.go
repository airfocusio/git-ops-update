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

func (p LocalGitProvider) Push(dir string, changeSet ChangeSet, callbacks ...func() error) error {
	warnLocalGitProvider()
	err := changeSet.Push(dir)
	if err != nil {
		return err
	}
	return runCallbacks(callbacks)
}

func (p LocalGitProvider) Request(dir string, changeSet ChangeSet, callbacks ...func() error) error {
	warnLocalGitProvider()
	err := p.Push(dir, changeSet)
	if err != nil {
		return err
	}
	return runCallbacks(callbacks)
}
