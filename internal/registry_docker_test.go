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
	assert.Greater(t, len(*output1), 0)

	reg2 := DockerRegistry{
		Url: "https://quay.io",
	}

	output2, err := reg2.FetchVersions("oauth2-proxy/oauth2-proxy")
	assert.NoError(t, err)
	assert.Greater(t, len(*output2), 0)
}
