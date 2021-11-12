package internal

import (
	"testing"
	"time"
)

func TestGitOpsUpdaterCache(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	yaml := `resources:
- registry: docker
  resource: library/ubuntu
  versions:
  - 21.04
  - 21.10
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
}
