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

type CacheAction struct {
	Identifier string    `yaml:"identifier"`
	Timestamp  time.Time `yaml:"timestamp"`
}

type Cache struct {
	Resources []CacheResource `yaml:"resources"`
	Actions   []CacheAction   `yaml:"actions"`
}

func LoadCache(bytes []byte) (*Cache, error) {
	cache := &Cache{}
	err := readYaml(bytes, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func SaveCache(cache Cache) (*[]byte, error) {
	bytes, err := writeYaml(cache)
	if err != nil {
		return nil, err
	}
	return &bytes, nil
}

func LoadCacheFromFile(file string) (*Cache, error) {
	cacheRaw, err := ioutil.ReadFile(file)
	if err != nil {
		return &Cache{}, nil
	}
	cache, err := LoadCache(cacheRaw)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func SaveCacheToFile(cache Cache, file string) error {
	cacheBytesOut, err := SaveCache(cache)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, *cacheBytesOut, 0644)
	if err != nil {
		return err
	}
	return nil
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
				Actions:   c.Actions,
			}
		}
	}

	return Cache{
		Resources: append(c.Resources, r),
		Actions:   c.Actions,
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

func (c Cache) AddAction(identifier string) Cache {
	for i, a := range c.Actions {
		if a.Identifier == identifier {
			return Cache{
				Resources: c.Resources,
				Actions:   append(c.Actions[:i], append([]CacheAction{CacheAction{Identifier: identifier, Timestamp: time.Now()}}, c.Actions[i+1:]...)...),
			}
		}
	}

	return Cache{
		Resources: c.Resources,
		Actions:   append(c.Actions, CacheAction{Identifier: identifier, Timestamp: time.Now()}),
	}
}

func (c Cache) HasAction(identifier string) bool {
	for _, a := range c.Actions {
		if a.Identifier == identifier {
			return true
		}
	}
	return false
}
