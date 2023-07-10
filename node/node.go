/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package node

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Node struct {
	Id       string
	Path     string
	Olp      string
	Title    string
	Tags     []string
	Category string
}

func New(id string, level int, props, path, fileTitle, nodeTitle string, nodeOlp sql.NullString) Node {
	category := ""
	if s := catRe.FindStringSubmatch(props); len(s) > 0 {
		category = s[1]
	}
	tags := []string{}
	if s := tagsRe.FindStringSubmatch(props); len(s) > 0 {
		for _, t := range strings.Split(s[1], ":") {
			tags = append(tags, t)
		}
	}
	sort.Strings(tags)
	olpParts := []string{strings.Trim(fileTitle, `"`)}
	if level > 0 {
		if nodeOlp.Valid {
			for _, s := range olpRe.FindAllStringSubmatch(nodeOlp.String, -1) {
				olpParts = append(olpParts, strings.Trim(s[0], `"`))
			}
		}
		olpParts = append(olpParts, strings.Trim(nodeTitle, `"`))
	}
	olp := strings.Join(olpParts, " > ")
	var titleBuilder strings.Builder
	if category != "" {
		fmt.Fprintf(&titleBuilder, "%v: ", category)
	}
	fmt.Fprint(&titleBuilder, olp)
	if len(tags) > 0 {
		fmt.Fprint(&titleBuilder, " ")
		for _, t := range tags {
			fmt.Fprintf(&titleBuilder, " #%v", t)
		}
	}
	return Node{
		Id:       strings.Trim(id, `"`),
		Path:     strings.Trim(path, `"`),
		Olp:      olp,
		Title:    titleBuilder.String(),
		Tags:     tags,
		Category: category,
	}
}

func (n Node) MarshalJSON() ([]byte, error) {
	jn := struct {
		Uid      string `json:"uid"`
		Title    string `json:"title"`
		Arg      string `json:"arg"`
		Subtitle string `json:"subtitle"`
	}{n.Id, n.Title, n.Id, n.Path}
	return json.Marshal(jn)
}

func (n Node) IsBoring() bool {
	for _, t := range n.Tags {
		if t == "ARCHIVE" {
			return true
		}
	}
	if strings.HasPrefix(n.Olp, "drive-shard") {
		return true
	}
	return false
}

func (n Node) Match(r *regexp.Regexp) bool {
	if r == nil {
		return true
	}
	return r.MatchString(n.Title)
}

var catRe, tagsRe, olpRe *regexp.Regexp

func init() {
	catRe = regexp.MustCompile(`.+"CATEGORY" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`.+"ALLTAGS" \. #\(":([^"]+):"`)
	olpRe = regexp.MustCompile(`"(?:\\.|[^\\"])*"`)
}
