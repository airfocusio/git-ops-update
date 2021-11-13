package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/choffmeister/git-ops-update/internal"
	"github.com/spf13/viper"
)

func main() {
	dryFlag := flag.Bool("dry", false, "Dry")
	dirFlag := flag.String("dir", ".", "Directory")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current directory: %v\n", err)
	}
	if !filepath.IsAbs(*dirFlag) {
		dir = filepath.Join(dir, *dirFlag)
	} else {
		dir = *dirFlag
	}

	opts := internal.UpdateVersionsOptions{
		Dry: *dryFlag,
	}

	viperInstance := viper.New()
	viperInstance.SetConfigName(".git-ops-update")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(dir)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()
	err = viperInstance.ReadInConfig()
	if err != nil {
		log.Fatalf("unable to load configuration: %v\n", err)
	}

	config, err := internal.LoadConfig(*viperInstance)
	if err != nil {
		log.Fatalf("unable to load configuration: %v\n", err)
	}
	err = internal.UpdateVersions(dir, *config, opts)
	if err != nil {
		log.Fatalf("unable to update versions: %v\n", err)
	}
}
