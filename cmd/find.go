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
			result := internal.DetectUpdates(dir, *config, cacheProvider)
			errorCount := 0

			if findCmdDry {
				for _, r := range result {
					if r.Error != nil {
						errorCount += 1
						internal.LogError("%v", r.Error)
					} else if r.Change != nil && r.Action != nil {
						internal.LogInfo("%s:%d the version can be updated from %s to %s", r.Change.File, r.Change.LineNum, r.Change.OldVersion, r.Change.NewVersion)
					}
				}
			} else {
				errored := internal.SliceFilter(result, func(r internal.UpdateVersionResult) bool {
					return r.Error != nil
				})
				for _, r := range errored {
					errorCount += 1
					internal.LogError("%v", r.Error)
				}

				unerrored := internal.SliceFilter(result, func(r internal.UpdateVersionResult) bool {
					return r.Error == nil && r.Action != nil && r.Change != nil
				})
				ungrouped := internal.SliceFilter(unerrored, func(r internal.UpdateVersionResult) bool {
					return r.Change.Group == ""
				})
				grouped := internal.SliceGroupBy(internal.SliceFilter(unerrored, func(u internal.UpdateVersionResult) bool {
					return u.Change.Group != ""
				}), func(u internal.UpdateVersionResult) string {
					return (*u.Action).Identifier() + ":" + u.Change.Group
				})

				type task struct {
					action    internal.Action
					changeSet internal.ChangeSet
				}

				tasks := []task{}
				tasks = append(tasks, internal.SliceMap(ungrouped, func(r internal.UpdateVersionResult) task {
					return task{
						action: *r.Action,
						changeSet: internal.ChangeSet{
							Changes: []internal.Change{*r.Change},
						},
					}
				})...)
				tasks = append(tasks, internal.MapMap(grouped, func(rs []internal.UpdateVersionResult, group string) task {
					return task{
						action: *rs[0].Action,
						changeSet: internal.ChangeSet{
							Group: group,
							Changes: internal.SliceMap(rs, func(r internal.UpdateVersionResult) internal.Change {
								return *r.Change
							}),
						},
					}
				})...)

				for _, t := range tasks {
					err := internal.ApplyUpdate(dir, *config, cacheProvider, t.action, t.changeSet)
					if err != nil {
						errorCount += 1
						internal.LogError("%v", err)
					} else {
						for _, c := range t.changeSet.Changes {
							internal.LogInfo("%s:%d the version was updated from %s to %s", c.File, c.LineNum, c.OldVersion, c.NewVersion)
						}
					}
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
