package internal

import (
	"fmt"
	"log"
	"strings"
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
		cache = &Cache{}
	}

	files, err := fileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return err
	}

	for _, file := range *files {
		log.Printf("Scanning file %s\n", file)
		fileDoc := &yaml.Node{}
		err := fileReadYaml(file, fileDoc)
		if err != nil {
			return err
		}

		err = VisitAnnotations(fileDoc, "git-ops-update", func(keyNode *yaml.Node, valueNode *yaml.Node, trace []string, annotation string) error {
			registryName, registry, resourceName, _, policy, format, err := parseAnnotation(*valueNode, annotation, config)
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

			currentValue := valueNode.Value
			currentVersion, err := (*format).ExtractVersion(currentValue)
			if err != nil {
				return err
			}
			nextVersion, err := policy.FindNext(*currentVersion, availableVersions)
			if err != nil {
				return err
			}

			if *currentVersion != *nextVersion {
				log.Printf("Update for %s/%s from %s to %s\n", *registryName, *resourceName, *currentVersion, *nextVersion)
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
			fileWriteYaml(file, fileDoc)
		}
	}

	return nil
}

func parseAnnotation(valueNode yaml.Node, annotation string, config Config) (*string, *Registry, *string, *string, *Policy, *Format, error) {
	segments := strings.Split(annotation, ":")

	registryName := segments[0]
	if len(segments) < 2 {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation is missing the resource", valueNode.Line, valueNode.Column)
	}
	registry, ok := config.Registries[registryName]
	if !ok {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation references unknown registry %s", valueNode.Line, valueNode.Column, registryName)
	}

	resourceName := segments[1]
	if len(segments) < 3 {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation is missing the policy", valueNode.Line, valueNode.Column)
	}

	policyName := segments[2]
	policy, ok := config.Policies[policyName]
	if !ok {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("line %d column %d: annotation references unknown policy %s", valueNode.Line, valueNode.Column, policyName)
	}

	formatName := "plain"
	if len(segments) >= 4 {
		formatName = segments[3]
	}
	format, err := GetFormat(formatName)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	return &registryName, &registry, &resourceName, &policyName, &policy, format, nil
}
