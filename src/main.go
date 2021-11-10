package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// UpdateVersionsOptions ...
type UpdateVersionsOptions struct {
	Dry    bool
	Config string
}

// Run ...
func Run(dir string, opts UpdateVersionsOptions) error {
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

	files, err := fileList(dir, config.Files.Includes, append(config.Files.Excludes, configFile))
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

		fileBytesOut, err := yaml.Marshal(&fileDoc)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(file, fileBytesOut, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	dry := flag.Bool("dry", false, "Dry run")
	config := flag.String("config", ".git-ops-update.yaml", "Config file")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current directory: %v\n", err)
	}
	opts := UpdateVersionsOptions{
		Dry:    *dry,
		Config: *config,
	}
	err = Run(dir, opts)
	if err != nil {
		log.Fatalf("unable to update versions: %v\n", err)
	}
}

func fileList(dir string, includes []string, excludes []string) (*[]string, error) {
	temp := []string{}
	for _, exclude := range excludes {
		fs, err := fileGlob(dir, exclude)
		if err != nil {
			return nil, err
		}
		temp = append(temp, *fs...)
	}
	files := []string{}
	for _, include := range includes {
		fs, err := fileGlob(dir, include)
		if err != nil {
			return nil, err
		}
		for _, f := range *fs {
			excluded := false
			for _, f2 := range temp {
				if f == f2 {
					excluded = true
					break
				}
			}
			if !excluded {
				files = append(files, f)
			}
		}
	}
	return &files, nil
}

func fileGlob(dir string, pattern string) (*[]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		_, file := filepath.Split(path)
		matched, err := filepath.Match(pattern, file)
		if err != nil {
			return err
		}
		if matched {
			files = append(files, path)
		}
		return nil
	})
	return &files, err
}
