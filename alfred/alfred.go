package alfred

type Item struct {
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle,omitempty"`
	Autocomplete string    `json:"autocomplete,omitempty"`
	Arg          string    `json:"arg"`
	Icon         Icon      `json:"icon,omitempty"`
	Variables    Variables `json:"variables,omitempty"`
}

type Icon struct {
	Path string `json:"path"`
}

type Variables struct {
	BrowserOverride string `json:"browser_override,omitempty"`
	Profile         string `json:"profile,omitempty"`
}

type Result struct {
	Items []Item `json:"items"`
}