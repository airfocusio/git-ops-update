package internal

import (
	"bytes"
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

func (t yamlTrace) Next(next interface{}) yamlTrace {
	result := make(yamlTrace, len(t))
	copy(result, t)
	result = append(result, next)
	return result
}

func (t yamlTrace) ToString() string {
	traceStr := ""
	for _, e := range t {
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
				traceStr = traceStr + "." + "???"
			}
		}
	}
	return traceStr
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

func visitYamlRecursion(trace yamlTrace, yamlNode *yaml.Node, fn func(trace yamlTrace, yamlNode *yaml.Node) error) error {
	if yamlNode.Kind == yaml.MappingNode {
		for i := 0; i < len(yamlNode.Content); i += 2 {
			keyNode := yamlNode.Content[i]
			valueNode := yamlNode.Content[i+1]
			err := visitYamlRecursion(trace.Next(keyNode.Value), valueNode, fn)
			if err != nil {
				return err
			}
		}
	} else if yamlNode.Kind == yaml.SequenceNode {
		for i, child := range yamlNode.Content {
			err := visitYamlRecursion(trace.Next(i), child, fn)
			if err != nil {
				return err
			}
		}
	} else if yamlNode.Kind == yaml.ScalarNode {
		err := fn(trace, yamlNode)
		if err != nil {
			return err
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
