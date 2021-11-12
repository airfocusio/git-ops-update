package internal

import (
	utiljson "encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func visitAnnotationsRecursion(trace []string, node *yaml.Node, identifier string, fn func(keyNode *yaml.Node, valueNode *yaml.Node, trace []string, annotation string) error) error {
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			if valueNode.Kind == yaml.ScalarNode {
				annotationStr := strings.TrimPrefix(valueNode.LineComment, "#")
				annotationJson := map[string]interface{}{}
				utiljson.Unmarshal([]byte(annotationStr), &annotationJson)
				annotation, ok := annotationJson["$"+identifier].(string)
				if ok {
					err := fn(keyNode, valueNode, append(trace, keyNode.Value), annotation)
					if err != nil {
						return err
					}
				}
			} else {
				err := visitAnnotationsRecursion(append(trace, keyNode.Value), valueNode, identifier, fn)
				if err != nil {
					return err
				}
			}
		}
	} else {
		for _, child := range node.Content {
			err := visitAnnotationsRecursion(trace, child, identifier, fn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func VisitAnnotations(node *yaml.Node, identifier string, fn func(keyNode *yaml.Node, valueNode *yaml.Node, trace []string, annotation string) error) error {
	return visitAnnotationsRecursion([]string{}, node, identifier, fn)
}
