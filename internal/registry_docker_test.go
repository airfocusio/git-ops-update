package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerFetchVersions(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}
	output, err := reg.FetchVersions("airfocusio/git-ops-update-test")
	assert.NoError(t, err)
	assert.Greater(t, len(output), 0)

	assert.Contains(t, output, "docker-v2-manifest-v0.0.1")
	assert.Contains(t, output, "docker-v2-manifest-v0.0.2")
	assert.Contains(t, output, "docker-v2-manifest-list-v0.0.1")
	assert.Contains(t, output, "docker-v2-manifest-list-v0.0.2")
	assert.Contains(t, output, "oci-v1-manifest-v0.0.1")
	assert.Contains(t, output, "oci-v1-manifest-v0.0.2")
	assert.Contains(t, output, "oci-v1-image-index-v0.0.1")
	assert.Contains(t, output, "oci-v1-image-index-v0.0.2")
}

func TestDockerRetrieveLabels(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}

	t.Run("docker-v2-manifest", func(t *testing.T) {
		output, err := reg.RetrieveLabels("airfocusio/git-ops-update-test", "docker-v2-manifest-v0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.1", output["org.opencontainers.image.version"])
		assert.Equal(t, "https://github.com/airfocusio/git-ops-update", output["org.opencontainers.image.source"])
		assert.Equal(t, "99a7aaa2eff5aff7c77d9caf0901454fedf6bf00", output["org.opencontainers.image.revision"])
	})

	t.Run("oci-v1-manifest", func(t *testing.T) {
		output, err := reg.RetrieveLabels("airfocusio/git-ops-update-test", "oci-v1-manifest-v0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.1", output["org.opencontainers.image.version"])
		assert.Equal(t, "https://github.com/airfocusio/git-ops-update", output["org.opencontainers.image.source"])
		assert.Equal(t, "99a7aaa2eff5aff7c77d9caf0901454fedf6bf00", output["org.opencontainers.image.revision"])
	})

	t.Run("docker-v2-manifest-list", func(t *testing.T) {
		output, err := reg.RetrieveLabels("airfocusio/git-ops-update-test", "docker-v2-manifest-list-v0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.1", output["org.opencontainers.image.version"])
		assert.Equal(t, "https://github.com/airfocusio/git-ops-update", output["org.opencontainers.image.source"])
		assert.Equal(t, "99a7aaa2eff5aff7c77d9caf0901454fedf6bf00", output["org.opencontainers.image.revision"])
	})

	t.Run("oci-v1-image-index", func(t *testing.T) {
		output, err := reg.RetrieveLabels("airfocusio/git-ops-update-test", "oci-v1-image-index-v0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.1", output["org.opencontainers.image.version"])
		assert.Equal(t, "https://github.com/airfocusio/git-ops-update", output["org.opencontainers.image.source"])
		assert.Equal(t, "99a7aaa2eff5aff7c77d9caf0901454fedf6bf00", output["org.opencontainers.image.revision"])
	})
}
