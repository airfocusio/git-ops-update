package main

import (
	"flag"
	"log"
	"os"

	"github.com/choffmeister/git-ops-update/internal"
)

func main() {
	dry := flag.Bool("dry", false, "Dry run")
	config := flag.String("config", ".git-ops-update.yaml", "Config file")
	cache := flag.String("cache", ".git-ops-update.cache.yaml", "Cache file")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current directory: %v\n", err)
	}
	opts := internal.UpdateVersionsOptions{
		Dry:        *dry,
		ConfigFile: *config,
		CacheFile:  *cache,
	}
	err = internal.UpdateVersions(dir, opts)
	if err != nil {
		log.Fatalf("unable to update versions: %v\n", err)
	}
}