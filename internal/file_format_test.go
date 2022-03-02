package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlFileFormatExtractLineComment(t *testing.T) {
	f := YamlFileFormat{}
	test := func(input string, expected string) {
		output, err := f.ExtractLineComment(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	}
	test("", "")
	test("#", "")
	test("# foobar", "")
	test("key: \"value\" # foobar", "foobar")
	test("  key: \"value\" # foobar", "foobar")
	test("  - key: \"value\" # foobar", "foobar")
	test("    - key: \"value\" # foobar", "foobar")
}

func TestYamlFileFormatReadValue(t *testing.T) {
	f := YamlFileFormat{}
	test := func(input string, expected string) {
		output, err := f.ReadValue(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	}
	test("", "")
	test("#", "")
	test("# foobar", "")
	test("key: \"23\"", "23")
	test("key: value", "value")
	test("  key: value", "value")
	test("  - key: value", "value")
	test("    - key: value", "value")
}

func TestYamlFileFormatWriteValue(t *testing.T) {
	f := YamlFileFormat{}
	test := func(input string, value string, expected string) {
		output, err := f.WriteValue(input, value)
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	}
	test("key: \"23\"", "new", "key: \"new\"")
	test("key: value", "new", "key: new")
	test("  key: value", "new", "  key: new")
	test("  - key: value", "new", "  - key: new")
	test("    - key: value", "new", "    - key: new")
}
