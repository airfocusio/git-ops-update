package cmd

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/airfocusio/git-ops-update/internal"
	"github.com/spf13/cobra"
)

var (
	findCmdDirectory     string
	findCmdDry           bool
	findCmdIncludes      []string
	findCmdExcludes      []string
	findCmdRegisterFlags = func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&findCmdDirectory, "dir", ".", "dir")
		cmd.Flags().BoolVar(&findCmdDry, "dry", false, "dry")
		cmd.Flags().StringArrayVar(&findCmdIncludes, "includes", []string{}, "includes")
		cmd.Flags().StringArrayVar(&findCmdExcludes, "excludes", []string{}, "excludes")
	}
	findCmd = &cobra.Command{
		Version:       Version.Version,
		Use:           "find",
		Short:         "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := findCmdDirectory
			fileBytes, err := ioutil.ReadFile(internal.FileResolvePath(dir, ".git-ops-update.yaml"))
			if err != nil {
				return fmt.Errorf("unable to initialize: %w", err)
			}
			config, err := internal.LoadConfig(fileBytes)
			if err != nil {
				return fmt.Errorf("unable to load configuration: %w", err)
			}
			if len(findCmdIncludes) > 0 {
				includes := []regexp.Regexp{}
				for _, i := range findCmdIncludes {
					regex, err := regexp.Compile(i)
					if err != nil {
						return fmt.Errorf("unable to parse includes: %w", err)
					}
					includes = append(includes, *regex)
				}
				config.Files.Includes = includes
			}
			if len(findCmdExcludes) > 0 {
				excludes := []regexp.Regexp{}
				for _, e := range findCmdExcludes {
					regex, err := regexp.Compile(e)
					if err != nil {
						return fmt.Errorf("unable to parse excludes: %w", err)
					}
					excludes = append(excludes, *regex)
				}
				config.Files.Excludes = excludes
			}
			cacheFile := internal.FileResolvePath(dir, ".git-ops-update.cache.yaml")
			cacheProvider := internal.FileCacheProvider{File: cacheFile}
			result := internal.ApplyUpdates(dir, *config, cacheProvider, findCmdDry)

			errorCount := 0
			for _, r := range result {
				if r.Error != nil {
					errorCount += 1
					internal.LogError("%v", r.Error)
				} else if r.SkipMessage == "dry run" {
					internal.LogInfo("%s:%d from %s to %s", r.Change.File, r.Change.LineNum, r.Change.OldVersion, r.Change.NewVersion)
				} else if r.SkipMessage != "" {
					internal.LogDebug("%s:%d the version could have been updated from %s to %s but as skipped (%s)", r.Change.File, r.Change.LineNum, r.Change.OldVersion, r.Change.NewVersion, r.SkipMessage)
				} else {
					internal.LogInfo("%s:%d the version was updated from %s to %s", r.Change.File, r.Change.LineNum, r.Change.OldVersion, r.Change.NewVersion)
				}
			}

			if errorCount > 0 {
				return fmt.Errorf("there where %d errors", errorCount)
			}
			return nil
		},
	}
)

func init() {
	findCmdRegisterFlags(findCmd)
}
