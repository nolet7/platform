package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Model registry operations",
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List models",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Listing models...")
		fmt.Println("(API integration will be added)")
		return nil
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
}
