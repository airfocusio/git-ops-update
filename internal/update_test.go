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
				result := DetectUpdates("../example", *config, &cacheProvider)
				if assert.Len(t, result, 5) {
					if assert.NotNil(t, result[0].Change) {
						change := result[0].Change
						assert.Equal(t, "deployment.yaml", change.File)
						assert.Equal(t, yamlTrace{"spec", "template", "spec", "containers", 0, "image"}, change.Trace)
						assert.Equal(t, "1.19.0-alpine", change.OldVersion)
						assert.Equal(t, "1.19.1-alpine", change.NewVersion)
						assert.Equal(t, "nginx:1.19.0-alpine", change.OldValue)
						assert.Equal(t, "nginx:1.19.1-alpine", change.NewValue)
					}

					if assert.NotNil(t, result[1].Error) {
						assert.EqualError(t, result[1].Error, `deployment.yaml:fail: annotation {"will":"fail1"} misses registry`)
					}

					if assert.NotNil(t, result[2].Change) {
						change := result[2].Change
						assert.Equal(t, "helm-release.yaml", change.File)
						assert.Equal(t, yamlTrace{"spec", "chart", "spec", "version"}, change.Trace)
						assert.Equal(t, "0.10.1", change.OldVersion)
						assert.Equal(t, "0.10.3", change.NewVersion)
						assert.Equal(t, "0.10.1", change.OldValue)
						assert.Equal(t, "0.10.3", change.NewValue)
					}

					if assert.NotNil(t, result[3].Error) {
						assert.EqualError(t, result[3].Error, `helm-release.yaml:fail: annotation {"will":"fail2"} misses registry`)
					}

					if assert.NotNil(t, result[4].Change) {
						change := result[4].Change
						assert.Equal(t, "kustomization.yaml", change.File)
						assert.Equal(t, yamlTrace{"bases", 0}, change.Trace)
						assert.Equal(t, "controller-v1.0.0", change.OldVersion)
						assert.Equal(t, "controller-v1.0.10", change.NewVersion)
						assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.0.0", change.OldValue)
						assert.Equal(t, "github.com/kubernetes/ingress-nginx/deploy/static/provider/kind?ref=controller-v1.0.10", change.NewValue)
					}
				}
			}
		}
	}
}
