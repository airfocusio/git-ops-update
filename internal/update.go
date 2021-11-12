package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// UpdateVersionsOptions ...
type UpdateVersionsOptions struct {
	Dry    bool
	Config string
}

// UpdateVersions ...
func UpdateVersions(dir string, opts UpdateVersionsOptions) error {
	configFile := opts.Config
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(dir, configFile)
	}
	configRaw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	config, registries, policies, err := LoadGitOpsUpdaterConfig(configRaw)
	if err != nil {
		return err
	}

	files, err := fileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return err
	}

	for _, file := range *files {
		log.Printf("Updating file %s\n", file)

		fileBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		fileDoc := yaml.Node{}
		err = yaml.Unmarshal(fileBytes, &fileDoc)
		if err != nil {
			return err
		}

		err = VisitAnnotations(&fileDoc, "git-ops-update", func(keyNode *yaml.Node, valueNode *yaml.Node, parentNodes []*yaml.Node, annotation string) error {
			segments := strings.Split(annotation, ":")

			registryName := segments[0]
			if len(segments) < 2 {
				return fmt.Errorf("line %d column %d: annotation is missing the resource", valueNode.Line, valueNode.Column)
			}
			registry, ok := (*registries)[registryName]
			if !ok {
				return fmt.Errorf("line %d column %d: annotation references unknown registry %s", valueNode.Line, valueNode.Column, registryName)
			}

			resourceName := segments[1]
			if len(segments) < 3 {
				return fmt.Errorf("line %d column %d: annotation is missing the policy", valueNode.Line, valueNode.Column)
			}

			policyName := segments[2]
			policy, ok := (*policies)[policyName]
			if !ok {
				return fmt.Errorf("line %d column %d: annotation references unknown policy %s", valueNode.Line, valueNode.Column, policyName)
			}

			formatName := "plain"
			if len(segments) >= 4 {
				formatName = segments[3]
			}
			format, err := GetFormat(formatName)
			if err != nil {
				return err
			}

			availableVersions, err := registry.FetchVersions(resourceName)
			if err != nil {
				return err
			}
			currentValue := valueNode.Value
			currentVersion, err := (*format).ExtractVersion(currentValue)
			if err != nil {
				return err
			}
			nextVersion, err := policy.FindNext(*currentVersion, *availableVersions)
			if err != nil {
				return err
			}

			if *currentVersion != *nextVersion {
				log.Printf("%s/%s: %s -> %s\n", registryName, resourceName, *currentVersion, *nextVersion)
				nextValue, err := (*format).ReplaceVersion(currentValue, *nextVersion)
				if err != nil {
					return err
				}
				valueNode.Value = *nextValue
			}

			return nil
		})
		if err != nil {
			return err
		}

		if !opts.Dry {
			fileBytesOut, err := yaml.Marshal(&fileDoc)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(file, fileBytesOut, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
