package internal

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

var _ Augmenter = (*GithubAugmenter)(nil)

type GithubAugmenter struct {
	AccessToken string
}

func (a GithubAugmenter) RenderMessage(config Config, change Change) (string, string, error) {
	registry, ok := config.Registries[change.RegistryName]
	if !ok {
		return "", "", nil
	}
	dockerRegistry, ok := registry.(DockerRegistry)
	if !ok {
		return "", "", nil
	}

	oldLabels, err := dockerRegistry.RetrieveLabels(change.ResourceName, change.OldVersion)
	if err != nil {
		return "", "", err
	}
	newLabels, err := dockerRegistry.RetrieveLabels(change.ResourceName, change.NewVersion)
	if err != nil {
		return "", "", err
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

		result1 := GithubLink{
			Title: "Compare",
			URL:   fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s", owner, repo, base, head),
		}.Render() + "\n"
		mentions := []string{}
		if err == nil && comparison != nil {
			pullRequests := []GithubLink{}
			commits := []GithubLink{}
			for _, commit := range comparison.Commits {
				if commit.Author != nil {
					mentions = append(mentions, *commit.Author.Login)
				}
				if commit.Commit != nil && commit.HTMLURL != nil {
					if commit.Commit.Message != nil {
						message, pullRequestNumbers := a.ExtractPullRequestNumbers(*commit.Commit.Message)
						for _, pullRequestNumber := range pullRequestNumbers {
							pullRequest, res, err := client.PullRequests.Get(ctx, owner, repo, pullRequestNumber)
							defer res.Body.Close()
							if pullRequest != nil && err == nil {
								if pullRequest.User != nil {
									mentions = append(mentions, *pullRequest.User.Login)
								}
								pullRequests = append(pullRequests, GithubLink{
									Title: *pullRequest.Title,
									URL:   *pullRequest.HTMLURL,
								})
							} else {
								pullRequests = append(pullRequests, GithubLink{
									Title: "???",
									URL:   fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pullRequestNumber),
								})
							}
						}
						title := strings.Trim(strings.Split(message, "\n")[0], " ")
						if title != "" {
							commits = append(commits, GithubLink{
								Title: title,
								URL:   *commit.HTMLURL,
							})
						}
					}
				}
			}
			if len(pullRequests) > 0 {
				result1 = result1 + "\nPull requests\n\n"
				for _, pr := range pullRequests {
					result1 = result1 + fmt.Sprintf("* %s", pr.Render()) + "\n"
				}
			}
			if len(commits) > 0 {
				result1 = result1 + "\nCommits\n\n"
				for _, c := range commits {
					result1 = result1 + fmt.Sprintf("* %s", c.Render()) + "\n"
				}
			}
		}
		result1 = result1 + "\n"
		result1 = strings.Trim(result1, "\n ")

		result2 := ""
		mentions = sliceFilter(sliceUnique(mentions), func(mention string) bool {
			return !strings.HasSuffix(mention, "[bot]")
		})
		if len(mentions) > 0 {
			result2 = result2 + "/cc " + strings.Join(sliceMap(mentions, func(mention string) string {
				return "@" + mention
			}), ", ")
			result2 = result2 + "\n"
		}
		result2 = strings.Trim(result2, "\n ")

		return result1, result2, nil
	} else if newGithubSourceMatch != nil && newGithubCommitMatch != nil {
		owner := newGithubSourceMatch[1]
		repo := newGithubSourceMatch[2]
		head := newGithubCommitMatch[1]
		return GithubLink{
			Title: "Commit",
			URL:   fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, head),
		}.Render(), "", nil
	}

	return "", "", nil
}

func (a GithubAugmenter) ExtractPullRequestNumbers(text string) (string, []int) {
	pullRequestNumberRegex := regexp.MustCompile(`\(?#(?P<number>\d+)\)?`)
	pullRequestNumbers := []int{}
	matches := pullRequestNumberRegex.FindAllStringSubmatch(text, 100)
	for _, m := range matches {
		if pullRequestNumber, err := strconv.Atoi(m[1]); err == nil {
			pullRequestNumbers = append(pullRequestNumbers, pullRequestNumber)
		}
	}
	return trimRightMultilineString(pullRequestNumberRegex.ReplaceAllLiteralString(text, ""), " "), pullRequestNumbers
}

type GithubLink struct {
	Title string
	URL   string
}

func (l GithubLink) Render() string {
	if l.Title != "" {
		return fmt.Sprintf("[%s](%s)", l.Title, l.URL)
	}
	return l.URL
}
