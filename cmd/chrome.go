/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/node"
	"github.com/spf13/cobra"
)

var chromeCmd = &cobra.Command{
	Use:   "chrome [--category cat] [--query query]",
	Short: "Output chrome alfred items matching the argument",
	Long:  `Output chrome alfred items matching the argument`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", rootCmdArgs.dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		rows, err := db.Query(chromeLinkPropsQuery)
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
			if chromeCmdArgs.category != "" && props.Category != chromeCmdArgs.category {
				continue
			}
			if url, title, err := props.ItemLinkData(); err == nil {
				if strings.Contains(title, chromeCmdArgs.query) || strings.Contains(props.Aliases, chromeCmdArgs.query) {
					result.Items = append(
						result.Items,
						alfred.Item{
							Title:        title,
							Subtitle:     url,
							Arg:          url,
							Autocomplete: url,
							Icon:         makeIcon(strings.ReplaceAll(title, " ", "_"), props.Icon),
							Variables:    alfred.Variables{BrowserOverride: props.BrowserOverride},
						})
					if len(result.Items) == 1 {
						result.Items = append(result.Items, makeDynamicItems(chromeCmdArgs.query)...)
					}
				}
			}
		}
		if len(result.Items) == 0 {
			result.Items = append(result.Items, makeDynamicItems(chromeCmdArgs.query)...)
		}
		for i := range result.Items {
			result.Items[i].Variables.Profile = chromeCmdArgs.category
		}
		if jsonResult, err := json.Marshal(result); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(jsonResult))
		}
	},
}

const chromeLinkPropsQuery = `SELECT
  nodes.properties
FROM nodes
INNER JOIN files ON nodes.file = files.file
WHERE nodes.level == 2 AND files.file LIKE '%/chrome.org%'`

func makeDynamicItems(alfredQuery string) []alfred.Item {
	items := []alfred.Item{}
	if alfredQuery != "" {
		if u, err := url.Parse(alfredQuery); err == nil && (strings.HasPrefix(u.Scheme, "http") || u.Scheme == "chrome") {
			items = append(
				items,
				alfred.Item{
					Title: fmt.Sprintf(`open "%v"`, alfredQuery),
					Arg:   alfredQuery,
					Icon:  makeIcon("chrome", ""),
				})
		} else {
			items = append(
				items,
				alfred.Item{
					Title: fmt.Sprintf(`search google for "%v"`, alfredQuery),
					Arg:   "https://www.google.com/search?q=" + alfredQuery,
					Icon:  makeIcon("chrome", ""),
				},
				alfred.Item{
					Title: fmt.Sprintf(`search map for "%v"`, alfredQuery),
					Arg:   "https://www.google.com/maps/search/" + alfredQuery,
					Icon:  makeIcon("map", ""),
				},
				alfred.Item{
					Title: fmt.Sprintf(`search youtube for "%v"`, alfredQuery),
					Arg:   "https://www.youtube.com/results?search_query=" + alfredQuery,
					Icon:  makeIcon("youtube", ""),
				},
			)
		}
	}
	return items
}

func makeIcon(titleBase, overrideBase string) alfred.Icon {
	baseDir := filepath.Join(chromeCmdArgs.orgDir, "alfred", "images")
	iconPath := ""
	if overrideBase != "" {
		iconPath = filepath.Join(baseDir, overrideBase+".png")
	} else {
		iconPath = filepath.Join(baseDir, titleBase+".png")
	}
	icon := alfred.Icon{}
	if info, err := os.Stat(iconPath); err == nil && !info.IsDir() {
		icon.Path = iconPath
	}
	return icon
}

var chromeCmdArgs struct {
	orgDir, query, category string
}

func init() {
	rootCmd.AddCommand(chromeCmd)
	u, _ := user.Current()
	chromeCmd.Flags().StringVar(&chromeCmdArgs.orgDir, "org_dir", filepath.Join(u.HomeDir, "org"), "Org directory")
	chromeCmd.Flags().StringVar(&chromeCmdArgs.query, "query", "", "Alfred input query")
	chromeCmd.Flags().StringVar(&chromeCmdArgs.category, "category", "", "Category to limit items to")
}
