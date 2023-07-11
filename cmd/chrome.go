/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var chromeCmd = &cobra.Command{
	Use:   "chrome [match]",
	Short: "Output chrome alfred items matching the argument",
	Long:  `Output chrome alfred items matching the argument`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("chrome called")
	},
}

func init() {
	rootCmd.AddCommand(chromeCmd)
}
