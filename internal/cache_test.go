package internal

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileCacheProvider(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "cache")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(file.Name())

	ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	yaml := `resources:
  - registry: docker
    resource: library/ubuntu
    versions:
      - "21.04"
      - "21.10"
    timestamp: 2006-01-02T15:04:05Z
`
	err = os.WriteFile(file.Name(), []byte(yaml), 0644)
	if err != nil {
		t.Error(err)
		return
	}

	cp := FileCacheProvider{File: file.Name()}
	c1, err := cp.Load()
	if err != nil {
		t.Error(err)
		return
	}
	c2 := Cache{
		Resources: []CacheResource{
			{
				RegistryName: "docker",
				ResourceName: "library/ubuntu",
				Versions:     []string{"21.04", "21.10"},
				Timestamp:    ts,
			},
		},
	}
	if !c1.Equal(c2) {
		t.Errorf("expected %v, got %v", c2, c1)
		return
	}
	err = cp.Save(c2)
	if err != nil {
		t.Error(err)
		return
	}
	yaml2, err := os.ReadFile(file.Name())
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, yaml, string(yaml2))

	assert.Nil(t, c2.FindResource("unknown", "thing"))
	assert.Equal(t, &CacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10"},
		Timestamp:    ts,
	}, c2.FindResource("docker", "library/ubuntu"))

	c3 := c2.UpdateResource(CacheResource{RegistryName: "unknown", ResourceName: "thing", Timestamp: ts})
	assert.Equal(t, &CacheResource{
		RegistryName: "unknown",
		ResourceName: "thing",
		Timestamp:    ts,
	}, c3.FindResource("unknown", "thing"))

	c4 := c3.UpdateResource(CacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10", "22.04"},
		Timestamp:    ts,
	})
	assert.Equal(t, &CacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10", "22.04"},
		Timestamp:    ts,
	}, c4.FindResource("docker", "library/ubuntu"))
}

type MemoryCacheProvider struct {
	File  string
	Cache *Cache
}

func (p MemoryCacheProvider) Load() (*Cache, error) {
	return p.Cache, nil
}

func (p MemoryCacheProvider) Save(cache Cache) error {
	return nil
}
