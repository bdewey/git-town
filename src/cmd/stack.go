package cmd

import (
	"github.com/Originate/git-town/src/cfmt"
	"github.com/Originate/git-town/src/git"

	"github.com/spf13/cobra"
)

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Display the current working stack",
	Long: `Display the current working stack

Shows information about the current stack of changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		printStackInfo()
	},
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return git.ValidateIsRepository()
	},
}

func printStackInfo() {
	currentBranch := git.GetCurrentBranchName()
	for _, branch := range git.GetAncestorBranches(currentBranch) {
		cfmt.Println(branch)
	}
	currentChildren := []string{currentBranch}
	for ; len(currentChildren) == 1; currentChildren = git.GetChildBranches(currentChildren[0]) {
		if currentChildren[0] == currentBranch {
			cfmt.Print("* ")
		}
		cfmt.Println(currentChildren[0])
	}
}

func init() {
	RootCmd.AddCommand(stackCmd)
}
