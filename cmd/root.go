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
	Short: "A collection of various Alfred tools",
}

var rootCmdArgs struct {
	pretty bool
}

var roamCmd = &cobra.Command{
	Use:   "roam",
	Short: "Output various nodes from the roam database as Alfred items",
}

var roamCmdArgs struct {
	dbPath string
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
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
	rootCmd.PersistentFlags().BoolVarP(&rootCmdArgs.pretty, "pretty", "p", false, "Pretty-print output")
	rootCmd.AddCommand(roamCmd)
	roamCmd.PersistentFlags().StringVar(&roamCmdArgs.dbPath, "db_path", filepath.Join(u.HomeDir, "org/.roam.db"), "Path to the org roam database")
}
