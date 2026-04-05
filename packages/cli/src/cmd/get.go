package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// ── get ──────────────────────────────────────────────────────────────────────

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show full content of a clip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		e, err := database.Get(id)
		if err != nil {
			return fmt.Errorf("clip %d not found", id)
		}

		pin := ""
		if e.Pinned {
			pin = " 📌 pinned"
		}
		fmt.Printf("Clip #%d%s  [%s]  source: %q\n", e.ID, pin, e.CreatedAt.Format("2006-01-02 15:04:05"), e.Source)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(e.Content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
