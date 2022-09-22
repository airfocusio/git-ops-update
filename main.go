package main

import (
	"os"

	"github.com/airfocusio/git-ops-update/cmd"
	"github.com/airfocusio/git-ops-update/internal"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	cmd.Version = cmd.FullVersion{Version: version, Commit: commit, Date: date, BuiltBy: builtBy}
	if err := cmd.Execute(); err != nil {
		internal.LogError("%v", err)
		os.Exit(1)
	}
}
