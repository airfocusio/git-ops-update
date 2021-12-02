package internal

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainFormat(t *testing.T) {
	format := PlainFormat{}

	actual, err := format.ExtractVersion("")
	if assert.NoError(t, err) {
		assert.Equal(t, "", *actual)
	}

	actual, err = format.ExtractVersion("any")
	if assert.NoError(t, err) {
		assert.Equal(t, "any", *actual)
	}

	actual, err = format.ReplaceVersion("any", "thing")
	if assert.NoError(t, err) {
		assert.Equal(t, "thing", *actual)
	}
}

func TestDockerImageFormat(t *testing.T) {
	format := DockerImageFormat{}

	_, err := format.ExtractVersion("")
	assert.Error(t, err)

	_, err = format.ExtractVersion("any")
	assert.Error(t, err)

	actual, err := format.ExtractVersion("image:version")
	if assert.NoError(t, err) {
		assert.Equal(t, "version", *actual)
	}

	_, err = format.ExtractVersion("image:version:more")
	assert.Error(t, err)

	actual, err = format.ReplaceVersion("image:version", "next")
	if assert.NoError(t, err) {
		assert.Equal(t, "image:next", *actual)
	}
}

func TestRegexpFormatTest(t *testing.T) {
	format := RegexpFormat{Pattern: *regexp.MustCompile(`^https://domain\.com/(?P<version>[^/]+)/dist$`)}
	format2 := RegexpFormat{Pattern: *regexp.MustCompile(`^https://domain\.com/(?P<version>[^/]+)/dist/(?P<version>[^/]+).tar$`)}
	format3 := RegexpFormat{Pattern: *regexp.MustCompile(`^https://(?P<domain>[^/]+)/(?P<version>[^/]+)/dist(?:/(?P<version>[^/]+).tar)?$`)}
	format4 := RegexpFormat{Pattern: *regexp.MustCompile(`(?P<version>\d+\.\d+\.\d+)`)}

	_, err := format.ExtractVersion("")
	assert.Error(t, err)

	_, err = format.ExtractVersion("any")
	assert.Error(t, err)

	actual, err := format.ExtractVersion("https://domain.com/1.2.3/dist")
	if assert.NoError(t, err) {
		assert.Equal(t, "1.2.3", *actual)
	}
	actual, err = format2.ExtractVersion("https://domain.com/1.2.3/dist/1.2.4.tar")
	if assert.NoError(t, err) {
		assert.Equal(t, "1.2.3", *actual)
	}
	actual, err = format3.ExtractVersion("https://other.com/1.2.3/dist/1.2.4.tar")
	if assert.NoError(t, err) {
		assert.Equal(t, "1.2.3", *actual)
	}
	actual, err = format4.ExtractVersion("foo-1.2.3-bar-1.2.4")
	if assert.NoError(t, err) {
		assert.Equal(t, "1.2.3", *actual)
	}
	_, err = format.ExtractVersion("https://domain.com/1.2.3/dist/foo")
	assert.Error(t, err)

	actual, err = format.ReplaceVersion("https://domain.com/1.2.3/dist", "1.2.10")
	if assert.NoError(t, err) {
		assert.Equal(t, "https://domain.com/1.2.10/dist", *actual)
	}
	actual, err = format2.ReplaceVersion("https://domain.com/1.2.3/dist/1.2.4.tar", "1.2.10")
	if assert.NoError(t, err) {
		assert.Equal(t, "https://domain.com/1.2.10/dist/1.2.10.tar", *actual)
	}
	actual, err = format3.ReplaceVersion("https://other.com/1.2.3/dist/1.2.4.tar", "1.2.10")
	if assert.NoError(t, err) {
		assert.Equal(t, "https://other.com/1.2.10/dist/1.2.10.tar", *actual)
	}
	actual, err = format3.ReplaceVersion("https://other.com/1.2.3/dist", "1.2.10")
	if assert.NoError(t, err) {
		assert.Equal(t, "https://other.com/1.2.10/dist", *actual)
	}
	actual, err = format4.ReplaceVersion("foo-1.2.3-bar-1.2.4", "1.2.10")
	if assert.NoError(t, err) {
		assert.Equal(t, "foo-1.2.10-bar-1.2.4", *actual)
	}
}
