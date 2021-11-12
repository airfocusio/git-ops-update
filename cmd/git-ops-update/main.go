package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/choffmeister/git-ops-update/internal"
	"github.com/spf13/viper"
)

func main() {
	dry := flag.Bool("dry", false, "Dry run")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current directory: %v\n", err)
	}
	opts := internal.UpdateVersionsOptions{
		Dry: *dry,
	}

	viperInstance := viper.New()
	viperInstance.SetConfigName(".git-ops-update")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(dir)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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
