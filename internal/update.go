package internal

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type UpdateVersionsOptions struct {
	Dry bool
}

func UpdateVersions(dir string, config Config, opts UpdateVersionsOptions) error {
	cacheFile := fileResolvePath(dir, ".git-ops-update.cache.yaml")
	cache, err := LoadCacheFromFile(cacheFile)
	if err != nil {
		return err
	}

	files, err := fileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return err
	}

	changes := Changes{}
	for _, file := range *files {
		fileRel, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}
		log.Printf("Scanning file %s\n", fileRel)

		fileDoc := &yaml.Node{}
		err = fileReadYaml(file, fileDoc)
		if err != nil {
			return err
		}

		err = VisitAnnotations(fileDoc, "git-ops-update", func(trace yamlTrace, yamlNode *yaml.Node, annotation string) error {
			registryName, registry, resourceName, _, policy, format, action, err := parseAnnotation(*yamlNode, annotation, config)
			if err != nil {
				return err
			}

			var availableVersions []string
			cachedResource := cache.FindResource(*registryName, *resourceName)
			if cachedResource == nil || cachedResource.Timestamp.Add(time.Duration((*registry).GetInterval())).Before(time.Now()) {
				versions, err := (*registry).FetchVersions(*resourceName)
				if err != nil {
					return err
				}
				availableVersions = *versions
				nextCache := cache.UpdateResource(CacheResource{
					RegistryName: *registryName,
					ResourceName: *resourceName,
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
			currentVersion, err := (*format).ExtractVersion(currentValue)
			if err != nil {
				return err
			}
			nextVersion, err := policy.FindNext(*currentVersion, availableVersions)
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
				nextValue, err := (*format).ReplaceVersion(currentValue, *nextVersion)
				if err != nil {
					return err
				}
				changes = append(changes, Change{
					RegistryName: *registryName,
					ResourceName: *resourceName,
					OldVersion:   *currentVersion,
					NewVersion:   *nextVersion,
					File:         fileRel,
					Trace:        trace,
					OldValue:     currentValue,
					NewValue:     *nextValue,
					Action:       *action,
				})
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	for _, c := range changes {
		log.Printf("%s\n", c.Message())
		if !opts.Dry {
			err := c.Action(dir, Changes{c})
			if err != nil {
				return err
			}
			err = SaveCacheToFile(*cache, cacheFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseAnnotation(valueNode yaml.Node, annotation string, config Config) (*string, *Registry, *string, *string, *Policy, *Format, *Action, error) {
	regex := regexp.MustCompile("^(?P<registryName>[^:]+):(?P<resourceName>.+):(?P<policyName>[^:]+):(?P<formatName>[^:]+):(?P<actionName>[^:]+)$")
	match := regex.FindStringSubmatch(annotation)
	if match == nil {
		return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation %s is malformed", valueNode.Line, valueNode.Column, annotation)
	}

	registryName := match[1]
	registry, ok := config.Registries[registryName]
	if !ok {
		return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation %s references unknown registry %s", valueNode.Line, valueNode.Column, annotation, registryName)
	}

	resourceName := match[2]

	policyName := match[3]
	policy, ok := config.Policies[policyName]
	if !ok {
		return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation %s references unknown policy %s", valueNode.Line, valueNode.Column, annotation, policyName)
	}

	formatName := match[4]
	format, err := getFormat(formatName)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	actionName := match[5]
	action, err := getAction(config.Git.Provider, actionName)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	return &registryName, &registry, &resourceName, &policyName, &policy, format, action, nil
}
