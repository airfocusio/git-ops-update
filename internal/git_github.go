package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

var _ GitProvider = (*GitHubGitProvider)(nil)

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

	message, _ := changes.Message()
	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Title()+"\n\n"+message, p.Author, callbacks...)
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
	remoteRefs, err := remote.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: "api",
			Password: p.AccessToken,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to list git branches: %w", err)
	}
	ownerName, repoName, err := extractGitHubOwnerRepoFromRemote(*remote)
	if err != nil {
		return fmt.Errorf("unable to extract github owner/repository from remote origin: %w", err)
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	existingBranches := []string{}
	existingPullRequests := []*github.PullRequest{}
	targetBranchFindPrefix := fmt.Sprintf("refs/heads/%s", changes.BranchFindPrefix(branchPrefix))
	targetBranchGroupHash := changes.GroupHash()
	targetBranchHash := changes.Hash()
	targetBranchExists := false
	for _, ref := range remoteRefs {
		refName := ref.Name().String()
		if strings.HasPrefix(refName, targetBranchFindPrefix) && strings.Contains(refName, targetBranchHash) {
			targetBranchExists = true
		} else if strings.HasPrefix(refName, targetBranchFindPrefix) && strings.Contains(refName, targetBranchGroupHash) {
			existingBranches = append(existingBranches, refName)
			pullRequests, res, err := client.PullRequests.List(context.Background(), *ownerName, *repoName, &github.PullRequestListOptions{
				State: "open",
				Head:  fmt.Sprintf("%s:%s", *ownerName, strings.TrimPrefix(refName, "refs/heads/")),
			})
			if err != nil {
				LogWarning("Unable to search pull request for branch %s from github repository %s/%s: %v", refName, *ownerName, *repoName, err)
			}
			defer res.Body.Close()
			existingPullRequests = append(existingPullRequests, pullRequests...)
		}
	}
	if targetBranchExists {
		return nil
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

	message, fullMessage := changes.Message()
	_, err = applyChangesAsCommit(*worktree, dir, changes, changes.Title()+"\n\n"+message, p.Author, callbacks...)
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

	LogDebug("Creating pull request for branch %s to github repository %s/%s", targetBranch.Short(), *ownerName, *repoName)
	pullRequestBase := string(baseBranch.Name())
	pullRequestHead := string(targetBranch)
	pullRequestTitle := changes.Title()
	pullRequestBody := fullMessage
	pullRequest, res, err := client.PullRequests.Create(context.Background(), *ownerName, *repoName, &github.NewPullRequest{
		Title: &pullRequestTitle,
		Base:  &pullRequestBase,
		Head:  &pullRequestHead,
		Body:  &pullRequestBody,
	})
	if err != nil {
		return fmt.Errorf("unable to create github pull request: %w", err)
	}
	defer res.Body.Close()
	existingPullRequestLabels := sliceUnique(sliceFlatMap(existingPullRequests, func(pr *github.PullRequest) []string {
		return sliceMap(pr.Labels, func(l *github.Label) string {
			return *l.Name
		})
	}))
	if len(existingPullRequestLabels) > 0 {
		LogDebug("Adding labels for pull request %d to github repository %s/%s", *pullRequest.Number, *ownerName, *repoName)
		_, res, err = client.Issues.AddLabelsToIssue(context.Background(), *ownerName, *repoName, *pullRequest.Number, existingPullRequestLabels)
		if err != nil {
			return fmt.Errorf("unable to add github pull request labels: %w", err)
		}
		defer res.Body.Close()
	}

	for _, existingPullRequest := range existingPullRequests {
		LogDebug("Commenting on superseded pull request %d to github repository %s/%s", *existingPullRequest.Number, *ownerName, *repoName)
		closeMessage := fmt.Sprintf("Superseded by #%d", *pullRequest.Number)
		_, res, err = client.Issues.CreateComment(context.Background(), *ownerName, *repoName, *existingPullRequest.Number, &github.IssueComment{
			Body: &closeMessage,
		})
		if err != nil {
			return fmt.Errorf("unable to add github pull request comment: %w", err)
		}
		defer res.Body.Close()
	}

	for _, refName := range existingBranches {
		LogDebug("Removing branch %s from github repository %s/%s", refName, *ownerName, *repoName)
		err = remote.Push(&git.PushOptions{
			Auth: &http.BasicAuth{
				Username: "api",
				Password: p.AccessToken,
			},
			RefSpecs: []config.RefSpec{
				config.RefSpec(":" + refName),
			},
		})
		if err != nil {
			LogWarning("Unable to remove branch %s from github repository %s/%s: %v", refName, *ownerName, *repoName, err)
		}
	}

	err = worktree.Checkout(&git.CheckoutOptions{Branch: baseBranch.Name()})
	if err != nil {
		return fmt.Errorf("unable to checkout to base branch: %w", err)
	}

	return nil
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
