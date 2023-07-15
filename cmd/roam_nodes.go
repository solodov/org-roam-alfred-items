/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/roam"
	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:                   "nodes [--category category] [--query regex]",
	DisableFlagsInUseLine: true,
	Short:                 "Find matching org roam nodes and output them as alfred items",
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", roamCmdArgs.dbPath)
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
			items                    []alfred.Item
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
			if nodesCmdArgs.category != "" && props.Category != "any" && props.Category != nodesCmdArgs.category {
				continue
			}
			if strings.Contains(props.Path, "/drive/") {
				continue
			}
			if props.Tags.ContainsAnyOf([]string{"ARCHIVE", "feeds", "chrome_link"}) {
				continue
			}
			title := makeNodeTitle(level, props, fileTitle, nodeTitle, olp)
			if titleRe != nil && !titleRe.MatchString(title) {
				continue
			}
			items = append(items, alfred.Item{Uid: id, Title: title, Arg: id, Subtitle: props.Path})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].Title < items[j].Title
		})
		printJson(alfred.Result{Items: items})
	},
}

func makeNodeTitle(level int, props roam.Props, fileTitle, nodeTitle string, nodeOlp sql.NullString) string {
	var titleBuilder strings.Builder
	if props.Category != "" {
		fmt.Fprint(&titleBuilder, props.Category, ": ")
	}
	fmt.Fprint(&titleBuilder, fileTitle)
	if level > 0 {
		fmt.Fprint(&titleBuilder, " > ")
		if nodeOlp.Valid {
			matches := olpRe.FindAllStringSubmatch(nodeOlp.String, -1)
			for _, match := range matches {
				fmt.Fprint(&titleBuilder, match[1], " > ")
			}
		}
		fmt.Fprint(&titleBuilder, nodeTitle)
	}
	if len(props.Tags) > 0 {
		tags := make([]string, len(props.Tags))
		for tag := range props.Tags {
			tags = append(tags, " #" + tag)
		}
		sort.Strings(tags)
		fmt.Fprint(&titleBuilder, strings.Join(tags, ""))
	}
	return titleBuilder.String()
}

var nodesCmdArgs struct {
	category, query string
}

var olpRe *regexp.Regexp

func init() {
	roamCmd.AddCommand(nodesCmd)
	nodesCmd.Flags().StringVar(&nodesCmdArgs.category, "category", "", "Category to limit items to")
	nodesCmd.Flags().StringVar(&nodesCmdArgs.query, "query", "", "Alfred input query")
	olpRe = regexp.MustCompile(`"((?:\\.|[^"])*)"`)
}