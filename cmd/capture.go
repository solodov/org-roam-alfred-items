/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use: "capture",
}

var captureItemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Perform org capture",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result := alfred.Result{}
		initVariables(&result.Variables)
		bs := browserState{}
		if result.Variables.BrowserState != "" {
			json.Unmarshal([]byte(result.Variables.BrowserState), &bs)
		}
		addItem := func(title, template string, valid bool) {
			result.Items = append(result.Items, captureItem(title, template, valid))
		}
		if captureCmdArgs.category == "home" {
			if result.Variables.Meeting != "" {
				addItem(fmt.Sprintf("capture meeting notes for \"%v\"", result.Variables.Meeting), "e", true)
			}
			addItem("capture note into inbox", "h", captureCmdArgs.query != "")
			if result.Variables.ClockedInTask != "" {
				addItem("capture note for the clocked-in task", "c", captureCmdArgs.query != "")
			}
			if bs.Url != "" {
				addItem(fmt.Sprintf("capture \"%s\" into inbox", bs), "bh", true)
			}
			if result.Variables.Meeting == "" {
				addItem("capture meeting notes for unknown meeting", "e", true)
			}
		} else if captureCmdArgs.category == "goog" {
			if result.Variables.Meeting != "" {
				addItem(fmt.Sprintf("capture meeting notes for \"%v\"", result.Variables.Meeting), "e", true)
			}
			addItem("capture note into inbox", "g", captureCmdArgs.query != "")
			if result.Variables.ClockedInTask != "" {
				addItem("capture note for the clocked-in task", "c", captureCmdArgs.query != "")
			}
			if bs.Url != "" {
				addItem(fmt.Sprintf("capture \"%s\" into inbox", bs), "bg", true)
				addItem(fmt.Sprintf("capture \"%s\" for ads doc review", bs), "bd", true)
				addItem(fmt.Sprintf("capture \"%s\" for ads fact", bs), "bf", true)
				addItem(fmt.Sprintf("capture \"%s\" for career reading", bs), "bc", true)
			}
			addItem("capture ads fact", "f", true)
			if result.Variables.Meeting == "" {
				addItem("capture meeting notes for unknown meeting", "e", true)
			}
		} else {
			log.Fatalf("unknown category: %v", captureCmdArgs.category)
		}
		printJson(result)
	},
}

func captureItem(title, template string, valid bool) (item alfred.Item) {
	item.Title = title
	item.Arg = captureCmdArgs.query
	item.Valid = valid
	item.Variables.Action = "capture"
	item.Variables.Arg = template
	if template == "e" {
		item.Subtitle = "continue editing"
	} else {
		item.Subtitle = "finish immediately"
	}
	return item
}

type browserState struct {
	Url, Title string
}

func (b browserState) String() string {
	if b.Title != "" {
		return b.Title
	} else {
		return b.Url
	}
}

func initVariables(variables *alfred.Variables) {
	for _, varData := range []struct {
		name   string
		dest   *string
		initFn func() string
	}{
		{"browser_state", &variables.BrowserState, fetchBrowserState},
		{"meeting", &variables.Meeting, fetchMeeting},
		{"clocked_in_task", &variables.ClockedInTask, fetchClockedInTask},
	} {
		if val, exists := os.LookupEnv(varData.name); exists {
			*varData.dest = val
		} else {
			*varData.dest = varData.initFn()
		}
	}
}

// TODO: implement this
var captureActCmd = &cobra.Command{
	Use: "act",
}

func fetchBrowserState() (state string) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	if stdin, err := cmd.StdinPipe(); err != nil {
		log.Println("starting osascript failed:", err)
	} else {
		stdin.Write([]byte(`
			const frontmostAppName = Application("System Events").applicationProcesses.where({frontmost: true}).name()[0];
			const frontmostApp = Application(frontmostAppName);
			const chromiumVariants = ["Google Chrome", "Chromium"];
			const webkitVariants = ["Safari", "Webkit"];
			let tab = null;
			if (chromiumVariants.some(app_name => frontmostAppName.startsWith(app_name))) {
				tab = frontmostApp.windows[0].activeTab;
			} else if (webkitVariants.some(app_name => frontmostAppName.startsWith(app_name))) {
				tab = frontmostApp.documents[0];
			} else {
				throw new Error("You need a supported browser as your frontmost app");
			}
			JSON.stringify({url: tab.url(), title: tab.name()});
	  `))
		stdin.Close()
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Println("osascript failed:", err)
		} else {
			state = string(output)
		}
	}
	return state
}

func fetchMeeting() string {
	// TODO: implement this
	return ""
}

func fetchClockedInTask() (t string) {
	if out, err := exec.Command("emacsclient", "-e", "(org-clock-is-active)").Output(); err != nil {
		log.Println("calling emacsclient failed:", err)
	} else if !strings.HasPrefix(string(out), "nil") {
		t = "yes"
	}
	// TODO: perhaps return the name of the current task
	return t
}

var captureCmdArgs struct {
	category, query string
}

func init() {
	rootCmd.AddCommand(captureCmd)
	captureCmd.AddCommand(captureItemsCmd)
	captureItemsCmd.Flags().StringVarP(&captureCmdArgs.category, "category", "c", "", "Category of capture items")
	captureItemsCmd.MarkFlagRequired("category")
	captureItemsCmd.Flags().StringVarP(&captureCmdArgs.query, "query", "q", "", "Alfred query")
	captureItemsCmd.MarkFlagRequired("query")
	captureCmd.AddCommand(captureActCmd)
}
