package internal

import (
	utiljson "encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type UpdateVersionsOptions struct {
	DryRun bool
}

func ApplyUpdates(dir string, config Config, opts UpdateVersionsOptions) error {
	changes, err := DetectUpdates(dir, config)
	if err != nil {
		return err
	}

	for _, c := range *changes {
		if !opts.DryRun {
			done, err := c.Action(dir, Changes{c})
			if err != nil {
				return err
			}
			if done {
				fmt.Printf("%s\n", c.Message())
			}
		} else {
			fmt.Printf("%s\n", c.Message())
		}
	}

	return nil
}

func DetectUpdates(dir string, config Config) (*Changes, error) {
	cacheFile := fileResolvePath(dir, ".git-ops-update.cache.yaml")
	cache, err := LoadCacheFromFile(cacheFile)
	if err != nil {
		fmt.Printf("unable to read cache: %v\n", err)
		cache = &Cache{}
	}

	files, err := fileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return nil, err
	}

	changes := Changes{}
	for _, file := range *files {
		fileRel, err := filepath.Rel(dir, file)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Scanning file %s\n", fileRel)

		fileDoc := &yaml.Node{}
		err = fileReadYaml(file, fileDoc)
		if err != nil {
			return nil, err
		}

		err = VisitYaml(fileDoc, func(trace yamlTrace, yamlNode *yaml.Node) error {
			lineComment := strings.TrimPrefix(yamlNode.LineComment, "#")
			if lineComment == "" {
				return nil
			}

			annotation, err := parseAnnotation(*yamlNode, lineComment, config)
			if err != nil {
				return err
			}
			if annotation == nil {
				return nil
			}

			var availableVersions []string
			cachedResource := cache.FindResource(annotation.RegistryName, annotation.ResourceName)
			if cachedResource == nil || cachedResource.Timestamp.Add(time.Duration((*annotation.Registry).GetInterval())).Before(time.Now()) {
				versions, err := (*annotation.Registry).FetchVersions(annotation.ResourceName)
				if err != nil {
					return err
				}
				availableVersions = *versions
				nextCache := cache.UpdateResource(CacheResource{
					RegistryName: annotation.RegistryName,
					ResourceName: annotation.ResourceName,
					Versions:     availableVersions,
					Timestamp:    time.Now(),
				})
				cache = &nextCache
				err = SaveCacheToFile(*cache, cacheFile)
				if err != nil {
					return err
				}
			} else {
				availableVersions = cachedResource.Versions
			}

			currentValue := yamlNode.Value
			currentVersion, err := (*annotation.Format).ExtractVersion(currentValue)
			if err != nil {
				return err
			}
			nextVersion, err := annotation.Policy.FindNext(*currentVersion, availableVersions)
			if err != nil {
				return err
			}

			if *currentVersion != *nextVersion {
				traceStr := ""
				for _, e := range trace {
					s, ok := e.(string)
					if ok {
						traceStr = traceStr + "." + s
					}
				}
				nextValue, err := (*annotation.Format).ReplaceVersion(currentValue, *nextVersion)
				if err != nil {
					return err
				}
				changes = append(changes, Change{
					RegistryName: annotation.RegistryName,
					ResourceName: annotation.ResourceName,
					OldVersion:   *currentVersion,
					NewVersion:   *nextVersion,
					File:         fileRel,
					Trace:        trace,
					OldValue:     currentValue,
					NewValue:     *nextValue,
					Action:       *annotation.Action,
				})
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	err = SaveCacheToFile(*cache, cacheFile)
	if err != nil {
		return nil, err
	}

	return &changes, nil
}

type annotation struct {
	RegistryName string `json:"registry"`
	Registry     *Registry
	ResourceName string `json:"resource"`
	PolicyName   string `json:"policy"`
	Policy       *Policy
	FormatName   string `json:"format"`
	Format       *Format
	ActionName   string `json:"action"`
	Action       *Action
}

func parseAnnotation(valueNode yaml.Node, annotationStr string, config Config) (*annotation, error) {
	regex := regexp.MustCompile(`git-ops-update\s*(\{.*?\})`)
	annotationStrMatch := regex.FindStringSubmatch(annotationStr)
	if annotationStrMatch == nil {
		return nil, nil
	}

	annotation := annotation{}
	err := utiljson.Unmarshal([]byte(annotationStrMatch[1]), &annotation)
	if err != nil || annotation.RegistryName == "" || annotation.ResourceName == "" || annotation.PolicyName == "" {
		return nil, nil
	}

	registry, ok := config.Registries[annotation.RegistryName]
	if !ok {
		return nil, fmt.Errorf("line %d column %d: annotation %s references unknown registry %s", valueNode.Line, valueNode.Column, annotationStr, annotation.RegistryName)
	}
	annotation.Registry = &registry

	policy, ok := config.Policies[annotation.PolicyName]
	if !ok {
		return nil, fmt.Errorf("line %d column %d: annotation %s references unknown policy %s", valueNode.Line, valueNode.Column, annotationStr, annotation.PolicyName)
	}
	annotation.Policy = &policy

	format, err := getFormat(annotation.FormatName)
	if err != nil {
		return nil, err
	}
	annotation.Format = format

	action, err := getAction(config.Git.Provider, annotation.ActionName)
	if err != nil {
		return nil, err
	}
	annotation.Action = action

	return &annotation, nil
}
