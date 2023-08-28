/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
		browserState := result.Variables.DecodeBrowserState()
		addItem := func(title, template string, valid bool) {
			result.Items = append(result.Items, captureItem(title, template, valid))
		}
		if captureCmdArgs.category == "home" {
			if result.Variables.Meeting != "nil" {
				addItem(fmt.Sprintf("capture meeting notes for %q", result.Variables.Meeting), "e", true)
			}
			addItem("capture note into inbox", "h", captureCmdArgs.query != "")
			if result.Variables.ClockedInTask != "nil" {
				addItem("capture note for the clocked-in task", "c", captureCmdArgs.query != "")
			}
			if browserState != nil {
				addItem(fmt.Sprintf("capture %q into inbox", browserState), "bh", true)
			}
		} else if captureCmdArgs.category == "goog" {
			if result.Variables.Meeting != "nil" {
				addItem(fmt.Sprintf("capture meeting notes for %q", result.Variables.Meeting), "e", true)
			}
			addItem("capture note into inbox", "g", captureCmdArgs.query != "")
			if result.Variables.ClockedInTask != "nil" {
				addItem("capture note for the clocked-in task", "c", captureCmdArgs.query != "")
			}
			if browserState != nil {
				addItem(fmt.Sprintf("capture %q into inbox", browserState), "bg", true)
				addItem(fmt.Sprintf("capture %q for ads doc review", browserState), "bd", true)
				addItem(fmt.Sprintf("capture %q for ads fact", browserState), "bf", true)
				addItem(fmt.Sprintf("capture %q for career reading", browserState), "bc", true)
			}
			addItem("capture ads fact", "f", true)
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
	item.Variables.Arg = template
	if template == "e" {
		item.Subtitle = "continue editing"
	} else {
		item.Subtitle = "finish immediately"
	}
	return item
}

var captureActCmd = &cobra.Command{
	Use:  "act",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		variables := alfred.Variables{}
		initVariables(&variables)
		browserState := variables.DecodeBrowserState()

		template := os.Getenv("arg")
		if template == "ie" {
			// e is for meetings, immediate finish (the i prefix) doesn't apply,
			// always edit meeting notes
			template = "e"
		} else if strings.HasPrefix(template, "ib") {
			// immediate finish for browser capture has its own series of templates,
			// alfred just adds i for simplicity.
			template = "y" + strings.TrimPrefix(template, "ib")
		}

		q := url.Values{}
		q.Set("template", template)
		if captureCmdArgs.query != "" {
			switch template {
			case "h", "ih", "g", "ig", "e", "f":
				q.Set("body", captureCmdArgs.query)
			default:
				// Immediate finish means add some empty lines.
				q.Set("body", captureCmdArgs.query+"\n\n")
			}
		}
		if template[0] == 'b' || template[0] == 'y' {
			if browserState == nil {
				log.Fatal("capture template requires browser state, but it's not provided")
			}
			// These are browser capture templates, add URL and title.
			q.Set("url", browserState.Url)
			q.Set("title", browserState.Title)
		}

		u := url.URL{Scheme: "org-protocol", Host: "capture", RawQuery: q.Encode()}

		log.Printf("browser state: %#v\n", browserState)
		log.Printf("arg: %#v\n", template)
		log.Printf("query: %#v\n", captureCmdArgs.query)
		log.Printf("url: %s\n", u.String())

		if template[0] != 'i' && template[0] != 'y' {
			// This is not an immediate finish template, raise emacs frame so
			// continuing to edit is nicer.
			if err := exec.Command("emacsclient", "-e", "(select-frame-set-input-focus (selected-frame))").Run(); err != nil {
				log.Fatal("setting frame focus failed: ", err)
			}
		}
		if err := exec.Command("emacsclient", "-n", u.String()).Run(); err != nil {
			log.Fatal("opening url failed: ", err)
		}
	},
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
	state = "nil"
	cmd := exec.Command("osascript", "-l", "JavaScript")
	if stdin, err := cmd.StdinPipe(); err != nil {
		log.Println("starting osascript failed: ", err)
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
			log.Println("osascript failed: ", err)
		} else {
			state = string(output)
		}
	}
	return state
}

func fetchMeeting() (meeting string) {
	meeting = "nil"
	cmd := exec.Command("lce", "now")
	if output, err := cmd.Output(); err != nil {
		log.Println("lce failed: ", err)
	} else if s := strings.Trim(string(output), "\n"); s != "" {
		// This is a pretty naive implementation, it doesn't check calendar name and
		// it doesn't care if there are multiple concurrent events. It's probably a
		// safe assumption for now as I'm rarely double-booked.
		meeting = strings.Split(s, "\n")[0]
	}
	return meeting
}

func fetchClockedInTask() (t string) {
	t = "nil"
	if out, err := exec.Command("emacsclient", "-e", "(org-clock-is-active)").Output(); err != nil {
		log.Println("calling emacsclient failed: ", err)
	} else {
		t = strings.Trim(string(out), "\n")
	}
	return t
}

var captureCmdArgs struct {
	category, query, orgDir string
}

func init() {
	rootCmd.AddCommand(captureCmd)
	captureCmd.AddCommand(captureItemsCmd)
	u, _ := user.Current()
	captureCmd.PersistentFlags().StringVar(&captureCmdArgs.orgDir, "org_dir", filepath.Join(u.HomeDir, "org"), "Path to the base org directory")
	captureItemsCmd.Flags().StringVarP(&captureCmdArgs.category, "category", "c", "", "Category of capture items")
	captureItemsCmd.MarkFlagRequired("category")
	captureItemsCmd.Flags().StringVarP(&captureCmdArgs.query, "query", "q", "", "Alfred query")
	captureItemsCmd.MarkFlagRequired("query")
	captureActCmd.Flags().StringVarP(&captureCmdArgs.query, "query", "q", "", "Alfred query")
	captureCmd.AddCommand(captureActCmd)
}
