package cmd

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/history"
	"github.com/solodov/org-roam-alfred-items/roam"
	"github.com/spf13/cobra"
)

var booksCmd = &cobra.Command{
	Use:   "books [--query query]",
	Short: "Output books alfred items matching the argument",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", roamCmdArgs.dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		row, err := db.Query(`
			SELECT nodes.properties
			FROM nodes
			INNER JOIN files ON nodes.file = files.file
			WHERE nodes.level == 2 AND files.file LIKE '%/books.org%'`)
		if err != nil {
			log.Fatal(err)
		}
		items := []alfred.Item{
			alfred.Item{
				Title:        booksCmdArgs.query,
				Subtitle:     "search goodreads for " + booksCmdArgs.query,
				Arg:          "https://www.goodreads.com/search?q=" + booksCmdArgs.query,
				Autocomplete: booksCmdArgs.query,
				Variables: alfred.Variables{
					Profile: "home",
					Query:   booksCmdArgs.query,
				},
				Save: true,
			},
		}
		items = append(items, history.FindMatchingItems(rootCmdArgs.trigger, booksCmdArgs.query)...)
		for row.Next() {
			var props roam.Props
			if err := row.Scan(&props); err != nil {
				log.Fatal(err)
			}
			if data, err := props.ItemLinkData(); err != nil {
				continue
			} else if strings.Contains(strings.ToLower(data.Title), strings.ToLower(booksCmdArgs.query)) {
				items = append(
					items,
					alfred.Item{
						Title:        data.Title,
						Subtitle:     data.Url,
						Arg:          data.Url,
						Autocomplete: data.Title,
						Variables: alfred.Variables{
							Profile: "home",
							Query:   booksCmdArgs.query,
						},
					},
				)
			}
		}
		for i := range items {
			if items[i].Save {
				res, _ := json.Marshal(items[i])
				items[i].Variables.HistItem = string(res)
			}
		}
		printJson(alfred.Result{Items: items})
	},
}

var booksCmdArgs struct {
	query string
}

func init() {
	roamCmd.AddCommand(booksCmd)
	booksCmd.Flags().StringVar(&booksCmdArgs.query, "query", "", "Alfred input query")
}
