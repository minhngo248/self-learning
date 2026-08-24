package cmd

import (
	"os"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/cli/add"
	"github.com/minhngo248/self-learning/to-do-app/go/internal/cli/list"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "tasks"}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(add.NewCommand())
}
