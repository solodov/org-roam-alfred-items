/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package roam

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type Tags map[string]bool

func (ts Tags) ContainsAnyOf(tags []string) bool {
	for _, tag := range tags {
		if ts[tag] {
			return true
		}
	}
	return false
}

func (ts Tags) String() string {
	tags := make([]string, len(ts))
	for tag := range ts {
		tags = append(tags, " #"+tag)
	}
	sort.Strings(tags)
	return strings.Join(tags, "")
}

type Props struct {
	Path            string
	Category        string
	Item            string
	Aliases         string
	Icon            string
	BrowserOverride string
	Tags            Tags
}

func (props *Props) ItemLinkData() (string, string, error) {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\[([^\]]+)\]\]`)
	if groups := re.FindStringSubmatch(props.Item); len(groups) > 0 {
		return groups[1], groups[2], nil
	}
	return "", "", fmt.Errorf("not a proper link")
}

func (props *Props) Scan(src any) error {
	val, ok := src.(string)
	if !ok {
		return fmt.Errorf("wrong source type, want string, got %v", reflect.TypeOf(src))
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
	props.Tags = Tags{}
	if matches := tagsRe.FindStringSubmatch(val); len(matches) > 0 {
		for _, tag := range strings.Split(matches[1], ":") {
			props.Tags[tag] = true
		}
	}
	return nil
}

var simplePropertyRe, tagsRe *regexp.Regexp

func init() {
	simplePropertyRe = regexp.MustCompile(`"([^"]+)" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`"ALLTAGS" \. .{0,2}":([^"]+):"`)
}
