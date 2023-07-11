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
	Id         string
	Path       string
	olp        string
	Title      string
	Properties Props
	isBoring   bool
}

func New(id string, level int, props Props, path, fileTitle, nodeTitle string, nodeOlp sql.NullString) Node {
	isBoring := strings.HasPrefix(fileTitle, "drive-shard")
	olpParts := []string{fileTitle}
	if level > 0 {
		if nodeOlp.Valid {
			for _, s := range olpRe.FindAllStringSubmatch(nodeOlp.String, -1) {
				olpParts = append(olpParts, s[1])
			}
		}
		olpParts = append(olpParts, nodeTitle)
	}
	olp := strings.Join(olpParts, " > ")
	var titleBuilder strings.Builder
	if props.Category != "" {
		fmt.Fprintf(&titleBuilder, "%v: ", props.Category)
	}
	fmt.Fprint(&titleBuilder, olp)
	if len(props.Tags) > 0 {
		fmt.Fprint(&titleBuilder, " ")
		for _, t := range props.Tags {
			fmt.Fprintf(&titleBuilder, " #%v", t)
			isBoring = isBoring || t == "ARCHIVE"
		}
	}
	return Node{
		Id:         id,
		Path:       path,
		olp:        olp,
		Title:      titleBuilder.String(),
		Properties: props,
		isBoring:   isBoring,
	}
}

func (n Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Uid      string `json:"uid"`
		Title    string `json:"title"`
		Arg      string `json:"arg"`
		Subtitle string `json:"subtitle"`
	}{n.Id, n.Title, n.Id, n.Path})
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
	Category string
	Tags     []string
}

func (p *Props) Scan(src any) error {
	strVal, ok := src.(string)
	if !ok {
		return errors.New(fmt.Sprint("wrong source type, want string, got", reflect.TypeOf(src)))
	}
	p.Category = ""
	if matches := catRe.FindStringSubmatch(strVal); len(matches) > 0 {
		p.Category = matches[1]
	}
	p.Tags = nil
	if matches := tagsRe.FindStringSubmatch(strVal); len(matches) > 0 {
		for _, tag := range strings.Split(matches[1], ":") {
			p.Tags = append(p.Tags, tag)
		}
	}
	sort.Strings(p.Tags)
	return nil
}

var (
	catRe,
	tagsRe,
	olpRe *regexp.Regexp
)

func init() {
	catRe = regexp.MustCompile(`.+"CATEGORY" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`.+"ALLTAGS" \. #\(":([^"]+):"`)
	olpRe = regexp.MustCompile(`"((?:\\.|[^"])*)"`)
}
