package internal

import (
	"testing"

	"github.com/google/go-github/v40/github"
	"github.com/stretchr/testify/assert"
)

func TestGithubGitProviderExtractInheritedLabels(t *testing.T) {

	testCases := []struct {
		config               GitHubGitProviderInheritLabels
		inputLabels          []string
		expectedOutputLabels []string
	}{
		{
			config:               GitHubGitProviderInheritLabels{},
			inputLabels:          []string{},
			expectedOutputLabels: []string{},
		},
		{
			config:               GitHubGitProviderInheritLabels{},
			inputLabels:          []string{"tag:1", "tag:2"},
			expectedOutputLabels: []string{},
		},
		{
			config:               GitHubGitProviderInheritLabels{Enabled: true},
			inputLabels:          []string{"tag:1", "tag:2"},
			expectedOutputLabels: []string{"tag:1", "tag:2"},
		},
		{
			config:               GitHubGitProviderInheritLabels{Enabled: true},
			inputLabels:          []string{"tag:1", "tag:2", "tag:1", "tag:2"},
			expectedOutputLabels: []string{"tag:1", "tag:2"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Includes: []string{"tag:1"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"tag:1"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Includes: []string{"tag:1", "note:a"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"tag:1", "note:a"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Excludes: []string{"tag:1"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"tag:2", "note:a", "note:b"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Excludes: []string{"tag:1", "note:a"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"tag:2", "note:b"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Includes: []string{"tag:*"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"tag:1", "tag:2"},
		},
		{
			config: GitHubGitProviderInheritLabels{
				Enabled:  true,
				Excludes: []string{"tag:*"},
			},
			inputLabels:          []string{"tag:1", "tag:2", "note:a", "note:b"},
			expectedOutputLabels: []string{"note:a", "note:b"},
		},
	}

	for _, tc := range testCases {
		p := GitHubGitProvider{InheritLabels: tc.config}
		prs := []*github.PullRequest{
			{},
			{
				Labels: SliceMap(tc.inputLabels, func(name string) *github.Label {
					return &github.Label{
						Name: &name,
					}
				}),
			},
		}
		assert.Equal(t, tc.expectedOutputLabels, p.ExtractInheritedLabels(prs))
	}
}
