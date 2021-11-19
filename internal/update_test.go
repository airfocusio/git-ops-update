package internal

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDetectUpdates(t *testing.T) {
	viperInstance := viper.New()
	viperInstance.SetConfigName(".git-ops-update")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath("../example")
	err := viperInstance.ReadInConfig()
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
			},
		}
		cacheProvider := MemoryCacheProvider{Cache: &cache}
		if assert.NoError(t, err) {
			config, err := LoadConfig(*viperInstance)
			if assert.NoError(t, err) {
				changes, err := DetectUpdates("../example", *config, &cacheProvider)
				if assert.NoError(t, err) {
					assert.Len(t, *changes, 2)

					assert.Equal(t, "deployment.yaml", (*changes)[0].File)
					assert.Equal(t, yamlTrace{"spec", "template", "spec", "containers", 0, "image"}, (*changes)[0].Trace)
					assert.Equal(t, "1.19.0-alpine", (*changes)[0].OldVersion)
					assert.Equal(t, "1.19.1-alpine", (*changes)[0].NewVersion)
					assert.Equal(t, "nginx:1.19.0-alpine", (*changes)[0].OldValue)
					assert.Equal(t, "nginx:1.19.1-alpine", (*changes)[0].NewValue)

					assert.Equal(t, "helm-release.yaml", (*changes)[1].File)
					assert.Equal(t, yamlTrace{"spec", "chart", "spec", "version"}, (*changes)[1].Trace)
					assert.Equal(t, "0.10.1", (*changes)[1].OldVersion)
					assert.Equal(t, "0.10.3", (*changes)[1].NewVersion)
					assert.Equal(t, "0.10.1", (*changes)[1].OldValue)
					assert.Equal(t, "0.10.3", (*changes)[1].NewValue)
				}
			}
		}
	}
}
