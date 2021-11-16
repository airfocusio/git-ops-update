package internal

import (
	utiljson "encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"
)

type UpdateVersionsOptions struct {
	DryRun bool
}

func ApplyUpdates(dir string, config Config, opts UpdateVersionsOptions) error {
	var result error

	changes, err := DetectUpdates(dir, config)
	if err != nil {
		result = multierror.Append(err)
	}

	for _, c := range *changes {
		if !opts.DryRun {
			done, err := c.Action(dir, Changes{c})
			if err != nil {
				result = multierror.Append(err, result)
			}
			if done {
				fmt.Printf("%s\n", c.Message())
			}
		} else {
			fmt.Printf("%s\n", c.Message())
		}
	}

	return result
}

func DetectUpdates(dir string, config Config) (*Changes, error) {
	var result error

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
			result = multierror.Append(fmt.Errorf("failed to get relative path to file %s: %w", file, err), result)
		}
		fmt.Printf("Scanning file %s\n", fileRel)

		fileDoc := &yaml.Node{}
		err = fileReadYaml(file, fileDoc)
		if err != nil {
			result = multierror.Append(fmt.Errorf("failed to read file %s: %w", file, err), result)
		}

		err = VisitYaml(fileDoc, func(trace yamlTrace, yamlNode *yaml.Node) error {
			lineComment := strings.TrimPrefix(yamlNode.LineComment, "#")
			if lineComment == "" {
				return nil
			}

			annotation, err := parseAnnotation(*yamlNode, lineComment, config)
			if err != nil {
				return fmt.Errorf("failed to parse annotation in file %s: %w", file, err)
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
			nextVersion, err := annotation.Policy.FindNext(*currentVersion, availableVersions, annotation.Prefix, annotation.Suffix)
			if err != nil {
				return fmt.Errorf("failed to find next version in file %s: %w", file, err)
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
			result = multierror.Append(err, result)
		}
	}

	err = SaveCacheToFile(*cache, cacheFile)
	if err != nil {
		result = multierror.Append(err, result)
	}

	return &changes, result
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
		return nil, fmt.Errorf("annotation %s malformed: %v", annotationStr, err)
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
