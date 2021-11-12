package internal

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

type GitOpsUpdaterCacheResource struct {
	RegistryName string    `yaml:"registry"`
	ResourceName string    `yaml:"resource"`
	Versions     []string  `yaml:"versions"`
	Timestamp    time.Time `yaml:"timestamp"`
}

type GitOpsUpdaterCache struct {
	Resources []GitOpsUpdaterCacheResource `yaml:"resources"`
}

func LoadGitOpsUpdaterCache(bytes []byte) (*GitOpsUpdaterCache, error) {
	cache := &GitOpsUpdaterCache{}
	err := yaml.Unmarshal(bytes, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func SaveGitOpsUpdaterCache(cache GitOpsUpdaterCache) (*[]byte, error) {
	bytes, err := yaml.Marshal(cache)
	if err != nil {
		return nil, err
	}
	return &bytes, nil
}

func (c1 GitOpsUpdaterCache) Equal(c2 GitOpsUpdaterCache) bool {
	return true &&
		cmp.Equal(c1.Resources, c2.Resources)
}

func (c GitOpsUpdaterCache) UpdateResource(r GitOpsUpdaterCacheResource) GitOpsUpdaterCache {
	for i, r2 := range c.Resources {
		if r2.RegistryName == r.RegistryName && r2.ResourceName == r.ResourceName {
			return GitOpsUpdaterCache{
				Resources: append(c.Resources[:i], append([]GitOpsUpdaterCacheResource{r}, c.Resources[i+1:]...)...),
			}
		}
	}

	return GitOpsUpdaterCache{
		Resources: append(c.Resources, r),
	}
}

func (c GitOpsUpdaterCache) FindResource(registryName string, resourceName string) *GitOpsUpdaterCacheResource {
	for _, r := range c.Resources {
		if r.RegistryName == registryName && r.ResourceName == resourceName {
			return &r
		}
	}
	return nil
}
