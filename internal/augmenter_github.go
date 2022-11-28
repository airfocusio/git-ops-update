package internal

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

var _ Augmenter = (*GithubAugmenter)(nil)

type GithubAugmenter struct {
	AccessToken string
}

func (a GithubAugmenter) RenderMessage(config Config, change Change) (string, error) {
	registry, ok := config.Registries[change.RegistryName]
	if !ok {
		return "", nil
	}
	dockerRegistry, ok := registry.(DockerRegistry)
	if !ok {
		return "", nil
	}

	oldLabels, err := dockerRegistry.RetrieveLabels(change.ResourceName, change.OldVersion)
	if err != nil {
		return "", err
	}
	newLabels, err := dockerRegistry.RetrieveLabels(change.ResourceName, change.NewVersion)
	if err != nil {
		return "", err
	}

	oldSource := oldLabels["org.opencontainers.image.source"]
	oldRevision := oldLabels["org.opencontainers.image.revision"]
	newSource := newLabels["org.opencontainers.image.source"]
	newRevision := newLabels["org.opencontainers.image.revision"]

	githubSourceRegex := regexp.MustCompile(`^https://github.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)$`)
	oldGithubSourceMatch := githubSourceRegex.FindStringSubmatch(oldSource)
	newGithubSourceMatch := githubSourceRegex.FindStringSubmatch(newSource)
	githubCommitRegex := regexp.MustCompile(`^(?P<hash>[0-9a-f]{40})$`)
	oldGithubCommitMatch := githubCommitRegex.FindStringSubmatch(oldRevision)
	newGithubCommitMatch := githubCommitRegex.FindStringSubmatch(newRevision)

	if newGithubSourceMatch != nil && oldGithubSourceMatch != nil && newGithubSourceMatch[2] == oldGithubSourceMatch[2] && newGithubSourceMatch[1] == oldGithubSourceMatch[1] && newGithubCommitMatch != nil && oldGithubCommitMatch != nil {
		owner := newGithubSourceMatch[1]
		repo := newGithubSourceMatch[2]
		base := oldGithubCommitMatch[1]
		head := newGithubCommitMatch[1]

		result := fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s", owner, repo, base, head) + "\n\n"

		ctx := context.Background()
		client := github.NewClient(&http.Client{})
		if a.AccessToken != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.AccessToken})
			tc := oauth2.NewClient(ctx, ts)
			client = github.NewClient(tc)
		}
		comparison, res, err := client.Repositories.CompareCommits(ctx, owner, repo, base, head, &github.ListOptions{
			PerPage: 100,
		})
		defer res.Body.Close()
		if err == nil && comparison != nil {
			for _, commit := range comparison.Commits {
				if commit.Commit != nil && commit.HTMLURL != nil && commit.Commit.Message != nil {
					result = result + fmt.Sprintf("* %s %s", *commit.Commit.Message, *commit.HTMLURL) + "\n"
				}
			}
		}

		return strings.Trim(result, "\n "), nil
	} else if newGithubSourceMatch != nil && newGithubCommitMatch != nil {
		owner := newGithubSourceMatch[1]
		repo := newGithubSourceMatch[2]
		head := newGithubCommitMatch[1]
		return fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, head), nil
	}

	return "", nil
}
