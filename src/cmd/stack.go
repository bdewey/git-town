package cmd

import (
	"fmt"

	"github.com/Originate/git-town/src/git"

	"github.com/spf13/cobra"
)

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Display the current working stack",
	Long: `Display the current working stack

Shows information about the current stack of changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, world")
	},
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return git.ValidateIsRepository()
	},
}

func init() {
	RootCmd.AddCommand(stackCmd)
}
