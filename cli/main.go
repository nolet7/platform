package main

import (
	"fmt"
	"os"
	"github.com/nolet7/platform/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
