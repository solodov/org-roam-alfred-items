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

	_ "github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/solodov/org-roam-alfred-items/roam"
	"github.com/spf13/cobra"
)

var chromeCmd = &cobra.Command{
	Use:   "chrome --category cat [--query query]",
	Short: "Output chrome alfred items matching the argument",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", roamCmdArgs.dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		rows, err := db.Query(`
			SELECT nodes.properties
			FROM nodes
			INNER JOIN files ON nodes.file = files.file
			WHERE nodes.level == 2 AND files.file LIKE '%/chrome.org%'`)
		if err != nil {
			log.Fatal(err)
		}
		var (
			props roam.Props
			items []alfred.Item
		)
		for rows.Next() {
			if err := rows.Scan(&props); err != nil {
				log.Fatal(err)
			}
			if props.Category != chromeCmdArgs.category {
				continue
			}
			if data, err := props.ItemLinkData(); err != nil {
				continue
			} else if strings.Contains(data.Title, chromeCmdArgs.query) || strings.Contains(props.Aliases, chromeCmdArgs.query) {
				items = append(
					items,
					alfred.Item{
						Title:        data.Title,
						Subtitle:     data.Url,
						Arg:          data.Url,
						Autocomplete: data.Url,
						Icon:         pickIcon(props.Icon, strings.ReplaceAll(data.Title, " ", "_")),
						Variables:    alfred.Variables{BrowserOverride: props.BrowserOverride},
					})
				if len(items) == 1 {
					items = append(items, makeDynamicItems(chromeCmdArgs.query)...)
				}
			}
		}
		if len(items) == 0 {
			items = makeDynamicItems(chromeCmdArgs.query)
		}
		for i := range items {
			items[i].Variables.Profile = chromeCmdArgs.category
		}
		printJson(alfred.Result{Items: items})
	},
}

func makeDynamicItems(alfredQuery string) (items []alfred.Item) {
	if alfredQuery == "" {
		return items
	}
	if u, err := url.Parse(alfredQuery); err == nil && (strings.HasPrefix(u.Scheme, "http") || u.Scheme == "chrome") {
		items = append(
			items,
			alfred.Item{
				Title: fmt.Sprintf(`open "%v"`, alfredQuery),
				Arg:   alfredQuery,
				Icon:  pickIcon("chrome"),
			})
	} else if chromeCmdArgs.category == "home" {
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
	return items
}

func pickIcon(bases ...string) (icon alfred.Icon) {
	dir := filepath.Join(chromeCmdArgs.orgDir, "alfred", "images")
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
	roamCmd.AddCommand(chromeCmd)
	u, _ := user.Current()
	chromeCmd.Flags().StringVar(&chromeCmdArgs.orgDir, "org_dir", filepath.Join(u.HomeDir, "org"), "Org directory")
	chromeCmd.Flags().StringVar(&chromeCmdArgs.query, "query", "", "Alfred input query")
	chromeCmd.Flags().StringVar(&chromeCmdArgs.category, "category", "", "Category to limit items to")
	chromeCmd.MarkFlagRequired("category")
}
