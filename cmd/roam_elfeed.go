/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/roam"
	"github.com/spf13/cobra"
)

var elfeedCmd = &cobra.Command{
	Use:   "elfeed",
	Short: "Alfred elfeed commands",
}

var elfeedItemsCmd = &cobra.Command{
	Use:                   "items",
	DisableFlagsInUseLine: true,
	Short:                 "Output elfeed alfred items",
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printJson(alfred.Result{Items: readElfeedItems()})
	},
}

var elfeedResolveCmd = &cobra.Command{
	Use:                   "resolve title",
	DisableFlagsInUseLine: true,
	Short:                 "Resolve elfeed title to its link",
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, item := range readElfeedItems() {
			if item.Title == args[0] {
				fmt.Print(item.Arg)
				return
			}
		}
		log.Fatalf("not found")
	},
}

func readElfeedItems() (items []alfred.Item) {
	db, err := sql.Open("sqlite3", roamCmdArgs.dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`
		SELECT nodes.properties
		FROM nodes
		INNER JOIN files ON nodes.file = files.file
		WHERE nodes.level == 2 AND files.file LIKE '%/feeds.org%'`)
	if err != nil {
		log.Fatal(err)
	}
	var props roam.Props
	for rows.Next() {
		if err := rows.Scan(&props); err != nil {
			log.Fatal(err)
		}
		if _, found := props.Tags["fomo"]; found {
			if data, err := props.ItemLinkData(); err == nil {
				url, _ := strings.CutPrefix(data.Url, "elfeed:")
				url = strings.Trim(url, " ")
				items = append(
					items,
					alfred.Item{
						Title: data.Title,
						// space at the end is to make searching in elfeed nicer so after typing / new search
						// term can be added without worrying about typing a space.
						Arg:      url + " ",
						Subtitle: url,
					},
				)
			}
		}
	}
	return items
}

func init() {
	roamCmd.AddCommand(elfeedCmd)
	elfeedCmd.AddCommand(elfeedItemsCmd)
	elfeedCmd.AddCommand(elfeedResolveCmd)
}
