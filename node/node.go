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
	olpParts := []string{fileTitle}
	if level > 0 {
		if nodeOlp.Valid {
			for _, s := range olpRe.FindAllStringSubmatch(nodeOlp.String, -1) {
				olpParts = append(olpParts, s[1])
			}
		}
		olpParts = append(olpParts, nodeTitle)
	}
	fmt.Fprint(&titleBuilder, strings.Join(olpParts, " > "))
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

func (node Node) Match(titleRe *regexp.Regexp) bool {
	// TODO: parameterize definition of boring to support extraction of chrome or feed links
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

type Props struct {
	Path     string
	Category string
	Tags     map[string]bool
}

func (props *Props) Scan(src any) error {
	val, ok := src.(string)
	if !ok {
		return errors.New(fmt.Sprint("wrong source type, want string, got", reflect.TypeOf(src)))
	}
	props.Path = ""
	if matches := fileRe.FindStringSubmatch(val); len(matches) > 0 {
		props.Path = matches[1]
	}
	props.Category = ""
	if matches := catRe.FindStringSubmatch(val); len(matches) > 0 {
		props.Category = matches[1]
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
	catRe,
	fileRe,
	tagsRe,
	olpRe *regexp.Regexp
)

func init() {
	catRe = regexp.MustCompile(`"CATEGORY" \. "([^"]+)"`)
	fileRe = regexp.MustCompile(`"FILE" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`"ALLTAGS" \. .{0,2}":([^"]+):"`)
	olpRe = regexp.MustCompile(`"((?:\\.|[^"])*)"`)
}
