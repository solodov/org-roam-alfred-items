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
	NewWindow       string
	Tags            Tags
}

func (props *Props) ItemLinkData() (data struct{ Url, Title string }, err error) {
	if groups := linkRe.FindStringSubmatch(props.Item); len(groups) > 0 {
		data.Url = groups[1]
		data.Title = groups[2]
	} else {
		err = fmt.Errorf("not a proper link")
	}
	return data, err
}

func (props *Props) Scan(src any) error {
	// Zero-out the receiver, Tags requires a special treatment because its zero value is nil, see
	// https://go.dev/ref/spec#The_zero_value
	*props = Props{Tags: Tags{}}
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
		"NEW_WINDOW":       &props.NewWindow,
	}
	if matches := simplePropertyRe.FindAllStringSubmatch(val, -1); len(matches) > 0 {
		for _, groups := range matches {
			if dest, found := matchDests[groups[1]]; found {
				*dest = groups[2]
			}
		}
	}
	if matches := tagsRe.FindStringSubmatch(val); len(matches) > 0 {
		for _, tag := range strings.Split(matches[1], ":") {
			props.Tags[tag] = true
		}
	}
	return nil
}

var simplePropertyRe, tagsRe, linkRe *regexp.Regexp

func init() {
	simplePropertyRe = regexp.MustCompile(`"([^"]+)" \. "([^"]+)"`)
	tagsRe = regexp.MustCompile(`"ALLTAGS" \. .{0,2}":([^"]+):"`)
	linkRe = regexp.MustCompile(`\[\[([^\]]+)\]\[([^\]]+)\]\]`)
}
