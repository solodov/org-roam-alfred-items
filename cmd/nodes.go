/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/roam"
	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:                   "nodes [--category category] [--query regex]",
	DisableFlagsInUseLine: true,
	Short:                 "Find matching org roam nodes and output them as alfred items",
	Long:                  "Find matching org roam nodes and output them as alfred items",
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", rootCmdArgs.dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		var titleRe *regexp.Regexp
		if nodesCmdArgs.query != "" {
			// TODO: collapse consecutive spaces into one prior to the replacement
			titleRe = regexp.MustCompile("(?i)" + strings.ReplaceAll(nodesCmdArgs.query, " ", ".*"))
		}
		rows, err := db.Query(`SELECT
  nodes.id,
  nodes.level,
  nodes.properties,
  files.title,
  nodes.title,
  nodes.olp
FROM nodes
INNER JOIN files ON nodes.file = files.file`)
		if err != nil {
			log.Fatal(err)
		}
		var (
			level                    int
			id, fileTitle, nodeTitle string
			props                    roam.Props
			olp                      sql.NullString
			nodes                    []roam.Node
		)
		scan := func(args ...any) error {
			if err := rows.Scan(args...); err != nil {
				return err
			}
			for _, a := range args {
				if v, ok := a.(*string); ok {
					*v, _ = strconv.Unquote(*v)
					*v = strings.ReplaceAll(*v, `\"`, `"`)
				}
			}
			return nil
		}
		for rows.Next() {
			if err := scan(&id, &level, &props, &fileTitle, &nodeTitle, &olp); err != nil {
				log.Fatal(err)
			}
			if node := roam.NewNode(id, level, props, fileTitle, nodeTitle, olp); matchNode(node, titleRe) {
				nodes = append(nodes, node)
			}
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Title < nodes[j].Title
		})
		printJson(struct {
			Items []roam.Node `json:"items"`
		}{Items: nodes})
	},
}

func matchNode(node roam.Node, titleRe *regexp.Regexp) bool {
	if nodesCmdArgs.category != "" && node.Props.Category != "any" && node.Props.Category != nodesCmdArgs.category {
		return false
	}
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

var nodesCmdArgs struct {
	category, query string
}

func init() {
	rootCmd.AddCommand(nodesCmd)
	nodesCmd.Flags().StringVar(&nodesCmdArgs.category, "category", "", "Category to limit items to")
	nodesCmd.Flags().StringVar(&nodesCmdArgs.query, "query", "", "Alfred input query")
}
