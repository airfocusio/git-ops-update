package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/choffmeister/git-ops-update/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	directory string
	dryRun    bool
	rootCmd   = &cobra.Command{
		Version:      "<version>",
		Use:          "git-ops-update",
		Short:        "An updater for docker images and helm charts in your infrastructure-as-code repository",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, viperInstance, err := initConfig()
			if err != nil {
				return fmt.Errorf("unable to initialize: %v", err)
			}
			config, err := internal.LoadConfig(*viperInstance)
			if err != nil {
				return fmt.Errorf("unable to load configuration: %v", err)
			}
			err = internal.UpdateVersions(*dir, *config, internal.UpdateVersionsOptions{
				DryRun: dryRun,
			})
			if err != nil {
				return fmt.Errorf("unable to update versions: %v", err)
			}
			return nil
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&directory, "dir", ".", "dir")
	rootCmd.Flags().BoolVar(&dryRun, "dry", false, "dry")
}

func initConfig() (*string, *viper.Viper, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	if !filepath.IsAbs(directory) {
		dir = filepath.Join(dir, directory)
	} else {
		dir = directory
	}

	viperInstance := viper.New()
	viperInstance.SetConfigName(".git-ops-update")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(dir)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()

	err = viperInstance.ReadInConfig()
	if err != nil {
		return nil, nil, err
	}

	return &dir, viperInstance, nil
}
