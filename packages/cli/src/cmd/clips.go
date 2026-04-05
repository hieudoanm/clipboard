package cmd

import (
	"fmt"
	"strconv"

	"github.com/hieudoanm/clipboard/src/db"
	"github.com/spf13/cobra"
)

// ── pin ───────────────────────────────────────────────────────────────────────

var pinCmd = &cobra.Command{
	Use:   "pin <id>",
	Short: "Pin a clip (keep it at the top and safe from clear)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return togglePin(args[0], true)
	},
}

var unpinCmd = &cobra.Command{
	Use:   "unpin <id>",
	Short: "Unpin a clip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return togglePin(args[0], false)
	},
}

func togglePin(idStr string, pin bool) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id: %s", idStr)
	}
	database, err := openDB()
	if err != nil {
		return err
	}
	defer database.Close()

	if err := database.Pin(id, pin); err != nil {
		return err
	}
	action := "Pinned"
	if !pin {
		action = "Unpinned"
	}
	fmt.Printf("✓ %s clip #%d\n", action, id)
	return nil
}

// ── clear ─────────────────────────────────────────────────────────────────────

var clearAll bool

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear clipboard history (keeps pinned clips by default)",
	Long: `Remove all clips from history. By default, pinned clips are preserved.
Use --all to also remove pinned clips.

Examples:
  clips clear           # remove all unpinned clips
  clips clear --all     # remove everything`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		keepPinned := !clearAll
		n, err := database.Clear(keepPinned)
		if err != nil {
			return err
		}

		if keepPinned {
			fmt.Printf("✓ Cleared %d unpinned clip(s). Pinned clips preserved.\n", n)
		} else {
			fmt.Printf("✓ Cleared all %d clip(s).\n", n)
		}
		return nil
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func openDB() (*db.DB, error) {
	path := dbPath
	if path == "" {
		path = db.DefaultPath()
	}
	return db.Open(path)
}

func init() {
	clearCmd.Flags().BoolVar(&clearAll, "all", false, "also clear pinned clips")

	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
	rootCmd.AddCommand(clearCmd)
}
