/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/node"
	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:                   "nodes [-c category] [regex]",
	DisableFlagsInUseLine: true,
	Short:                 "Find matching org roam nodes and output them as alfred items",
	Long:                  "Find matching org roam nodes and output them as alfred items",
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
			level                    int
			id, fileTitle, nodeTitle string
			props                    node.Props
			olp                      sql.NullString
		)
		result := struct {
			Items []node.Node `json:"items"`
		}{}
		for rows.Next() {
			if err := scan(rows, &id, &level, &props, &fileTitle, &nodeTitle, &olp); err != nil {
				log.Fatal(err)
			}
			if node := node.New(id, level, props, fileTitle, nodeTitle, olp); matches(node, titleRe) {
				result.Items = append(result.Items, node)
			}
		}
		sort.Slice(result.Items, func(i, j int) bool {
			return result.Items[i].Title < result.Items[j].Title
		})
		if res, err := json.Marshal(result); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(res))
		}
	},
}

func matches(node node.Node, titleRe *regexp.Regexp) bool {
	for _, tag := range []string{"ARCHIVE", "feeds", "chrome_link"} {
		if _, found := node.Props.Tags[tag]; found {
			return false
		}
	}
	if strings.Contains(node.Props.Path, "/drive/") {
		return false
	}
	if titleRe == nil {
		return true
	}
	return titleRe.MatchString(node.Title)
}

func scan(rows *sql.Rows, args ...any) error {
	if err := rows.Scan(args...); err != nil {
		return err
	}
	for _, a := range args {
		v, ok := a.(*string)
		if ok {
			*v = strings.ReplaceAll(strings.Trim(*v, `"`), `\"`, `"`)
		}
	}
	return nil
}

const query = `SELECT
  nodes.id,
  nodes.level,
  nodes.properties,
  files.title,
  nodes.title,
  nodes.olp
FROM nodes
INNER JOIN files ON nodes.file = files.file`

func init() {
	rootCmd.AddCommand(nodesCmd)
}