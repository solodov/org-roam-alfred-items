package cmd

import (
	"encoding/json"
	"log"
	"time"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/history"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Alfred history commands",
}

var addCmd = &cobra.Command{
	Use:   "add --trigger trigger --item item",
	Short: "Add selected item to history",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// For recording history the structure doesn't matter, as long as it's valid JSON.
		if err := json.Unmarshal([]byte(addCmdArgs.item), &alfred.Item{}); err != nil {
			log.Fatalf("invalid item json: %v", err)
		}
		db, err := history.Open()
		if err != nil {
			log.Fatalf("failed to open history database: %v", err)
		}
		defer db.Close()
		if _, err := db.Exec(
			"INSERT INTO items (ts, trigger, query, item) VALUES (?, ?, ?, ?)",
			time.Now().UnixMilli(), rootCmdArgs.trigger, addCmdArgs.query, addCmdArgs.item,
		); err != nil {
			log.Fatal(err)
		}
	},
}

var addCmdArgs struct {
	item  string
	query string
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&addCmdArgs.item, "item", "i", "", "JSON string of the alfred item to add to history")
	addCmd.Flags().StringVar(&addCmdArgs.query, "query", "", "Alfred input query")
}
