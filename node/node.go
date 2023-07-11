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
	Id       string
	Title    string
	Props    Props
	isBoring bool
}

func New(id string, level int, props Props, fileTitle, nodeTitle string, nodeOlp sql.NullString) Node {
	isBoring := strings.HasPrefix(fileTitle, "drive-shard")
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
		for _, tag := range props.Tags {
			fmt.Fprint(&titleBuilder, " #", tag)
			// TODO: parameterize definition of boring to support extraction of chrome or feed links
			isBoring = isBoring || tag == "ARCHIVE" || tag == "feeds" || tag == "chrome_link"
		}
	}
	return Node{
		Id:       id,
		Title:    titleBuilder.String(),
		Props:    props,
		isBoring: isBoring,
	}
}

func (n Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Uid      string `json:"uid"`
		Title    string `json:"title"`
		Arg      string `json:"arg"`
		Subtitle string `json:"subtitle"`
	}{n.Id, n.Title, n.Id, n.Props.Path})
}

func (n Node) Match(r *regexp.Regexp) bool {
	if n.isBoring {
		return false
	}
	if r == nil {
		return true
	}
	return r.MatchString(n.Title)
}

type Props struct {
	Path     string
	Category string
	// TODO: make this a set for easy lookups
	Tags []string
}

func (p *Props) Scan(src any) error {
	val, ok := src.(string)
	if !ok {
		return errors.New(fmt.Sprint("wrong source type, want string, got", reflect.TypeOf(src)))
	}
	p.Path = ""
	if matches := fileRe.FindStringSubmatch(val); len(matches) > 0 {
		p.Path = matches[1]
	}
	p.Category = ""
	if matches := catRe.FindStringSubmatch(val); len(matches) > 0 {
		p.Category = matches[1]
	}
	p.Tags = nil
	if matches := tagsRe.FindStringSubmatch(val); len(matches) > 0 {
		p.Tags = strings.Split(matches[1], ":")
	}
	sort.Strings(p.Tags)
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
