package internal

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetectUpdates(t *testing.T) {
	cache := Cache{
		Resources: []CacheResource{
			{
				RegistryName: "my-docker-registry",
				ResourceName: "airfocusio/git-ops-update-test",
				Versions:     []string{"docker-v2-manifest-v0.0.1", "docker-v2-manifest-v0.0.2"},
				Timestamp:    time.Now(),
			},
			{
				RegistryName: "my-helm-registry",
				ResourceName: "nginx-ingress",
				Versions:     []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.11.0", "0.11.1", "1.0.0"},
				Timestamp:    time.Now(),
			},
			{
				RegistryName: "my-git-hub-tag-registry",
				ResourceName: "kubernetes/ingress-nginx",
				Versions:     []string{"controller-v1.0.0", "controller-v1.0.10", "controller-v1.1.0", "controller-v2.0.0"},
				Timestamp:    time.Now(),
			},
		},
	}
	cacheProvider := MemoryCacheProvider{Cache: &cache}
	config := Config{
		Files: ConfigFiles{
			Includes: []regexp.Regexp{
				*regexp.MustCompile(`update_test_.*\.yaml$`),
			},
		},
		Registries: map[string]Registry{
			"my-docker-registry": DockerRegistry{
				Interval: time.Hour,
				Url:      "https://ghcr.io",
			},
			"my-helm-registry": HelmRegistry{
				Interval: time.Hour,
				Url:      "https://helm.nginx.com/stable",
			},
			"my-git-hub-tag-registry": GitHubTagRegistry{
				Interval: time.Hour,
			},
		},
		Policies: map[string]Policy{
			"my-semver-policy": {
				Pattern: regexp.MustCompile(`^v?(?P<version>.*)$`),
				Extracts: []Extract{
					{
						Value: "<version>",
						Strategy: SemverExtractStrategy{

							PinMajor: true,
							Relaxed:  true,
						},
					},
				},
			},
		},
	}

	result := DetectUpdates(".", config, &cacheProvider)
	if assert.Len(t, result, 6) {
		if assert.NotNil(t, result[0].Change) {
			change := result[0].Change
			assert.Equal(t, "update_test_deployment.yaml", change.File)
			assert.Equal(t, 19, change.LineNum)
			assert.Equal(t, "docker-v2-manifest-v0.0.1", change.OldVersion)
			assert.Equal(t, "docker-v2-manifest-v0.0.2", change.NewVersion)
			assert.Equal(t, "ghcr.io/airfocusio/git-ops-update-test:docker-v2-manifest-v0.0.1", change.OldValue)
			assert.Equal(t, "ghcr.io/airfocusio/git-ops-update-test:docker-v2-manifest-v0.0.2", change.NewValue)
		}

		if assert.NotNil(t, result[1].Change) {
			change := result[1].Change
			assert.Equal(t, "update_test_deployment.yaml", change.File)
			assert.Equal(t, 39, change.LineNum)
			assert.Equal(t, "docker-v2-manifest-v0.0.1", change.OldVersion)
			assert.Equal(t, "docker-v2-manifest-v0.0.2", change.NewVersion)
			assert.Equal(t, "ghcr.io/airfocusio/git-ops-update-test:docker-v2-manifest-v0.0.1", change.OldValue)
			assert.Equal(t, "ghcr.io/airfocusio/git-ops-update-test:docker-v2-manifest-v0.0.2", change.NewValue)
		}

		if assert.NotNil(t, result[2].Error) {
			assert.EqualError(t, result[2].Error, `update_test_deployment.yaml:7: annotation {"will":"fail1"} misses registry`)
		}

		if assert.NotNil(t, result[3].Change) {
			change := result[3].Change
			assert.Equal(t, "update_test_helm_release.yaml", change.File)
			assert.Equal(t, 13, change.LineNum)
			assert.Equal(t, "0.10.1", change.OldVersion)
			assert.Equal(t, "0.11.1", change.NewVersion)
			assert.Equal(t, "0.10.1", change.OldValue)
			assert.Equal(t, "0.11.1", change.NewValue)
		}

		if assert.NotNil(t, result[4].Error) {
			assert.EqualError(t, result[4].Error, `update_test_helm_release.yaml:5: annotation {"will":"fail2"} misses registry`)
		}

		if assert.NotNil(t, result[5].Change) {
			change := result[5].Change
			assert.Equal(t, "update_test_kustomization.yaml", change.File)
			assert.Equal(t, 3, change.LineNum)
			assert.Equal(t, "controller-v1.0.0", change.OldVersion)
			assert.Equal(t, "controller-v1.1.0", change.NewVersion)
			assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.0.0", change.OldValue)
			assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.1.0", change.NewValue)
		}
	}
}
