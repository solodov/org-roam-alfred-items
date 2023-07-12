/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/node"
	"github.com/spf13/cobra"
)

var elfeedCmd = &cobra.Command{
	Use:                   "elfeed [--query query]",
	DisableFlagsInUseLine: true,
	Short:                 "Output elfeed alfred items",
	Long:                  "Output elfeed alfred items",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", rootCmdArgs.dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		rows, err := db.Query(elfeedPropsQuery)
		if err != nil {
			log.Fatal(err)
		}
		var (
			props  node.Props
			result alfred.Result
		)
		for rows.Next() {
			if err := rows.Scan(&props); err != nil {
				log.Fatal(err)
			}
			if _, found := props.Tags["fomo"]; !found {
				continue
			}
			if url, title, err := props.ItemLinkData(); err == nil {
				url, _ = strings.CutPrefix(url, "elfeed:")
				result.Items = append(
					result.Items,
					alfred.Item{
						Title:    title,
						Arg:      url,
						Subtitle: strings.Trim(url, " "),
					},
				)
			}
		}
		if jsonResult, err := json.Marshal(result); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(jsonResult))
		}
	},
}

const elfeedPropsQuery = `SELECT
  nodes.properties
FROM nodes
INNER JOIN files ON nodes.file = files.file
WHERE nodes.level == 2 AND files.file LIKE '%/feeds.org%'`

func init() {
	rootCmd.AddCommand(elfeedCmd)
}
