package node

import (
	"database/sql"
	"reflect"
	"regexp"
	"strings"
)

type ObsoleteNode struct {
	Id        string
	Level     int
	Props     string
	Path      string
	FileTitle string
	NodeTitle string
	Olp       sql.NullString
	fullOlp   *string
	tags      *map[string]bool
	category  *string
}

func (n *ObsoleteNode) Cleanup() {
	v := reflect.ValueOf(n).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Type().Kind() == reflect.String {
			f.SetString(strings.Trim(f.String(), `"`))
		}
	}
}

func (n *ObsoleteNode) Category() *string {
	if n.category == nil {
		re := regexp.MustCompile(`.+"CATEGORY" \. "([^"]+)"`)
		var cat string
		if s := re.FindStringSubmatch(n.Props); len(s) > 0 {
			cat = s[1]
		}
		n.category = &cat
	}
	return n.category
}

func (n *ObsoleteNode) Tags() *map[string]bool {
	if n.tags == nil {
		tagsRe := regexp.MustCompile(`.+"ALLTAGS" \. #\(":([^"]+):"`)
		tags := make(map[string]bool)
		if s := tagsRe.FindStringSubmatch(n.Props); len(s) > 0 {
			for _, v := range strings.Split(s[1], ":") {
				tags[v] = true
			}
		}
		n.tags = &tags
	}
	return n.tags
}

func (n *ObsoleteNode) FullOlp() *string {
	if n.fullOlp == nil {
		olpParts := []string{n.FileTitle}
		if n.Level > 0 {
			if n.Olp.Valid {
				re := regexp.MustCompile(`"(?:\\.|[^\\"])*"`)
				for _, m := range re.FindAllStringSubmatch(n.Olp.String, -1) {
					olpParts = append(olpParts, strings.Trim(m[0], `"`))
				}
			}
			olpParts = append(olpParts, n.NodeTitle)
		}
		olp := strings.Join(olpParts, " > ")
		n.fullOlp = &olp
	}
	return n.fullOlp
}

func (n *ObsoleteNode) IsTaggedWith(t string) bool {
	_, found := (*n.Tags())[t]
	return found
}

func (n *ObsoleteNode) IsBoring() bool {
	return strings.HasPrefix(n.FileTitle, "drive-shard") || n.IsTaggedWith("ARCHIVE")
}
