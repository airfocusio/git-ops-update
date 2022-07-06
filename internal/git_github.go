package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

type GitHubGitProvider struct {
	Author      GitAuthor
	AccessToken string
}

func (p GitHubGitProvider) Push(dir string, changes Changes, callbacks ...func() error) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("unable to open git repository: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to open git worktree: %w", err)
	}

	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Title()+"\n\n"+changes.Message(), p.Author, callbacks...)
	if err != nil {
		return fmt.Errorf("unable to commit changes: %w", err)
	}
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to push changes: %w", err)
	}
	return nil
}

func (p GitHubGitProvider) Request(dir string, changes Changes, callbacks ...func() error) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("unable to open git repository: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to open git worktree: %w", err)
	}
	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("unable to get git remote origin: %w", err)
	}
	if err != nil {
		return fmt.Errorf("unable to list git branches: %w", err)
	}
	targetBranch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", changes.Branch(branchPrefix)))

	baseBranch, err := repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get base branch: %w", err)
	}
	LogDebug("Creating git branch %s", targetBranch.Short())
	err = worktree.Checkout(&git.CheckoutOptions{Branch: targetBranch, Create: true})
	if err != nil {
		return fmt.Errorf("unable to create target branch: %w", err)
	}
	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Title()+"\n\n"+changes.Message(), p.Author, callbacks...)
	if err != nil {
		return fmt.Errorf("unable to commit changes: %w", err)
	}
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to push changes: %w", err)
	}

	ownerName, repoName, err := extractGitHubOwnerRepoFromRemote(*remote)
	if err != nil {
		return fmt.Errorf("unable to extract github owner/repository from remote origin: %w", err)
	}
	LogDebug("Creating pull request for branch %s to github repository %s/%s", targetBranch.Short(), *ownerName, *repoName)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	pullBase := string(baseBranch.Name())
	pullHead := string(targetBranch)
	pullTitle := changes.Title()
	pullBody := changes.Message()
	_, res, err := client.PullRequests.Create(context.Background(), *ownerName, *repoName, &github.NewPullRequest{
		Title: &pullTitle,
		Base:  &pullBase,
		Head:  &pullHead,
		Body:  &pullBody,
	})
	if err != nil {
		return fmt.Errorf("unable to create github pull request: %w", err)
	}
	defer res.Body.Close()

	err = worktree.Checkout(&git.CheckoutOptions{Branch: baseBranch.Name()})
	if err != nil {
		return fmt.Errorf("unable to checkout to base branch: %w", err)
	}
	return nil
}

func (p GitHubGitProvider) AlreadyRequested(dir string, changes Changes) bool {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return false
	}
	remote, err := repo.Remote("origin")
	if err != nil {
		return false
	}
	remoteRefs, err := remote.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return false
	}
	targetBranchFindPrefix := fmt.Sprintf("refs/heads/%s", changes.BranchFindPrefix(branchPrefix))
	targetBranchFindSuffix := changes.BranchFindSuffix()
	targetBranchExists := false
	for _, ref := range remoteRefs {
		refName := ref.Name().String()
		if strings.HasPrefix(refName, targetBranchFindPrefix) && strings.HasSuffix(refName, targetBranchFindSuffix) {
			targetBranchExists = true
			break
		}
	}
	return targetBranchExists
}

func extractGitHubOwnerRepoFromRemote(remote git.Remote) (*string, *string, error) {
	httpRegex := regexp.MustCompile(`^https://github.com/(?P<owner>[^/]+)/(?P<repo>.*)$`)
	sshRegex := regexp.MustCompile(`^git@github.com:(?P<owner>[^/]+)/(?P<repo>.*)$`)
	for _, url := range remote.Config().URLs {
		httpMatch := httpRegex.FindStringSubmatch(url)
		if httpMatch != nil {
			repoName := strings.TrimSuffix(httpMatch[2], ".git")
			return &httpMatch[1], &repoName, nil
		}
		sshMatch := sshRegex.FindStringSubmatch(url)
		if sshMatch != nil {
			repoName := strings.TrimSuffix(sshMatch[2], ".git")
			return &sshMatch[1], &repoName, nil
		}
	}
	return nil, nil, fmt.Errorf("none of the git remote %s urls %v could be recognized as a github repository", remote.Config().Name, remote.Config().URLs)
}
