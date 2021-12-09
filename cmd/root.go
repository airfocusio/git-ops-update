package cmd

import (
	"fmt"
	"io/ioutil"
	"runtime/debug"

	"github.com/airfocusio/git-ops-update/internal"
	"github.com/spf13/cobra"
)

type rootCmd struct {
	cmd       *cobra.Command
	directory string
	dry       bool
	verbose   bool
}

func newRootCmd(version FullVersion) *rootCmd {
	result := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "git-ops-update",
		Short:        "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := result.directory
			fileBytes, err := ioutil.ReadFile(internal.FileResolvePath(dir, ".git-ops-update.yaml"))
			if err != nil {
				return fmt.Errorf("unable to initialize: %v", err)
			}
			config, err := internal.LoadConfig(fileBytes)
			if err != nil {
				return fmt.Errorf("unable to load configuration: %v", err)
			}
			cacheFile := internal.FileResolvePath(dir, ".git-ops-update.cache.yaml")
			cacheProvider := internal.FileCacheProvider{File: cacheFile}
			err = internal.ApplyUpdates(dir, *config, cacheProvider, internal.UpdateVersionsOptions{
				Dry:     result.dry,
				Verbose: result.verbose,
			})
			if err != nil {
				return fmt.Errorf("unable to update versions: %v", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&result.directory, "dir", ".", "dir")
	cmd.Flags().BoolVar(&result.dry, "dry", false, "dry")
	cmd.Flags().BoolVar(&result.verbose, "verbose", false, "verbose")

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
