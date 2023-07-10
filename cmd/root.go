/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/node"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                   "org-roam-alfred-items [-c category] [regex]",
	DisableFlagsInUseLine: true,
	Short:                 "Find matching org roam nodes and output them as alfred items",
	Long:                  "Find matching org roam nodes and output them as alfred items",
	Args:                  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		var titleRe *regexp.Regexp
		if len(args) == 1 {
			// TODO: collapse consecutive spaces into one prior to the replacement
			titleRe = regexp.MustCompile("(?i)" + strings.ReplaceAll(args[0], " ", ".*"))
		}
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		var (
			level                                 int
			id, props, path, fileTitle, nodeTitle string
			olp                                   sql.NullString
		)
		nodes := []node.Node{}
		for rows.Next() {
			if err := rows.Scan(&id, &level, &props, &path, &fileTitle, &nodeTitle, &olp); err != nil {
				log.Fatal(err)
			}
			if n := node.New(id, level, props, path, fileTitle, nodeTitle, olp); !n.IsBoring() && n.Match(titleRe) {
				nodes = append(nodes, n)
			}
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Olp < nodes[j].Olp
		})
		result := struct {
			Items []node.Node `json:"items"`
		}{nodes}
		if jsonResult, err := json.Marshal(result); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(jsonResult))
		}
	},
}

const query = `SELECT
  nodes.id,
  nodes.level,
  nodes.properties,
  files.file,
  files.title,
  nodes.title,
  nodes.olp
FROM nodes
INNER JOIN files ON nodes.file = files.file`

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var dbPath string

func init() {
	u, _ := user.Current()
	rootCmd.Flags().StringVar(&dbPath, "db_path", filepath.Join(u.HomeDir, "org/.roam.db"), "Path to the org roam database")
}
