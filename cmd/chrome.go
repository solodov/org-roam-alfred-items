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

	"github.com/solodov/org-roam-alfred-items/node"
	"github.com/spf13/cobra"
)

var chromeCmd = &cobra.Command{
	Use:   "chrome [match]",
	Short: "Output chrome alfred items matching the argument",
	Long:  `Output chrome alfred items matching the argument`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		query := ""
		if len(args) == 1 {
			query = args[0]
		}
		rows, err := db.Query(chromeNodeQuery)
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
			Items []chromeNode `json:"items"`
		}{}
		for rows.Next() {
			if err := scan(rows, &id, &level, &props, &fileTitle, &nodeTitle, &olp); err != nil {
				log.Fatal(err)
			}
			node := node.New(id, level, props, fileTitle, nodeTitle, olp)
			if category != "" && node.Props.Category != category {
				continue
			}
			if url, title, err := node.Props.ItemLinkData(); err == nil {
				if strings.Contains(title, query) || strings.Contains(node.Props.Aliases, query) {
					icon := iconPath(node.Props.Icon, title)
					result.Items = append(
						result.Items,
						chromeNode{
							Title:        title,
							Subtitle:     url,
							Arg:          url,
							Autocomplete: url,
							Icon:         icon,
							Variables:    variables{Profile: category, BrowserOverride: node.Props.BrowserOverride},
						})
					if len(result.Items) == 1 {
						result.Items = append(result.Items, dynamicNodes(query)...)
					}
				}
			}
		}
		if len(result.Items) == 0 {
			result.Items = append(result.Items, dynamicNodes(query)...)
		}
		if jsonResult, err := json.Marshal(result); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(jsonResult))
		}
	},
}

const chromeNodeQuery = `SELECT
  nodes.id,
  nodes.level,
  nodes.properties,
  files.title,
  nodes.title,
  nodes.olp
FROM nodes
INNER JOIN files ON nodes.file = files.file
WHERE nodes.level == 2 AND files.file LIKE '%/chrome.org%'`

func dynamicNodes(query string) []chromeNode {
	nodes := []chromeNode{}
	if u, err := url.Parse(query); err == nil && (strings.HasPrefix(u.Scheme, "http") || u.Scheme == "chrome") {
		nodes = append(
			nodes,
			chromeNode{
				Title: fmt.Sprintf(`open "%v"`, query),
				Arg:   query,
				Icon:  icon{Path: filepath.Join(orgDir, "alfred", "images", "chrome.png")},
			})
	} else if query != "" {
		nodes = append(
			nodes,
			chromeNode{
				Title:     fmt.Sprintf(`search google for "%v"`, query),
				Arg:       "https://www.google.com/search?q=" + query,
				Icon:      icon{Path: filepath.Join(orgDir, "alfred", "images", "chrome.png")},
				Variables: variables{Profile: category},
			},
			chromeNode{
				Title:     fmt.Sprintf(`search map for "%v"`, query),
				Arg:       "https://www.google.com/maps/search/" + query,
				Icon:      icon{Path: filepath.Join(orgDir, "alfred", "images", "map.png")},
				Variables: variables{Profile: category},
			},
			chromeNode{
				Title:     fmt.Sprintf(`search youtube for "%v"`, query),
				Arg:       "https://www.youtube.com/results?search_query=" + query,
				Icon:      icon{Path: filepath.Join(orgDir, "alfred", "images", "youtube.png")},
				Variables: variables{Profile: category},
			},
		)
	}
	return nodes
}

type chromeNode struct {
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle,omitempty"`
	Autocomplete string    `json:"autocomplete,omitempty"`
	Arg          string    `json:"arg"`
	Icon         icon      `json:"icon,omitempty"`
	Variables    variables `json:"variables,omitempty"`
}

type icon struct {
	Path string `json:"path"`
}

type variables struct {
	BrowserOverride string `json:"browser_override,omitempty"`
	Profile         string `json:"profile"`
}

func iconPath(i, title string) icon {
	base := strings.ReplaceAll(title, " ", "_")
	if i != "" {
		base = i
	}
	path := filepath.Join(orgDir, "alfred", "images", base+".png")
	if _, err := os.Stat(path); err == nil {
		return icon{Path: path}
	}
	return icon{}
}

var orgDir string

func init() {
	rootCmd.AddCommand(chromeCmd)
	u, _ := user.Current()
	chromeCmd.Flags().StringVar(&orgDir, "org_dir", filepath.Join(u.HomeDir, "org"), "Org directory")
}
