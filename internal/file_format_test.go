package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlFileFormatExtractAnnotations(t *testing.T) {
	f := YamlFileFormat{}
	test := func(input []string, expected []FileFormatAnnotation) {
		output, err := f.ExtractAnnotations(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	}

	test(strings.Split("", "\n"), []FileFormatAnnotation{})
	test(strings.Split("key: value", "\n"), []FileFormatAnnotation{})
	test(strings.Split("key: value\n\n---\n\nkey2: value2", "\n"), []FileFormatAnnotation{})
	test(strings.Split("key: value # git-ops-update {}\n\n---\n\nkey2: value2 # git-ops-update {\"other\":1}", "\n"), []FileFormatAnnotation{
		{LineNum: 1, AnnotationRaw: "git-ops-update {}"},
		{LineNum: 5, AnnotationRaw: "git-ops-update {\"other\":1}"},
	})
	test(strings.Split(`
foo1: 1 # ignored
foo2: |
  {
    "embedded": "json" # ignored
  }
foo3: | # ignored
  a
  b
foo4: > # ignored
  a
  b
bar1: value # working 1
bar2: 'value' # working 2
bar3: "value" # working 3
`, "\n"), []FileFormatAnnotation{
		{LineNum: 13, AnnotationRaw: "working 1"},
		{LineNum: 14, AnnotationRaw: "working 2"},
		{LineNum: 15, AnnotationRaw: "working 3"},
	})
}

func TestYamlFileFormatReadValue(t *testing.T) {
	f := YamlFileFormat{}
	test := func(input string, expected string) {
		lines := strings.Split(input, "\n")
		output, err := f.ReadValue(lines, 1)
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
		lines := strings.Split(input, "\n")
		err := f.WriteValue(lines, 1, value)
		assert.NoError(t, err)
		output := strings.Join(lines, "\n")
		assert.Equal(t, expected, output)
	}
	test("key: \"23\"", "new", "key: \"new\"")
	test("key: value", "new", "key: new")
	test("  key: value", "new", "  key: new")
	test("  - key: value", "new", "  - key: new")
	test("    - key: value", "new", "    - key: new")
}
