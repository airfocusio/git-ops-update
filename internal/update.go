package internal

import (
	utiljson "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type UpdateVersionResult struct {
	Error  error
	Change *Change

	SkipMessage string
}

func ApplyUpdates(dir string, config Config, cacheProvider CacheProvider, dry bool) []UpdateVersionResult {
	result := DetectUpdates(dir, config, cacheProvider)

	for i := range result {
		result := &result[i]
		if result.Error == nil && result.Change != nil {
			if !dry {
				changes := Changes{*result.Change}
				if result.Change.Action == nil {
					result.SkipMessage = "marked as disabled"
				} else if (*result.Change.Action).AlreadyApplied(dir, changes) {
					result.SkipMessage = "already applied"
				} else {
					err := (*result.Change.Action).Apply(dir, changes)
					if err != nil {
						result.Error = fmt.Errorf("%s:%d: %w", result.Change.File, result.Change.LineNum, err)
					}
				}
			} else {
				result.SkipMessage = "dry run"
			}
		}
	}

	return result
}

func DetectUpdates(dir string, config Config, cacheProvider CacheProvider) []UpdateVersionResult {
	cache, err := cacheProvider.Load()
	if err != nil {
		LogWarning("Unable to read cache: %v", err)
		cache = &Cache{}
	}

	files, err := FileList(dir, config.Files.Includes, config.Files.Excludes)
	if err != nil {
		return []UpdateVersionResult{{Error: err}}
	}

	result := []UpdateVersionResult{}
	for _, file := range files {
		fileRel, err := filepath.Rel(dir, file)
		if err != nil {
			result = append(result, UpdateVersionResult{Error: err})
			continue
		}
		LogDebug("Scanning file %s", fileRel)

		bytes, err := os.ReadFile(file)
		if err != nil {
			result = append(result, UpdateVersionResult{Error: fmt.Errorf("%s: %w", fileRel, err)})
			continue
		}
		lines := strings.Split(string(bytes), "\n")

		errs := []error{}
		fileFormat, err := GuessFileFormatFromExtension(file)
		if err != nil {
			result = append(result, UpdateVersionResult{Error: fmt.Errorf("%s: %w", fileRel, err)})
			continue
		}
		fileAnnotations, err := fileFormat.ExtractAnnotations(lines)
		if err != nil {
			result = append(result, UpdateVersionResult{Error: fmt.Errorf("%s: %w", fileRel, err)})
			continue
		}

		for _, fileAnnotation := range fileAnnotations {
			annotation, err := parseAnnotation(fileAnnotation.AnnotationRaw, config)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
				continue
			}
			if annotation == nil {
				continue
			}

			var availableVersions []string
			cachedResource := cache.FindResource(annotation.RegistryName, annotation.ResourceName)
			if cachedResource == nil || cachedResource.Timestamp.Add(time.Duration((*annotation.Registry).GetInterval())).Before(time.Now()) {
				LogDebug("Fetching new versions for %s/%s ", annotation.RegistryName, annotation.ResourceName)
				versions, err := (*annotation.Registry).FetchVersions(annotation.ResourceName)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
					continue
				}
				availableVersions = versions
				nextCache := cache.UpdateResource(CacheResource{
					RegistryName: annotation.RegistryName,
					ResourceName: annotation.ResourceName,
					Versions:     availableVersions,
					Timestamp:    time.Now(),
				})
				cache = &nextCache
				err = cacheProvider.Save(*cache)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
					continue
				}
			} else {
				LogDebug("Using cached versions for %s/%s ", annotation.RegistryName, annotation.ResourceName)
				availableVersions = cachedResource.Versions
			}

			currentValue, err := fileFormat.ReadValue(lines, fileAnnotation.LineNum)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
				continue
			}
			currentVersion, err := (*annotation.Format).ExtractVersion(currentValue)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
				continue
			}
			nextVersion, err := annotation.Policy.FindNext(*currentVersion, availableVersions, annotation.Prefix, annotation.Suffix, annotation.Filter)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
				continue
			}

			if *currentVersion != *nextVersion {
				nextValue, err := (*annotation.Format).ReplaceVersion(currentValue, *nextVersion)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
					continue
				}
				metadata, err := (*annotation.Registry).RetrieveMetadata(annotation.ResourceName, *nextVersion)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s:%d: %w", fileRel, fileAnnotation.LineNum, err))
					continue
				}
				change := Change{
					RegistryName: annotation.RegistryName,
					ResourceName: annotation.ResourceName,
					Metadata:     metadata,
					OldVersion:   *currentVersion,
					NewVersion:   *nextVersion,
					File:         fileRel,
					FileFormat:   fileFormat,
					LineNum:      fileAnnotation.LineNum,
					OldValue:     currentValue,
					NewValue:     *nextValue,
					Action:       annotation.Action,
				}
				result = append(result, UpdateVersionResult{Change: &change})
			}
		}
		for _, err := range errs {
			result = append(result, UpdateVersionResult{Error: err})
		}
	}

	return result
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
	Prefix       string                 `json:"prefix"`
	Suffix       string                 `json:"suffix"`
	Filter       map[string]interface{} `json:"filter"`
	Exec         []string               `json:"exec"`
}

func parseAnnotation(annotationStrFull string, config Config) (*annotation, error) {
	regex := regexp.MustCompile(`git-ops-update\s*(\{.*)`)
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

	action, err := getAction(config.Git.Provider, annotation.ActionName, annotation.Exec)
	if err != nil {
		return nil, err
	}
	annotation.Action = action

	return &annotation, nil
}
