package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/xanzy/go-gitlab"
)

var _ GitProvider = (*GitLabGitProvider)(nil)

type GitLabGitProvider struct {
	Author      GitAuthor
	URL         string
	AccessToken string
	AssigneeIDs []int
}

func (p GitLabGitProvider) Push(dir string, changes Changes, callbacks ...func() error) error {
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

func (p GitLabGitProvider) Request(dir string, changes Changes, callbacks ...func() error) error {
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
	projectId, err := extractGitLabProjectIdFromRemote(p.URL, *remote)
	if err != nil {
		return fmt.Errorf("unable to extract gitlab project id from remote origin: %w", err)
	}
	client, err := gitlab.NewOAuthClient(
		p.AccessToken,
		gitlab.WithBaseURL(p.URL))
	if err != nil {
		return fmt.Errorf("failed to connect to gitlab: %w", err)
	}

	existingBranches := []string{}
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

	LogDebug("Creating pull request for branch %s to gitlab project %s", targetBranch.Short(), *projectId)
	pullBase := string(baseBranch.Name().Short())
	pullHead := changes.Branch(branchPrefix)
	pullTitle := changes.Title()
	pullBody := fullMessage
	removeSourceBranch := true
	_, res, err := client.MergeRequests.CreateMergeRequest(*projectId, &gitlab.CreateMergeRequestOptions{
		Title:              &pullTitle,
		SourceBranch:       &pullHead,
		TargetBranch:       &pullBase,
		Description:        &pullBody,
		RemoveSourceBranch: &removeSourceBranch,
		AssigneeIDs:        &p.AssigneeIDs,
	})
	if err != nil {
		return fmt.Errorf("unable to create gitlab merge request: %w", err)
	}
	defer res.Body.Close()

	if err != nil {
		return fmt.Errorf("unable to create github pull request: %w", err)
	}
	defer res.Body.Close()

	for _, refName := range existingBranches {
		LogDebug("Removing branch %s from gitlab project %s", refName, *projectId)
		err := remote.Push(&git.PushOptions{
			Auth: &http.BasicAuth{
				Username: "api",
				Password: p.AccessToken,
			},
			RefSpecs: []config.RefSpec{
				config.RefSpec(":" + refName),
			},
		})
		if err != nil {
			LogWarning("Unable to remove branch %s from gitlab project %s: %v", refName, *projectId, err)
		}
	}

	err = worktree.Checkout(&git.CheckoutOptions{Branch: baseBranch.Name()})
	if err != nil {
		return fmt.Errorf("unable to checkout to base branch: %w", err)
	}
	return nil
}

func extractGitLabProjectIdFromRemote(baseURL string, remote git.Remote) (*string, error) {
	urlRegex := regexp.MustCompile(`^(?:(?:ssh|https?)\:\/\/)(?:[^@\/]+@)?[^\/]+\/(?P<projectid>.+).git\/?$`)
	scpRegex := regexp.MustCompile(`^(?:[^@\/]+@)?[^:]+:(?P<projectid>.+).git\/?$`)
	for _, url := range remote.Config().URLs {
		urlMatch := urlRegex.FindStringSubmatch(url)
		if urlMatch != nil {
			return &urlMatch[1], nil
		}
		scpMatch := scpRegex.FindStringSubmatch(url)
		if scpMatch != nil {
			return &scpMatch[1], nil
		}
	}

	return nil, fmt.Errorf("none of the git remote %s urls %v could be recognized as a gitlab project", remote.Config().Name, remote.Config().URLs)
}
