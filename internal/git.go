package internal

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Change struct {
	RegistryName string
	ResourceName string
	OldVersion   string
	NewVersion   string
	File         string
	Trace        []interface{}
	OldValue     string
	NewValue     string
}

type Git struct {
	Provider GitProvider
}

type GitProvider interface {
	Apply(dir string, changes []Change) error
	Request(dir string, changes []Change) error
}

type LocalGitProvider struct{}

type GitHubGitProvider struct {
	Owner       string
	Repo        string
	AccessToken string
}

func (p LocalGitProvider) Apply(dir string, changes []Change) error {
	for _, change := range changes {
		file := filepath.Join(dir, change.File)
		doc := yaml.Node{}
		err := fileReadYaml(file, &doc)
		if err != nil {
			return err
		}

		VisitYaml(&doc, func(trace yamlTrace, node *yaml.Node) error {
			if trace.Equal(change.Trace) {
				node.Value = change.NewValue
			}
			return nil
		})

		err = fileWriteYaml(file, &doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p LocalGitProvider) Request(dir string, changes []Change) error {
	return fmt.Errorf("local git provider does not support request")
}

func (p GitHubGitProvider) Apply(dir string, changes []Change) error {
	// TODO implement making a commit and pushing it to github repository default branch
	return fmt.Errorf("gitHub git provider apply is not yet implemented")
}

func (p GitHubGitProvider) Request(dir string, changes []Change) error {
	// TODO implement making a commit to a new branch and creating a pull request on github repository against default branch
	return fmt.Errorf("gitHub git provider apply is not yet implemented")
}
