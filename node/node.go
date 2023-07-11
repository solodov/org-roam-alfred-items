/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package node

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type Node struct {
	Id    string
	Title string
	Props Props
}

func New(id string, level int, props Props, fileTitle, nodeTitle string, nodeOlp sql.NullString) Node {
	var titleBuilder strings.Builder
	if props.Category != "" {
		fmt.Fprint(&titleBuilder, props.Category, ": ")
	}
	fmt.Fprint(&titleBuilder, fileTitle)
	if level > 0 {
		fmt.Fprint(&titleBuilder, " > ")
		if nodeOlp.Valid {
			matches := olpRe.FindAllStringSubmatch(nodeOlp.String, -1)
			for i, match := range matches {
				fmt.Fprint(&titleBuilder, match[1])
				if i < len(matches)-1 {
					fmt.Fprint(&titleBuilder, " > ")
				}
			}
		}
		fmt.Fprint(&titleBuilder, nodeTitle)
	}
	if len(props.Tags) > 0 {
		fmt.Fprint(&titleBuilder, " ")
		tags := []string{}
		for tag := range props.Tags {
			tags = append(tags, tag)
		}
		sort.Strings(tags)
		for _, tag := range tags {
			fmt.Fprint(&titleBuilder, " #", tag)
		}
	}
	return Node{
		Id:    id,
		Title: titleBuilder.String(),
		Props: props,
	}
}

func (node Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Uid      string `json:"uid"`
		Title    string `json:"title"`
		Arg      string `json:"arg"`
		Subtitle string `json:"subtitle"`
	}{node.Id, node.Title, node.Id, node.Props.Path})
}

type Props struct {
	Path            string
	Category        string
	Item            string
	Aliases         string
	Icon            string
	BrowserOverride string
	Tags            map[string]bool
}

func (props *Props) Scan(src any) error {
	val, ok := src.(string)
	if !ok {
		return errors.New(fmt.Sprint("wrong source type, want string, got", reflect.TypeOf(src)))
	}
	matchDests := map[string]*string{
		"FILE":             &props.Path,
		"CATEGORY":         &props.Category,
		"ITEM":             &props.Item,
		"ALIASES":          &props.Aliases,
		"ICON":             &props.Icon,
		"BROWSER_OVERRIDE": &props.BrowserOverride,
	}
	for _, dest := range matchDests {
		*dest = ""
	}
	if matches := simplePropertyRe.FindAllStringSubmatch(val, -1); len(matches) > 0 {
		for _, groups := range matches {
			if dest, found := matchDests[groups[1]]; found {
				*dest = groups[2]
			}
		}
	}
	// TODO: go 1.21 has new clear function that achieves the same:
	// clear(p.Tags)
	props.Tags = make(map[string]bool)
	if matches := tagsRe.FindStringSubmatch(val); len(matches) > 0 {
		for _, tag := range strings.Split(matches[1], ":") {
			props.Tags[tag] = true
		}
	}
	return nil
}

var (
	simplePropertyRe,
	tagsRe,
	olpRe *regexp.Regexp
)

func init() {
	simplePropertyRe = regexp.MustCompile(`"([^"]+)" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`"ALLTAGS" \. .{0,2}":([^"]+):"`)
	olpRe = regexp.MustCompile(`"((?:\\.|[^"])*)"`)
}
