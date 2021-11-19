package internal

import (
	"io/ioutil"
	"time"

	"github.com/google/go-cmp/cmp"
)

type CacheResource struct {
	RegistryName string    `yaml:"registry"`
	ResourceName string    `yaml:"resource"`
	Versions     []string  `yaml:"versions"`
	Timestamp    time.Time `yaml:"timestamp"`
}

type Cache struct {
	Resources []CacheResource `yaml:"resources"`
}

type CacheProvider interface {
	Load() (*Cache, error)
	Save(cache Cache) error
}

func (c1 Cache) Equal(c2 Cache) bool {
	return true &&
		cmp.Equal(c1.Resources, c2.Resources)
}

func (c Cache) UpdateResource(r CacheResource) Cache {
	for i, r2 := range c.Resources {
		if r2.RegistryName == r.RegistryName && r2.ResourceName == r.ResourceName {
			return Cache{
				Resources: append(c.Resources[:i], append([]CacheResource{r}, c.Resources[i+1:]...)...),
			}
		}
	}

	return Cache{
		Resources: append(c.Resources, r),
	}
}

func (c Cache) FindResource(registryName string, resourceName string) *CacheResource {
	for _, r := range c.Resources {
		if r.RegistryName == registryName && r.ResourceName == resourceName {
			return &r
		}
	}
	return nil
}

type FileCacheProvider struct {
	File string
}

func (p FileCacheProvider) Load() (*Cache, error) {
	cacheRaw, err := ioutil.ReadFile(p.File)
	if err != nil {
		return &Cache{}, nil
	}
	cache, err := loadCache(cacheRaw)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (p FileCacheProvider) Save(cache Cache) error {
	cacheBytesOut, err := saveCache(cache)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(p.File, *cacheBytesOut, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadCache(bytes []byte) (*Cache, error) {
	cache := &Cache{}
	err := readYaml(bytes, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func saveCache(cache Cache) (*[]byte, error) {
	bytes, err := writeYaml(cache)
	if err != nil {
		return nil, err
	}
	return &bytes, nil
}
