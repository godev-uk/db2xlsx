package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "db2xlsx",
	Short: "db2xlsx is a command line tool to export a MySQL database as an Excel file",
	Long:  "db2xlsx is a command line tool to export a MySQL database as an Excel file",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred when executing db2xlsx: '%s'\n", err)
		os.Exit(1)
	}
}
