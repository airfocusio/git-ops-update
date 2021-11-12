package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGitOpsUpdaterCache(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	yaml := `resources:
  - registry: docker
    resource: library/ubuntu
    versions:
      - "21.04"
      - "21.10"
    timestamp: 2006-01-02T15:04:05Z
`

	c1, err := LoadGitOpsUpdaterCache([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}
	c2 := GitOpsUpdaterCache{
		Resources: []GitOpsUpdaterCacheResource{
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
	yaml2, err := SaveGitOpsUpdaterCache(c2)
	if assert.NoError(t, err) {
		assert.Equal(t, yaml, string(*yaml2))
	}

	assert.Nil(t, c2.FindResource("unknown", "thing"))
	assert.Equal(t, &GitOpsUpdaterCacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10"},
		Timestamp:    ts,
	}, c2.FindResource("docker", "library/ubuntu"))

	c3 := c2.UpdateResource(GitOpsUpdaterCacheResource{RegistryName: "unknown", ResourceName: "thing", Timestamp: ts})
	assert.Equal(t, &GitOpsUpdaterCacheResource{
		RegistryName: "unknown",
		ResourceName: "thing",
		Timestamp:    ts,
	}, c3.FindResource("unknown", "thing"))

	c4 := c3.UpdateResource(GitOpsUpdaterCacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10", "22.04"},
		Timestamp:    ts,
	})
	assert.Equal(t, &GitOpsUpdaterCacheResource{
		RegistryName: "docker",
		ResourceName: "library/ubuntu",
		Versions:     []string{"21.04", "21.10", "22.04"},
		Timestamp:    ts,
	}, c4.FindResource("docker", "library/ubuntu"))
}