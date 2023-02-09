package internal

import (
	"context"
	"fmt"
	"path/filepath"
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

type GitHubGitProviderInheritLabels struct {
	Enabled  bool
	Includes []string
	Excludes []string
}

type GitHubGitProvider struct {
	Author        GitAuthor
	AccessToken   string
	InheritLabels GitHubGitProviderInheritLabels
}

func (p GitHubGitProvider) Push(dir string, changeSet ChangeSet, callbacks ...func() error) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("unable to open git repository: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to open git worktree: %w", err)
	}

	message, _ := changeSet.Message()
	_, err = applyChangesAsCommit(*worktree, dir, changeSet, changeSet.Title()+"\n\n"+message, p.Author, callbacks...)
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

func (p GitHubGitProvider) Request(dir string, changeSet ChangeSet, callbacks ...func() error) error {
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
	targetBranchFindPrefix := fmt.Sprintf("refs/heads/%s", changeSet.BranchFindPrefix(branchPrefix))
	targetBranchGroupHash := changeSet.GroupHash()
	targetBranchHash := changeSet.Hash()
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
	targetBranch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", changeSet.Branch(branchPrefix)))

	baseBranch, err := repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get base branch: %w", err)
	}
	LogDebug("Creating git branch %s", targetBranch.Short())
	err = worktree.Checkout(&git.CheckoutOptions{Branch: targetBranch, Create: true})
	if err != nil {
		return fmt.Errorf("unable to create target branch: %w", err)
	}

	message, fullMessage := changeSet.Message()
	_, err = applyChangesAsCommit(*worktree, dir, changeSet, changeSet.Title()+"\n\n"+message, p.Author, callbacks...)
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
	pullRequestTitle := changeSet.Title()
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

	inheritedLabels := p.ExtractInheritedLabels(existingPullRequests)
	if len(inheritedLabels) > 0 {
		LogDebug("Adding labels for pull request %d to github repository %s/%s", *pullRequest.Number, *ownerName, *repoName)
		_, res, err = client.Issues.AddLabelsToIssue(context.Background(), *ownerName, *repoName, *pullRequest.Number, inheritedLabels)
		if err != nil {
			return fmt.Errorf("unable to add github pull request labels: %w", err)
		}
		defer res.Body.Close()
	}

	for _, existingPullRequest := range existingPullRequests {
		LogDebug("Commenting on superseded pull request %d to github repository %s/%s", *existingPullRequest.Number, *ownerName, *repoName)
		body := fmt.Sprintf("Superseded by #%d", *pullRequest.Number)
		_, res, err = client.Issues.CreateComment(context.Background(), *ownerName, *repoName, *existingPullRequest.Number, &github.IssueComment{
			Body: &body,
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

func (p *GitHubGitProvider) ExtractInheritedLabels(pullRequests []*github.PullRequest) []string {
	result := []string{}
	if p.InheritLabels.Enabled {
		allLabels := SliceUnique(SliceFlatMap(pullRequests, func(pr *github.PullRequest) []string {
			return SliceMap(pr.Labels, func(l *github.Label) string {
				return *l.Name
			})
		}))
		for _, label := range allLabels {
			included := true
			excluded := false
			if len(p.InheritLabels.Includes) > 0 {
				included = false
				for _, pattern := range p.InheritLabels.Includes {
					matches, err := filepath.Match(pattern, label)
					if err == nil && matches {
						included = true
						break
					}
				}
			}
			if len(p.InheritLabels.Excludes) > 0 {
				for _, pattern := range p.InheritLabels.Excludes {
					matches, err := filepath.Match(pattern, label)
					if err == nil && matches {
						excluded = true
						break
					}
				}
			}
			if included && !excluded {
				result = append(result, label)
			}
		}
	}
	return result
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
