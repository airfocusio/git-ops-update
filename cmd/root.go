package cmd

import (
	"github.com/airfocusio/git-ops-update/internal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	rootCmdVerbose       bool
	rootCmdNoColor       bool
	rootCmdRegisterFlags = func(cmd *cobra.Command) {
		cmd.PersistentFlags().BoolVarP(&rootCmdVerbose, "verbose", "v", false, "")
		cmd.PersistentFlags().BoolVar(&rootCmdNoColor, "no-color", false, "no-color")
		findCmdRegisterFlags(cmd)
	}
	rootCmd = &cobra.Command{
		Use:           "git-ops-update",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if rootCmdVerbose {
				internal.SetLogVerbosity(true)
			}
			if rootCmdNoColor {
				color.NoColor = true
			}
		},
		RunE: findCmd.RunE,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(findCmd)
	rootCmdRegisterFlags(rootCmd)
}
