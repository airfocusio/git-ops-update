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
	m, err := a.RenderMessage(c, Change{
		RegistryName: "docker",
		ResourceName: "airfocusio/git-ops-update-test",
		OldVersion:   "docker-v2-manifest-v0.0.1",
		NewVersion:   "docker-v2-manifest-v0.0.2",
	})
	if assert.NoError(t, err) {
		assert.Equal(
			t,
			"https://github.com/airfocusio/git-ops-update/compare/99a7aaa2eff5aff7c77d9caf0901454fedf6bf00...700688d22e0b5c1ce02d341c42d0bf7c11e80aaa\n"+
				"\n"+
				"* add file1.txt https://github.com/airfocusio/git-ops-update/commit/0246e72c58f839d09ff0975f80955c0666f168ca\n"+
				"* add file2.txt v0.0.2 https://github.com/airfocusio/git-ops-update/commit/700688d22e0b5c1ce02d341c42d0bf7c11e80aaa",
			m,
		)
	}
}
