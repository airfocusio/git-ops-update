package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"

	"github.com/airfocusio/git-ops-update/internal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type rootCmd struct {
	cmd       *cobra.Command
	directory string
	dry       bool
	verbose   bool
	noColor   bool
}

func newRootCmd(version FullVersion) *rootCmd {
	result := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "git-ops-update",
		Short:        "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			internal.SetLogVerbosity(result.verbose)
			if result.noColor {
				color.NoColor = true
			}
			dir := result.directory
			fileBytes, err := ioutil.ReadFile(internal.FileResolvePath(dir, ".git-ops-update.yaml"))
			if err != nil {
				internal.LogError("Unable to initialize: %v", err)
				os.Exit(1)
			}
			config, err := internal.LoadConfig(fileBytes)
			if err != nil {
				internal.LogError("Unable to load configuration: %v", err)
				os.Exit(1)
			}
			cacheFile := internal.FileResolvePath(dir, ".git-ops-update.cache.yaml")
			cacheProvider := internal.FileCacheProvider{File: cacheFile}
			errs := internal.ApplyUpdates(dir, *config, cacheProvider, internal.UpdateVersionsOptions{
				Dry: result.dry,
			})
			if len(errs) > 0 {
				for _, err := range errs {
					internal.LogError("%v", err)
				}
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&result.directory, "dir", ".", "dir")
	cmd.Flags().BoolVar(&result.dry, "dry", false, "dry")
	cmd.Flags().BoolVar(&result.verbose, "verbose", false, "verbose")
	cmd.Flags().BoolVar(&result.noColor, "no-color", false, "no-color")
	result.cmd = cmd
	return result
}

func Execute(version FullVersion) error {
	rootCmd := newRootCmd(version)
	return rootCmd.cmd.Execute()
}

type FullVersion struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func (v FullVersion) ToString() string {
	result := v.Version
	if v.Commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, v.Commit)
	}
	if v.Date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, v.Date)
	}
	if v.BuiltBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, v.BuiltBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}
