package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "clips",
	Short: "A fast, local clipboard manager",
	Long: `clips — a terminal clipboard manager backed by SQLite.

Store, search, pin, and manage clipboard history right from your terminal.

Examples:
  clips add "hello world"          # add a clip
  clips list                       # list recent clips
  clips list --search "hello"      # search clips
  clips get 1                      # show clip by ID
  clips copy 1                     # copy clip to clipboard
  clips pin 1                      # pin a clip
  clips delete 1                   # delete a clip
  clips clear                      # clear all unpinned clips
  clips stats                      # show statistics`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database (default: ~/.clipboard-manager/clips.db)")
}
