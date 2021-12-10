package internal

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetectUpdates(t *testing.T) {
	bytes, err := ioutil.ReadFile("../example/.git-ops-update.yaml")
	if assert.NoError(t, err) {
		cache := Cache{
			Resources: []CacheResource{
				{
					RegistryName: "my-docker-registry",
					ResourceName: "library/nginx",
					Versions:     []string{"1.19.0-alpine", "1.19.1-alpine", "1.20.0-alpine", "1.20.1-alpine", "1.21.0-alpine", "1.21.1-alpine"},
					Timestamp:    time.Now(),
				},
				{
					RegistryName: "my-helm-registry",
					ResourceName: "nginx-ingress",
					Versions:     []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.11.0", "0.11.1"},
					Timestamp:    time.Now(),
				},
				{
					RegistryName: "my-git-hub-tag-registry",
					ResourceName: "kubernetes/ingress-nginx",
					Versions:     []string{"controller-v1.0.0", "controller-v1.0.10", "controller-v1.1.0"},
					Timestamp:    time.Now(),
				},
			},
		}
		cacheProvider := MemoryCacheProvider{Cache: &cache}
		if assert.NoError(t, err) {
			config, err := LoadConfig(bytes)
			if assert.NoError(t, err) {
				changes, errs := DetectUpdates("../example", *config, &cacheProvider)
				if assert.Len(t, changes, 3) {
					assert.Equal(t, "deployment.yaml", changes[0].File)
					assert.Equal(t, yamlTrace{"spec", "template", "spec", "containers", 0, "image"}, changes[0].Trace)
					assert.Equal(t, "1.19.0-alpine", changes[0].OldVersion)
					assert.Equal(t, "1.19.1-alpine", changes[0].NewVersion)
					assert.Equal(t, "nginx:1.19.0-alpine", changes[0].OldValue)
					assert.Equal(t, "nginx:1.19.1-alpine", changes[0].NewValue)

					assert.Equal(t, "helm-release.yaml", changes[1].File)
					assert.Equal(t, yamlTrace{"spec", "chart", "spec", "version"}, changes[1].Trace)
					assert.Equal(t, "0.10.1", changes[1].OldVersion)
					assert.Equal(t, "0.10.3", changes[1].NewVersion)
					assert.Equal(t, "0.10.1", changes[1].OldValue)
					assert.Equal(t, "0.10.3", changes[1].NewValue)

					assert.Equal(t, "kustomization.yaml", changes[2].File)
					assert.Equal(t, yamlTrace{"bases", 0}, changes[2].Trace)
					assert.Equal(t, "controller-v1.0.0", changes[2].OldVersion)
					assert.Equal(t, "controller-v1.0.10", changes[2].NewVersion)
					assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.0.0", changes[2].OldValue)
					assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.0.10", changes[2].NewValue)
				}
				if assert.Len(t, errs, 2) {
					assert.EqualError(t, errs[0], `deployment.yaml:fail: annotation {"will":"fail1"} misses registry`)
					assert.EqualError(t, errs[1], `helm-release.yaml:fail: annotation {"will":"fail2"} misses registry`)
				}
			}
		}
	}
}
