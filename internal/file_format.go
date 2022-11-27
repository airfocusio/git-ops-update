package internal

import (
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type FileFormat interface {
	ExtractAnnotations(all []string) ([]FileFormatAnnotation, error)
	ReadValue(lines []string, lineNum int) (string, error)
	WriteValue(lines []string, lineNum int, value string) error
}

type FileFormatAnnotation struct {
	LineNum       int
	AnnotationRaw string
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

var _ FileFormat = (*YamlFileFormat)(nil)

type YamlFileFormat struct{}

func (f YamlFileFormat) ExtractAnnotations(lines []string) ([]FileFormatAnnotation, error) {
	documentsLines := [][]string{{}}
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "---") {
			documentsLines = append(documentsLines, []string{})
		} else {
			documentsLines[len(documentsLines)-1] = append(documentsLines[len(documentsLines)-1], lines[i])
		}
	}

	result := []FileFormatAnnotation{}
	firstDocumentLine := 0
	for _, documentLines := range documentsLines {
		documentNode := &yaml.Node{}
		if err := yaml.Unmarshal([]byte(strings.Join(documentLines, "\n")), documentNode); err != nil {
			return result, err
		}

		if err := visitYaml(documentNode, func(node *yaml.Node) error {
			if node.Tag == "!!str" && (node.Style == 0 || node.Style == yaml.SingleQuotedStyle || node.Style == yaml.DoubleQuotedStyle) && node.LineComment != "" {
				result = append(result, FileFormatAnnotation{
					LineNum:       firstDocumentLine + node.Line,
					AnnotationRaw: strings.TrimLeft(node.LineComment, "# "),
				})
			}
			return nil
		}); err != nil {
			return result, err
		}
		firstDocumentLine = firstDocumentLine + len(documentLines) + 1
	}
	return result, nil
}

func (f YamlFileFormat) ReadValue(lines []string, lineNum int) (string, error) {
	line := lines[lineNum-1]
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

func (f YamlFileFormat) WriteValue(lines []string, lineNum int, value string) error {
	line := lines[lineNum-1]
	lws, rest := separateLeadingWhitspaces(line)
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(rest), node); err != nil {
		return err
	}
	if err := visitYaml(node, func(node *yaml.Node) error {
		node.Value = value
		return nil
	}); err != nil {
		return err
	}
	output, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	lines[lineNum-1] = lws + strings.TrimSuffix(string(output), "\n")
	return nil
}

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
