/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "org-roam-alfred-items",
	Short: "Output various nodes from the roam database as alfred items",
	Long:  "Output various nodes from the roam database as alfred items",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var dbPath, category string

func init() {
	u, _ := user.Current()
	rootCmd.PersistentFlags().StringVar(&dbPath, "db_path", filepath.Join(u.HomeDir, "org/.roam.db"), "Path to the org roam database")
	rootCmd.PersistentFlags().StringVarP(&category, "category", "c", "", "Category to limit items to")
}
