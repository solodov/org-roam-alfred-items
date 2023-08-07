package alfred

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
}

type Icon struct {
	Path string `json:"path"`
}

type Variables struct {
	BrowserOverride string `json:"browser_override,omitempty"`
	Profile         string `json:"profile,omitempty"`
	BrowserState    string `json:"browser_state"`
	Meeting         string `json:"meeting"`
	ClockedInTask   string `json:"clocked_in_task"`
	Action          string `json:"action,omitempty"`
	Template        string `json:"template,omitempty"`
	Arg             string `json:"arg,omitempty"`
}

type Result struct {
	Items     []Item    `json:"items"`
	Variables Variables `json:"variables,omitempty"`
}

type Text struct {
	Copy      string `json:"copy"`
	LargeType string `json:"largetype"`
}
