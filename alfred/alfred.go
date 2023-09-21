package alfred

import "encoding/json"

type Item struct {
	Uid          string    `json:"uid,omitempty"`
	Title        string    `json:"title,omitempty"`
	Subtitle     string    `json:"subtitle,omitempty"`
	Autocomplete string    `json:"autocomplete,omitempty"`
	Text         Text      `json:"text,omitempty"`
	Arg          string    `json:"arg,omitempty"`
	Icon         Icon      `json:"icon,omitempty"`
	Variables    Variables `json:"variables,omitempty"`
	Valid        bool      `json:"valid,omitempty"`
	Save         bool      `json:"-"` // indicates whether this item should be saved in history
}

type Icon struct {
	Path string `json:"path"`
}

type Variables struct {
	BrowserOverride string `json:"browser_override,omitempty"`
	Profile         string `json:"profile,omitempty"`
	BrowserState    string `json:"browser_state,omitempty"`
	Meeting         string `json:"meeting,omitempty"`
	ClockedInTask   string `json:"clocked_in_task,omitempty"`
	Template        string `json:"template,omitempty"`
	Arg             string `json:"arg,omitempty"`
	HistItem        string `json:"hist_item,omitempty"`
	Query           string `json:"query"`
}

type Result struct {
	Items     []Item    `json:"items"`
	Variables Variables `json:"variables,omitempty"`
}

type Text struct {
	Copy      string `json:"copy"`
	LargeType string `json:"largetype"`
}

type BrowserState struct {
	Url, Title string
}

func (v Variables) DecodeBrowserState() (b *BrowserState) {
	if v.BrowserState != "nil" {
		b = &BrowserState{}
		json.Unmarshal([]byte(v.BrowserState), b)
	}
	return b
}

func (b BrowserState) String() string {
	if b.Title != "" {
		return b.Title
	} else {
		return b.Url
	}
}
