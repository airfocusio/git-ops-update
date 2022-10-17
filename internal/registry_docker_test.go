package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerFetchVersions(t *testing.T) {
	reg1 := DockerRegistry{
		Url: "https://registry-1.docker.io",
	}
	output1, err := reg1.FetchVersions("library/nginx")
	assert.NoError(t, err)
	assert.Greater(t, len(output1), 0)

	reg2 := DockerRegistry{
		Url: "https://quay.io",
	}
	output2, err := reg2.FetchVersions("oauth2-proxy/oauth2-proxy")
	assert.NoError(t, err)
	assert.Greater(t, len(output2), 0)
}

func TestDockerRetrieveMetadata(t *testing.T) {
	reg1 := DockerRegistry{
		Url: "https://registry-1.docker.io",
	}
	output1, err := reg1.RetrieveMetadata("library/nginx", "1.23.0")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"maintainer": "NGINX Docker Maintainers <docker-maint@nginx.com>"}, output1)

	reg2 := DockerRegistry{
		Url: "https://quay.io",
	}
	output2, err := reg2.RetrieveMetadata("oauth2-proxy/oauth2-proxy", "v7.3.0")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{}, output2)
}

func TestDockerRetrieveDockerV2ManifestMetadata(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}
	output, err := reg.RetrieveMetadata("airfocusio/git-ops-update-test", "docker-v2-manifest-v0.0.1")
	assert.NoError(t, err)
	assert.Contains(t, output, "foo")
	assert.Equal(t, "bar", output["foo"])
}

func TestDockerRetrieveDockerV2ManifestListMetadata(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}
	output, err := reg.RetrieveMetadata("airfocusio/git-ops-update-test", "docker-v2-manifest-list-v0.0.1")
	assert.NoError(t, err)
	assert.Contains(t, output, "foo")
	assert.Equal(t, "bar", output["foo"])
}

func TestDockerRetrieveOciV1ManifestMetadata(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}
	output, err := reg.RetrieveMetadata("airfocusio/git-ops-update-test", "oci-v1-manifest-v0.0.1")
	assert.NoError(t, err)
	assert.Contains(t, output, "foo")
	assert.Equal(t, "bar", output["foo"])
}

func TestDockerRetrieveOciV1ImageIndexMetadata(t *testing.T) {
	reg := DockerRegistry{
		Url: "https://ghcr.io",
	}
	output, err := reg.RetrieveMetadata("airfocusio/git-ops-update-test", "oci-v1-image-index-v0.0.1")
	assert.NoError(t, err)
	assert.Contains(t, output, "foo")
	assert.Equal(t, "bar", output["foo"])
}
