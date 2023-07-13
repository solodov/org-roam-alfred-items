/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"database/sql"
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
		rows, err := db.Query(`SELECT nodes.properties
FROM nodes
INNER JOIN files ON nodes.file = files.file
WHERE nodes.level == 2 AND files.file LIKE '%/chrome.org%'`)
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
							Icon:         pickIcon(props.Icon, strings.ReplaceAll(title, " ", "_")),
							Variables:    alfred.Variables{BrowserOverride: props.BrowserOverride},
						})
					if len(result.Items) == 1 {
						result.Items = append(result.Items, makeDynamicItems(chromeCmdArgs.query)...)
					}
				}
			}
		}
		if len(result.Items) == 0 {
			result.Items = makeDynamicItems(chromeCmdArgs.query)
		}
		for i := range result.Items {
			result.Items[i].Variables.Profile = chromeCmdArgs.category
		}
		printJson(result)
	},
}

func makeDynamicItems(alfredQuery string) []alfred.Item {
	items := []alfred.Item{}
	if alfredQuery != "" {
		if u, err := url.Parse(alfredQuery); err == nil && (strings.HasPrefix(u.Scheme, "http") || u.Scheme == "chrome") {
			items = append(
				items,
				alfred.Item{
					Title: fmt.Sprintf(`open "%v"`, alfredQuery),
					Arg:   alfredQuery,
					Icon:  pickIcon("chrome"),
				})
		} else {
			if chromeCmdArgs.category == "home" {
				items = append(
					items,
					alfred.Item{
						Title: fmt.Sprintf(`search google for "%v"`, alfredQuery),
						Arg:   "https://www.google.com/search?q=" + alfredQuery,
						Icon:  pickIcon("chrome"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search map for "%v"`, alfredQuery),
						Arg:   "https://www.google.com/maps/search/" + alfredQuery,
						Icon:  pickIcon("map"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search youtube for "%v"`, alfredQuery),
						Arg:   "https://www.youtube.com/results?search_query=" + alfredQuery,
						Icon:  pickIcon("youtube"),
					},
				)
			} else if chromeCmdArgs.category == "goog" {
				items = append(
					items,
					alfred.Item{
						Title: fmt.Sprintf(`search moma for "%v"`, alfredQuery),
						Arg:   "https://moma.corp.google.com/search?q" + alfredQuery,
						Icon:  pickIcon("moma"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`code search for "%v"`, alfredQuery),
						Arg:   "https://source.corp.google.com/search?q=" + alfredQuery,
						Icon:  pickIcon("cs"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search google for "%v"`, alfredQuery),
						Arg:   "https://www.google.com/search?q=" + alfredQuery,
						Icon:  pickIcon("search"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search glossary for "%v"`, alfredQuery),
						Arg:   "https://moma.corp.google.com/search?hq=type:glossary&q=" + alfredQuery,
						Icon:  pickIcon("glossary"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search who for "%v"`, alfredQuery),
						Arg:   "https://moma.corp.google.com/search?hq=type:people&q=" + alfredQuery,
						Icon:  pickIcon("who"),
					},
					alfred.Item{
						Title: fmt.Sprintf(`search go links for "%v"`, alfredQuery),
						Arg:   "https://moma.corp.google.com/go2/search?q=" + alfredQuery,
						Icon:  pickIcon("go_links"),
					},
				)
			}
		}
	}
	return items
}

func pickIcon(bases ...string) alfred.Icon {
	dir := filepath.Join(chromeCmdArgs.orgDir, "alfred", "images")
	icon := alfred.Icon{}
	for _, base := range bases {
		path := filepath.Join(dir, base+".png")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			icon.Path = path
			break
		}
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
