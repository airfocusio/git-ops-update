package internal

import (
	utiljson "encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func VisitAnnotations(node *yaml.Node, identifier string, fn func(trace yamlTrace, yamlNode *yaml.Node, annotation string) error) error {
	return VisitYaml(node, func(trace yamlTrace, yamlNode *yaml.Node) error {
		annotationStr := strings.TrimPrefix(yamlNode.LineComment, "#")
		annotationJson := map[string]interface{}{}
		utiljson.Unmarshal([]byte(annotationStr), &annotationJson)
		annotation, ok := annotationJson["$"+identifier].(string)
		if ok {
			err := fn(trace, yamlNode, annotation)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
