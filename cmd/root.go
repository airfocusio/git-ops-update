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
	cmd          *cobra.Command
	directory    string
	dry          bool
	verbose      bool
	ignoreErrors bool
	noColor      bool
}

func newRootCmd(version FullVersion) *rootCmd {
	cmdCfg := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "git-ops-update",
		Short:        "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			internal.SetLogVerbosity(cmdCfg.verbose)
			if cmdCfg.noColor {
				color.NoColor = true
			}
			dir := cmdCfg.directory
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
			result := internal.ApplyUpdates(dir, *config, cacheProvider, cmdCfg.dry)

			errorCount := 0
			for _, r := range result {
				if r.Error != nil {
					errorCount += 1
					internal.LogError("%v", r.Error)
				} else if r.SkipMessage == "dry run" {
					internal.LogInfo("%s:%d from %s to %s", r.Change.File, r.Change.Line, r.Change.OldVersion, r.Change.NewVersion)
				} else if r.SkipMessage != "" {
					internal.LogDebug("%s:%d the version could have been updated from %s to %s but as skipped (%s)", r.Change.File, r.Change.Line, r.Change.OldVersion, r.Change.NewVersion, r.SkipMessage)
				} else {
					internal.LogInfo("%s:%d the version was updated from %s to %s", r.Change.File, r.Change.Line, r.Change.OldVersion, r.Change.NewVersion)
				}
			}

			if errorCount > 0 && !cmdCfg.ignoreErrors {
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cmdCfg.directory, "dir", ".", "dir")
	cmd.Flags().BoolVar(&cmdCfg.dry, "dry", false, "dry")
	cmd.Flags().BoolVar(&cmdCfg.verbose, "verbose", false, "verbose")
	cmd.Flags().BoolVar(&cmdCfg.ignoreErrors, "ignore-errors", false, "ignore-errors")
	cmd.Flags().BoolVar(&cmdCfg.noColor, "no-color", false, "no-color")
	cmdCfg.cmd = cmd
	return cmdCfg
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
