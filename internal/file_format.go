package internal

import (
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type FileFormat interface {
	ExtractLineComment(line string) (string, error)
	ReadValue(line string) (string, error)
	WriteValue(line string, value string) (string, error)
}

func GuessFileFormatFromExtension(file string) (FileFormat, error) {
	ext := path.Ext(file)
	switch ext {
	case ".yml", ".yaml":
		format := &YamlFileFormat{}
		return format, nil
	default:
		return nil, fmt.Errorf("unsupported file extension %s", ext)
	}
}

type YamlFileFormat struct{}

func (f YamlFileFormat) ExtractLineComment(line string) (string, error) {
	_, rest := separateLeadingWhitspaces(line)
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(rest), node); err != nil {
		return "", err
	}
	value := ""
	if err := visitYaml(node, func(node *yaml.Node) error {
		value = node.LineComment
		return nil
	}); err != nil {
		return "", err
	}
	return strings.TrimLeft(value, " #"), nil
}

func (f YamlFileFormat) ReadValue(line string) (string, error) {
	_, rest := separateLeadingWhitspaces(line)
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(rest), node); err != nil {
		return "", err
	}
	value := ""
	if err := visitYaml(node, func(node *yaml.Node) error {
		value = node.Value
		return nil
	}); err != nil {
		return "", err
	}
	return value, nil
}

func (f YamlFileFormat) WriteValue(line string, value string) (string, error) {
	lws, rest := separateLeadingWhitspaces(line)
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(rest), node); err != nil {
		return "", err
	}
	if err := visitYaml(node, func(node *yaml.Node) error {
		node.Value = value
		return nil
	}); err != nil {
		return "", err
	}
	output, err := yaml.Marshal(node)
	if err != nil {
		return line, err
	}
	return lws + strings.TrimSuffix(string(output), "\n"), nil
}

var _ FileFormat = (*YamlFileFormat)(nil)

func separateLeadingWhitspaces(str string) (string, string) {
	for i := 0; i < len(str); i++ {
		if str[i] != ' ' && str[i] != '\t' {
			return str[0:i], str[i:]
		}
	}
	return str, ""
}

func visitYaml(node *yaml.Node, fn func(node *yaml.Node) error) error {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, node := range node.Content {
			if err := visitYaml(node, fn); err != nil {
				return err
			}
		}
		return nil
	case yaml.SequenceNode:
		for _, node := range node.Content {
			if err := visitYaml(node, fn); err != nil {
				return err
			}
		}
		return nil
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			// keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			if err := visitYaml(valueNode, fn); err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		return fn(node)
	default:
		return nil
	}
}
