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
	Dry bool
}

func ApplyUpdates(dir string, config Config, cacheProvider CacheProvider, opts UpdateVersionsOptions) []error {
	changes, errs := DetectUpdates(dir, config, cacheProvider)

	for _, c := range changes {
		if !opts.Dry {
			done, err := c.Action(dir, Changes{c})
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if done {
				LogInfo("%s:%s was updated from %s to %s", c.File, c.Trace.ToString(), c.OldValue, c.NewValue)
			} else {
				LogDebug("%s:%s could have been updated from %s to %s but as skipped", c.File, c.Trace.ToString(), c.OldValue, c.NewValue)
			}
		} else {
			LogInfo("%s:%s can be updated from %s to %s", c.File, c.Trace.ToString(), c.OldValue, c.NewValue)
		}
	}

	return errs
}

func DetectUpdates(dir string, config Config, cacheProvider CacheProvider) (Changes, []error) {
	cache, err := cacheProvider.Load()
	if err != nil {
		LogWarning("Unable to read cache: %v", err)
		cache = &Cache{}
	}

	files, err := FileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return nil, []error{err}
	}

	changes := Changes{}
	errors := []error{}
	for _, file := range *files {
		fileRel, err := filepath.Rel(dir, file)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		LogDebug("Scanning file %s", fileRel)

		fileDoc := &yaml.Node{}
		err = fileReadYaml(file, fileDoc)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", fileRel, err))
			continue
		}

		errs := VisitYaml(fileDoc, func(trace yamlTrace, yamlNode *yaml.Node) error {
			lineComment := strings.TrimPrefix(yamlNode.LineComment, "#")
			if lineComment == "" {
				return nil
			}

			annotation, err := parseAnnotation(*yamlNode, lineComment, config)
			if err != nil {
				return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
			}
			if annotation == nil {
				return nil
			}

			var availableVersions []string
			cachedResource := cache.FindResource(annotation.RegistryName, annotation.ResourceName)
			if cachedResource == nil || cachedResource.Timestamp.Add(time.Duration((*annotation.Registry).GetInterval())).Before(time.Now()) {
				versions, err := (*annotation.Registry).FetchVersions(annotation.ResourceName)
				if err != nil {
					return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
				}
				availableVersions = *versions
				nextCache := cache.UpdateResource(CacheResource{
					RegistryName: annotation.RegistryName,
					ResourceName: annotation.ResourceName,
					Versions:     availableVersions,
					Timestamp:    time.Now(),
				})
				cache = &nextCache
				err = cacheProvider.Save(*cache)
				if err != nil {
					return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
				}
			} else {
				availableVersions = cachedResource.Versions
			}

			currentValue := yamlNode.Value
			currentVersion, err := (*annotation.Format).ExtractVersion(currentValue)
			if err != nil {
				return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
			}
			nextVersion, err := annotation.Policy.FindNext(*currentVersion, availableVersions, annotation.Prefix, annotation.Suffix)
			if err != nil {
				return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
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
					return fmt.Errorf("%s:%s: %w", fileRel, trace.ToString(), err)
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
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}
	}

	return changes, errors
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
	Prefix       string `json:"prefix"`
	Suffix       string `json:"suffix"`
}

func parseAnnotation(valueNode yaml.Node, annotationStrFull string, config Config) (*annotation, error) {
	regex := regexp.MustCompile(`git-ops-update\s*(\{.*?\})`)
	annotationStrMatch := regex.FindStringSubmatch(annotationStrFull)
	if annotationStrMatch == nil {
		return nil, nil
	}
	annotationStr := annotationStrMatch[1]

	annotation := annotation{}
	err := utiljson.Unmarshal([]byte(annotationStr), &annotation)
	if err != nil {
		return nil, fmt.Errorf("annotation %s malformed: %w", annotationStr, err)
	}

	if annotation.RegistryName == "" {
		return nil, fmt.Errorf("annotation %s misses registry", annotationStr)
	}
	registry, ok := config.Registries[annotation.RegistryName]
	if !ok {
		return nil, fmt.Errorf("annotation %s references unknown registry %s", annotationStr, annotation.RegistryName)
	}
	annotation.Registry = &registry

	if annotation.ResourceName == "" {
		return nil, fmt.Errorf("annotation %s misses resource", annotationStr)
	}

	if annotation.PolicyName == "" {
		return nil, fmt.Errorf("annotation %s misses policy", annotationStr)
	}
	policy, ok := config.Policies[annotation.PolicyName]
	if !ok {
		return nil, fmt.Errorf("annotation %s references unknown policy %s", annotationStr, annotation.PolicyName)
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
