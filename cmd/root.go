/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	// "regexp"
	// "strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type node struct {
	Id        string
	Level     int
	Props     string
	Path      string
	FileTitle string
	NodeTitle string
	Olp       sql.NullString
}

func (n *node) cleanup() {
	v := reflect.ValueOf(n).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Type().Kind() == reflect.String {
			f.SetString(strings.Trim(f.String(), "\""))
		}
	}
}

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
		nodeTitleMatchReStr := ""
		if len(args) == 1 {
			// TODO: collapse consecutive spaces into one prior to the replacement
			nodeTitleMatchReStr = strings.ReplaceAll(args[0], " ", ".*")
		}
		nodeTitleMatchRe := regexp.MustCompile("(?i)" + nodeTitleMatchReStr)
		fmt.Println(nodeTitleMatchRe)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var n node
			if err := rows.Scan(
				&n.Id,
				&n.Level,
				&n.Props,
				&n.Path,
				&n.FileTitle,
				&n.NodeTitle,
				&n.Olp,
			); err != nil {
				log.Fatal(err)
			}
			n.cleanup()
			if strings.HasPrefix(n.FileTitle, "drive-shard") {
				continue
			}
			tagsRe := regexp.MustCompile(`.+"ALLTAGS" \. #\(":([^"]+):"`)
			tags := make(map[string]bool)
			if s := tagsRe.FindStringSubmatch(n.Props); len(s) > 0 {
				for _, v := range strings.Split(s[1], ":") {
					tags[v] = true
				}
			}
			if _, found := tags["ARCHIVE"]; found {
				continue
			}
			categoryRe := regexp.MustCompile(`.+"CATEGORY" \. "([^"]+)"`)
			var category string
			if s := categoryRe.FindStringSubmatch(n.Props); len(s) > 0 {
				category = s[1]
				// Any category is included regardless of the value of the category argument.
				if category != "any" && categoryArg != "" && category != categoryArg {
					continue
				}
			}
			if nodeTitleMatchRe.MatchString(n.NodeTitle) {
				fmt.Println(n.NodeTitle, category, tags)
			}
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
var categoryArg string

func init() {
	u, _ := user.Current()
	rootCmd.Flags().StringVar(&dbPath, "db_path", filepath.Join(u.HomeDir, "org/.roam.db"), "Path to the org roam database")
	rootCmd.Flags().StringVarP(&categoryArg, "category", "c", "", "Limit notes to the category")
}
