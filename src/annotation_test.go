package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestVisitAnnotations(t *testing.T) {
	input := `apple: pie
foo:
  bar:
    apple: pie # {"$append":"-new"}
    another: thing # {"$append":"-new2"}
`
	doc := yaml.Node{}
	err := yaml.Unmarshal([]byte(input), &doc)
	if assert.NoError(t, err) {
		err = VisitAnnotations(&doc, "append", func(keyNode *yaml.Node, valueNode *yaml.Node, parentNodes []*yaml.Node, annotation string) error {
			valueNode.Value = valueNode.Value + annotation
			return nil
		})
		if assert.NoError(t, err) {
			actualOutputBytes, err := yaml.Marshal(&doc)
			if assert.NoError(t, err) {
				actualOutput := string(actualOutputBytes)
				expectedOutput := `apple: pie
foo:
    bar:
        apple: pie-new # {"$append":"-new"}
        another: thing-new2 # {"$append":"-new2"}
`
				assert.Equal(t, expectedOutput, actualOutput)
			}
		}
	}
}
