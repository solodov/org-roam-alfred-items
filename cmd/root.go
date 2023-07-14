/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Output various nodes from the roam database as alfred items",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmdArgs struct {
	dbPath string
	pretty bool
}

func printJson(data any) {
	var (
		result []byte
		err    error
	)
	if rootCmdArgs.pretty {
		result, err = json.MarshalIndent(data, "", " ")
	} else {
		result, err = json.Marshal(data)
	}
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(result))
	}
}

func init() {
	u, _ := user.Current()
	rootCmd.PersistentFlags().StringVar(&rootCmdArgs.dbPath, "db_path", filepath.Join(u.HomeDir, "org/.roam.db"), "Path to the org roam database")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdArgs.pretty, "pretty", "p", false, "Pretty-print output")
}
