package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/choffmeister/git-ops-update/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rootCmd struct {
	cmd       *cobra.Command
	directory string
	dryRun    bool
}

func newRootCmd(version FullVersion) *rootCmd {
	result := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "git-ops-update",
		Short:        "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, viperInstance, err := initConfig(result)
			if err != nil {
				return fmt.Errorf("unable to initialize: %v", err)
			}
			config, err := internal.LoadConfig(*viperInstance)
			if err != nil {
				return fmt.Errorf("unable to load configuration: %v", err)
			}
			err = internal.ApplyUpdates(*dir, *config, internal.UpdateVersionsOptions{
				DryRun: result.dryRun,
			})
			if err != nil {
				return fmt.Errorf("unable to update versions: %v", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&result.directory, "dir", ".", "dir")
	cmd.Flags().BoolVar(&result.dryRun, "dry", false, "dry")

	result.cmd = cmd
	return result
}

func initConfig(rootCmd *rootCmd) (*string, *viper.Viper, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	if !filepath.IsAbs(rootCmd.directory) {
		dir = filepath.Join(dir, rootCmd.directory)
	} else {
		dir = rootCmd.directory
	}

	viperInstance := viper.New()
	viperInstance.SetConfigName(".git-ops-update")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(dir)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.SetEnvPrefix("GIT_OPS_UPDATE")
	viperInstance.AutomaticEnv()

	err = viperInstance.ReadInConfig()
	if err != nil {
		return nil, nil, err
	}

	return &dir, viperInstance, nil
}

func Execute(version FullVersion) error {
	rootCmd := newRootCmd(version)
	initConfig(rootCmd)
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
