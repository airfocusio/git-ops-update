package internal

import (
	utiljson "encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func visitAnnotationsRecursion(parentNodes []*yaml.Node, node *yaml.Node, identifier string, fn func(keyNode *yaml.Node, valueNode *yaml.Node, parentNodes []*yaml.Node, annotation string) error) error {
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
					err := fn(keyNode, valueNode, []*yaml.Node{}, annotation)
					if err != nil {
						return err
					}
				}
			} else {
				err := visitAnnotationsRecursion(append(parentNodes, node), valueNode, identifier, fn)
				if err != nil {
					return err
				}
			}
		}
	} else {
		for _, child := range node.Content {
			err := visitAnnotationsRecursion(append(parentNodes, node), child, identifier, fn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func VisitAnnotations(node *yaml.Node, identifier string, fn func(keyNode *yaml.Node, valueNode *yaml.Node, parentNodes []*yaml.Node, annotation string) error) error {
	return visitAnnotationsRecursion([]*yaml.Node{}, node, identifier, fn)
}
