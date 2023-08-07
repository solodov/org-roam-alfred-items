/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Perform org capture",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result := alfred.Result{}
		initVariables(&result.Variables)
		bs := browserState{}
		if result.Variables.BrowserState != "" {
			json.Unmarshal([]byte(result.Variables.BrowserState), &bs)
		}
		printJson(result)
	},
}

type browserState struct {
	Url, Title string
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

func fetchBrowserState() (state string) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	if stdin, err := cmd.StdinPipe(); err != nil {
		fmt.Fprintf(os.Stderr, "starting osascript failed: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "osascript failed: %v", err)
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
		fmt.Fprintf(os.Stderr, "calling emacsclient failed: %v\n", err)
	} else if !strings.HasPrefix(string(out), "nil") {
		t = "yes"
	}
	// TODO: perhaps return the name of the current task
	return t
}

var captureCmdArgs struct {
	category string
}

func init() {
	rootCmd.AddCommand(captureCmd)
	captureCmd.Flags().StringVarP(&captureCmdArgs.category, "category", "c", "", "Category of capture items")
	captureCmd.MarkFlagRequired("category")
}
