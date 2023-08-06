/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Perform org capture",
	Run: func(cmd *cobra.Command, args []string) {
		data, _ := browserState()
		fmt.Fprintln(os.Stderr, data)
	},
}

func initVariables() (variables alfred.Variables) {
	if _, exists := os.LookupEnv("browser_state"); exists {
	}
	if val, exists := os.LookupEnv("meeting"); exists {
		variables.Meeting = val
	}
	return variables
}

func browserState() (data alfred.BrowserState, err error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	if stdin, e := cmd.StdinPipe(); e != nil {
		err = fmt.Errorf("starting osascript failed: %v\n", e)
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
		if output, e := cmd.CombinedOutput(); e != nil {
			err = fmt.Errorf("osascript failed: %v\n", e)
		} else if e := json.Unmarshal(output, &data); e != nil {
			err = fmt.Errorf("decoding failed: %v\n", e)
		}
	}
	return data, err
}

func init() {
	rootCmd.AddCommand(captureCmd)
}
