/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var elfeedCmd = &cobra.Command{
	Use:                   "elfeed",
	DisableFlagsInUseLine: true,
	Short:                 "Output elfeed alfred items",
	Long:                  "Output elfeed alfred items",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("elfeed called")
	},
}

func init() {
	rootCmd.AddCommand(elfeedCmd)
}
