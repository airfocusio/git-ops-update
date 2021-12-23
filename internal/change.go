package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Change struct {
	RegistryName string
	ResourceName string
	OldVersion   string
	NewVersion   string
	File         string
	Trace        yamlTrace
	OldValue     string
	NewValue     string
	Action       *Action
}

type Changes []Change

const gitHubMaxPullRequestTitleLength = 256

func (c Change) Identifier() string {
	return c.File + "#" + c.Trace.ToString() + "#" + c.NewValue
}

func (c Change) Hash() []byte {
	identifier := c.Identifier()
	hash := sha256.Sum256([]byte(identifier))
	return hash[:]
}

func (cs Changes) Hash() []byte {
	temp := []byte{}
	for _, c := range cs {
		temp = append(temp, c.Hash()...)
	}
	hash := sha256.Sum256(temp)
	return hash[:]
}

func (cs Changes) Branch(prefix string) string {
	return prefix + "/" + hex.EncodeToString(cs.Hash()[:8])
}

func (c Change) Message() string {
	return fmt.Sprintf("Update %s:%s from %s to %s", c.File, c.Trace.ToString(), c.OldValue, c.NewValue)
}

func (cs Changes) Title() string {
	updates := []string{}
	for _, change := range cs {
		updates = append(updates, fmt.Sprintf("%s:%s", change.ResourceName, change.NewVersion))
	}
	result := fmt.Sprintf("Update %s", strings.Join(updates, ", "))
	if len(result) > gitHubMaxPullRequestTitleLength {
		return result[0:gitHubMaxPullRequestTitleLength]
	}
	return result
}

func (cs Changes) Message() string {
	lines := []string{}
	if len(cs) > 1 {
		lines = append(lines, "Update", "")
	}
	for _, c := range cs {
		lines = append(lines, c.Message())
	}
	return strings.Join(lines, "\n")
}

func (c Change) Push(dir string, fileHooks ...func(file string) error) error {
	file := filepath.Join(dir, c.File)
	doc := yaml.Node{}
	err := fileReadYaml(file, &doc)
	if err != nil {
		return err
	}

	VisitYaml(&doc, func(trace yamlTrace, node *yaml.Node) error {
		if trace.Equal(c.Trace) {
			node.Value = c.NewValue
		}
		return nil
	})

	err = fileWriteYaml(file, &doc)
	if err != nil {
		return err
	}
	for _, hook := range fileHooks {
		err := hook(c.File)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs Changes) Push(dir string, fileHooks ...func(file string) error) error {
	for _, c := range cs {
		err := c.Push(dir, fileHooks...)
		if err != nil {
			return err
		}
	}
	return nil
}
