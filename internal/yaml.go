package internal

import (
	"bytes"
	"log"
	"strconv"

	"gopkg.in/yaml.v3"
)

type yamlTrace []interface{}

func (t1 yamlTrace) Equal(t2 yamlTrace) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := 0; i < len(t1); i++ {
		if t1[i] != t2[i] {
			return false
		}
	}
	return true
}

func readYaml(bytes []byte, v interface{}) error {
	return yaml.Unmarshal(bytes, v)
}

func writeYaml(v interface{}) ([]byte, error) {
	var bs bytes.Buffer
	enc := yaml.NewEncoder(&bs)
	enc.SetIndent(2)
	err := enc.Encode(v)
	return bs.Bytes(), err
}

func yamlTraceToString(trace yamlTrace) string {
	traceStr := ""
	for _, e := range trace {
		s, ok := e.(string)
		if ok {
			if len(traceStr) == 0 {
				traceStr = s
			} else {
				traceStr = traceStr + "." + s
			}
		} else {
			i, ok := e.(int)
			if ok {
				traceStr = traceStr + "." + strconv.Itoa(i)
			} else {
				log.Fatalf("unexpected incompatible yaml trace")
			}
		}
	}
	return traceStr
}

func visitYamlRecursion(trace yamlTrace, yamlNode *yaml.Node, fn func(trace yamlTrace, yamlNode *yaml.Node) error) error {
	if yamlNode.Kind == yaml.MappingNode {
		for i := 0; i < len(yamlNode.Content); i += 2 {
			keyNode := yamlNode.Content[i]
			valueNode := yamlNode.Content[i+1]
			if valueNode.Kind == yaml.ScalarNode {
				err := fn(nextYamlTrace(trace, keyNode.Value), valueNode)
				if err != nil {
					return err
				}
			} else {
				err := visitYamlRecursion(nextYamlTrace(trace, keyNode.Value), valueNode, fn)
				if err != nil {
					return err
				}
			}
		}
	} else if yamlNode.Kind == yaml.SequenceNode {
		for i, child := range yamlNode.Content {
			err := visitYamlRecursion(nextYamlTrace(trace, i), child, fn)
			if err != nil {
				return err
			}
		}
	} else {
		for _, child := range yamlNode.Content {
			err := visitYamlRecursion(trace, child, fn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func VisitYaml(yamlNode *yaml.Node, fn func(trace yamlTrace, yamlNode *yaml.Node) error) error {
	return visitYamlRecursion(yamlTrace{}, yamlNode, fn)
}

func nextYamlTrace(trace yamlTrace, next interface{}) yamlTrace {
	result := make(yamlTrace, len(trace))
	copy(result, trace)
	result = append(result, next)
	return result
}
