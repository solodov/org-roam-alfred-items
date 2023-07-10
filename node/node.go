/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package node

import (
	"database/sql"
	"regexp"
	"strings"
)

type Node struct {
	Id       string
	Path     string
	Olp      string
	Tags     map[string]bool
	Category string
}

func New(id string, level int, props, path, fileTitle, nodeTitle string, nodeOlp sql.NullString) Node {
	category := ""
	if s := catRe.FindStringSubmatch(props); len(s) > 0 {
		category = s[1]
	}
	tags := make(map[string]bool)
	if s := tagsRe.FindStringSubmatch(props); len(s) > 0 {
		for _, v := range strings.Split(s[1], ":") {
			tags[v] = true
		}
	}
	olpParts := []string{strings.Trim(fileTitle, `"`)}
	if level > 0 {
		if nodeOlp.Valid {
			for _, s := range olpRe.FindAllStringSubmatch(nodeOlp.String, -1) {
				olpParts = append(olpParts, strings.Trim(s[0], `"`))
			}
		}
		olpParts = append(olpParts, strings.Trim(nodeTitle, `"`))
	}
	return Node{
		Id:       strings.Trim(id, `"`),
		Path:     strings.Trim(path, `"`),
		Olp:      strings.Join(olpParts, " > "),
		Tags:     tags,
		Category: category,
	}
}

func (n Node) IsBoring() bool {
	if _, found := n.Tags["ARCHIVE"]; found {
		return true
	}
	if strings.HasPrefix(n.Olp, "drive-shard") {
		return true
	}
	return false
}

var catRe, tagsRe, olpRe *regexp.Regexp

func init() {
	catRe = regexp.MustCompile(`.+"CATEGORY" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`.+"ALLTAGS" \. #\(":([^"]+):"`)
	olpRe = regexp.MustCompile(`"(?:\\.|[^\\"])*"`)
}
