package internal

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

type Git struct {
	Provider GitProvider
}

type GitProvider interface {
	Apply(dir string, changes Changes) error
	Request(dir string, changes Changes) error
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

type Action func(dir string, changes Changes) error

var branchPrefix = "git-ops-update"

func (p LocalGitProvider) Apply(dir string, changes Changes) error {
	err := changes.Apply(dir)
	if err != nil {
		return err
	}
	return nil
}

func (p LocalGitProvider) Request(dir string, changes Changes) error {
	log.Printf("Local git provider does not support request mode. Will apply changes directly instead")
	return p.Apply(dir, changes)
}

func (p GitHubGitProvider) Apply(dir string, changes Changes) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Message(), p.Author)
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (p GitHubGitProvider) Request(dir string, changes Changes) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	targetBranch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", changes.Branch(branchPrefix)))
	existingBranchesIter, err := repo.Branches()
	if err != nil {
		return err
	}
	targetBranchExists := false
	existingBranchesIter.ForEach(func(b *plumbing.Reference) error {
		if b.Name() == targetBranch {
			targetBranchExists = true
		}
		return nil
	})

	if !targetBranchExists {
		baseBranch, err := repo.Head()
		if err != nil {
			return err
		}
		err = worktree.Checkout(&git.CheckoutOptions{Branch: targetBranch, Create: true})
		if err != nil {
			return err
		}
		_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Message(), p.Author)
		if err != nil {
			return err
		}
		err = repo.Push(&git.PushOptions{})
		if err != nil {
			return err
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			return err
		}
		owner, repo, err := extractGitHubOwnerRepoFromRemote(*remote)
		if err != nil {
			return err
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
			return err
		}
		defer res.Body.Close()

		worktree.Checkout(&git.CheckoutOptions{Branch: baseBranch.Name()})
	}

	return nil
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
	httpRegex := regexp.MustCompile(`^https://github.com/(?P<owner>[^/]+)/(?P<repo>.*)(?:\.git)$`)
	sshRegex := regexp.MustCompile(`^git@github.com:(?P<owner>[^/]+)/(?P<repo>.*)(?:\.git)$`)
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
	return nil, nil, fmt.Errorf("unable to extract github owner/repository from remote %s", remote.Config().Name)
}

func getAction(p GitProvider, actionName string) (*Action, error) {
	switch actionName {
	case "apply":
		fn := Action(func(dir string, changes Changes) error {
			return p.Apply(dir, changes)
		})
		return &fn, nil
	case "request":
		fn := Action(func(dir string, changes Changes) error {
			return p.Request(dir, changes)
		})
		return &fn, nil
	default:
		return nil, fmt.Errorf("unknown action %s", actionName)
	}
}
