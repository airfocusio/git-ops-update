package internal

import (
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

func TestTagFormat(t *testing.T) {
	format := TagFormat{}

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
