/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package roam

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type TagSet map[string]bool

func (ts *TagSet) ContainsAnyOf(tags []string) bool {
	for _, tag := range tags {
		if _, found := (*ts)[tag]; found {
			return true
		}
	}
	return false
}

type Props struct {
	Path            string
	Category        string
	Item            string
	Aliases         string
	Icon            string
	BrowserOverride string
	Tags            TagSet
}

func (props *Props) ItemLinkData() (string, string, error) {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\[([^\]]+)\]\]`)
	if groups := re.FindStringSubmatch(props.Item); len(groups) > 0 {
		return groups[1], groups[2], nil
	}
	return "", "", errors.New("not a proper link")
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

var simplePropertyRe, tagsRe *regexp.Regexp

func init() {
	simplePropertyRe = regexp.MustCompile(`"([^"]+)" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`"ALLTAGS" \. .{0,2}":([^"]+):"`)
}