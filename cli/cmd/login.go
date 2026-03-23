package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the platform",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Ì¥ê Login flow will be implemented with Keycloak device code flow")
		fmt.Println("For now, this is a placeholder")
		return nil
	},
}
