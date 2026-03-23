package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "platform",
	Short: "Enterprise Platform CLI",
	Long:  "Command-line interface for the internal developer platform",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(catalogCmd)
	rootCmd.AddCommand(modelCmd)
}
