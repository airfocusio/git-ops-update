package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubAugmenterRenderMessage(t *testing.T) {
	c := Config{
		Registries: map[string]Registry{
			"docker": DockerRegistry{
				Url: "https://ghcr.io",
			},
		},
	}
	a := GithubAugmenter{}

	t.Run("only second labeled", func(t *testing.T) {
		m1, m2, err := a.RenderMessage(c, Change{
			RegistryName: "docker",
			ResourceName: "airfocusio/git-ops-update-test",
			OldVersion:   "docker-v2-manifest-v0.0.0",
			NewVersion:   "docker-v2-manifest-v0.0.2",
		})
		if assert.NoError(t, err) {
			assert.Equal(
				t,
				"[Commit](https://github.com/airfocusio/git-ops-update/commit/7f4304f54fd1f89aecc15ec3e70975e5bacfeb68)",
				m1,
			)
			assert.Equal(t, "", m2)
		}
	})

	t.Run("both labeled", func(t *testing.T) {
		m1, m2, err := a.RenderMessage(c, Change{
			RegistryName: "docker",
			ResourceName: "airfocusio/git-ops-update-test",
			OldVersion:   "docker-v2-manifest-v0.0.1",
			NewVersion:   "docker-v2-manifest-v0.0.2",
		})
		if assert.NoError(t, err) {
			assert.Equal(
				t,
				"[Compare](https://github.com/airfocusio/git-ops-update/compare/99a7aaa2eff5aff7c77d9caf0901454fedf6bf00...7f4304f54fd1f89aecc15ec3e70975e5bacfeb68)\n"+
					"\n"+
					"Pull requests\n"+
					"\n"+
					"* [aggregate errors](https://github.com/airfocusio/git-ops-update/pull/1)\n"+
					"* [Improve error handling and logging](https://github.com/airfocusio/git-ops-update/pull/2)\n"+
					"\n"+
					"Commits\n"+
					"\n"+
					"* [add file1.txt](https://github.com/airfocusio/git-ops-update/commit/d71b2804a1c736d85e25d24739a2c5a67946b628)\n"+
					"* [add file2.txt v0.0.2](https://github.com/airfocusio/git-ops-update/commit/7f4304f54fd1f89aecc15ec3e70975e5bacfeb68)",
				m1,
			)
			assert.Equal(t, "/cc @choffmeister, @DizTortion", m2)
		}
	})
}

func TestGithubAugmenterExtractPullRequestNumbers(t *testing.T) {
	a := GithubAugmenter{}

	t.Run("empty", func(t *testing.T) {
		text, prs := a.ExtractPullRequestNumbers("")
		assert.Equal(t, "", text)
		assert.Equal(t, []int{}, prs)
	})

	t.Run("none", func(t *testing.T) {
		text, prs := a.ExtractPullRequestNumbers("Hello World")
		assert.Equal(t, "Hello World", text)
		assert.Equal(t, []int{}, prs)
	})

	t.Run("single", func(t *testing.T) {
		text, prs := a.ExtractPullRequestNumbers("Hello World #1")
		assert.Equal(t, "Hello World", text)
		assert.Equal(t, []int{1}, prs)
	})

	t.Run("single in parentheses", func(t *testing.T) {
		text, prs := a.ExtractPullRequestNumbers("Hello World (#1)")
		assert.Equal(t, "Hello World", text)
		assert.Equal(t, []int{1}, prs)
	})

	t.Run("multiple", func(t *testing.T) {
		text, prs := a.ExtractPullRequestNumbers("Hello World (#1)\n\nAnother one (#234)")
		assert.Equal(t, "Hello World\n\nAnother one", text)
		assert.Equal(t, []int{1, 234}, prs)
	})
}
