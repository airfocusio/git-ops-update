package internal

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

type Git struct {
	Provider GitProvider
}

type GitProvider interface {
	Apply(dir string, changes Changes) (bool, error)
	Request(dir string, changes Changes) (bool, error)
}

type GitAuthor struct {
	Name  string
	Email string
}

type LocalGitProvider struct {
	Author GitAuthor
}

type GitHubGitProvider struct {
	Author      GitAuthor
	AccessToken string
}

type Action func(dir string, changes Changes) (bool, error)

var branchPrefix = "git-ops-update"

func (p LocalGitProvider) Apply(dir string, changes Changes) (bool, error) {
	err := changes.Apply(dir)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p LocalGitProvider) Request(dir string, changes Changes) (bool, error) {
	fmt.Printf("Local git provider does not support request mode. Will apply changes directly instead")
	return p.Apply(dir, changes)
}

func (p GitHubGitProvider) Apply(dir string, changes Changes) (bool, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return false, fmt.Errorf("unable to open git repository: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("unable to open git worktree: %v", err)
	}

	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Message(), p.Author)
	if err != nil {
		return false, fmt.Errorf("unable to commit changes: %v", err)
	}
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return false, fmt.Errorf("unable to push changes: %v", err)
	}
	return true, nil
}

func (p GitHubGitProvider) Request(dir string, changes Changes) (bool, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return false, fmt.Errorf("unable to open git repository: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("unable to open git worktree: %v", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return false, fmt.Errorf("unable to get git remote origin: %v", err)
	}
	remoteRefs, err := remote.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return false, fmt.Errorf("unable to list git branches: %v", err)
	}
	targetBranch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", changes.Branch(branchPrefix)))
	targetBranchExists := false
	for _, ref := range remoteRefs {
		if ref.Name() == targetBranch {
			targetBranchExists = true
			break
		}
	}

	if !targetBranchExists {
		baseBranch, err := repo.Head()
		if err != nil {
			return false, fmt.Errorf("unable to get base branch: %v", err)
		}
		err = worktree.Checkout(&git.CheckoutOptions{Branch: targetBranch, Create: true})
		if err != nil {
			return false, fmt.Errorf("unable to create target branch: %v", err)
		}
		_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Message(), p.Author)
		if err != nil {
			return false, fmt.Errorf("unable to commit changes: %v", err)
		}
		err = repo.Push(&git.PushOptions{
			Auth: &http.BasicAuth{
				Username: "api",
				Password: p.AccessToken,
			},
		})
		if err != nil {
			return false, fmt.Errorf("unable to push changes: %v", err)
		}

		owner, repo, err := extractGitHubOwnerRepoFromRemote(*remote)
		if err != nil {
			return false, fmt.Errorf("unable to extract github owner/repository from remote origin: %v", err)
		}

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.AccessToken})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		pullTitle := "Update"
		pullBase := string(baseBranch.Name())
		pullHead := string(targetBranch)
		pullBody := changes.Message()
		_, res, err := client.PullRequests.Create(context.Background(), *owner, *repo, &github.NewPullRequest{
			Title: &pullTitle,
			Base:  &pullBase,
			Head:  &pullHead,
			Body:  &pullBody,
		})
		if err != nil {
			return false, fmt.Errorf("unable to create github pull request: %v", err)
		}
		defer res.Body.Close()

		err = worktree.Checkout(&git.CheckoutOptions{Branch: baseBranch.Name()})
		if err != nil {
			return true, fmt.Errorf("unable to checkout to bae branch: %v", err)
		}
		return true, nil
	}

	return false, nil
}

func applyChangesAsCommit(worktree git.Worktree, dir string, changes Changes, message string, author GitAuthor) (*plumbing.Hash, error) {
	changes.Apply(dir, func(file string) error {
		_, err := worktree.Add(file)
		return err
	})

	signature := object.Signature{Name: author.Name, Email: author.Email, When: time.Now()}
	commit, err := worktree.Commit(message, &git.CommitOptions{Author: &signature})
	if err != nil {
		return nil, err
	}

	return &commit, nil
}

func extractGitHubOwnerRepoFromRemote(remote git.Remote) (*string, *string, error) {
	httpRegex := regexp.MustCompile(`^https://github.com/(?P<owner>[^/]+)/(?P<repo>.*)(?:\.git)?$`)
	sshRegex := regexp.MustCompile(`^git@github.com:(?P<owner>[^/]+)/(?P<repo>.*)(?:\.git)?$`)
	for _, url := range remote.Config().URLs {
		httpMatch := httpRegex.FindStringSubmatch(url)
		if httpMatch != nil {
			return &httpMatch[1], &httpMatch[2], nil
		}
		sshMatch := sshRegex.FindStringSubmatch(url)
		if sshMatch != nil {
			return &sshMatch[1], &sshMatch[2], nil
		}
	}
	return nil, nil, fmt.Errorf("non of the git remote %s urls %v could be recognized as a github repository", remote.Config().Name, remote.Config().URLs)
}

func getAction(p GitProvider, actionName string) (*Action, error) {
	switch actionName {
	case "":
		return getAction(p, "apply")
	case "apply":
		fn := Action(func(dir string, changes Changes) (bool, error) {
			return p.Apply(dir, changes)
		})
		return &fn, nil
	case "request":
		fn := Action(func(dir string, changes Changes) (bool, error) {
			return p.Request(dir, changes)
		})
		return &fn, nil
	default:
		return nil, fmt.Errorf("unknown action %s", actionName)
	}
}
