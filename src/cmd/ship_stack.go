package cmd

import (
	"github.com/Originate/exit"
	"github.com/Originate/git-town/src/drivers"
	"github.com/Originate/git-town/src/git"
	"github.com/Originate/git-town/src/prompt"
	"github.com/Originate/git-town/src/script"
	"github.com/Originate/git-town/src/steps"
	"github.com/Originate/git-town/src/util"

	"github.com/spf13/cobra"
)

type shipAllConfig struct {
	BranchToMergeInto string
	BranchesToShip    []string
	InitialBranch     string
}

var shipAllCommitMessage string

var shipAllCmd = &cobra.Command{
	Use:   "ship-stack",
	Short: "Deliver a stack of feature branches",
	Long: `Deliver a stack of feature branches

Squash-merges the stack of feature branches up to the current branch (or <branch_name> if given)
into the main branch, resulting in linear history on the main branch.

- syncs the main branch
- pulls remote updates for <branch_name>
- merges the main branch into <branch_name>
- squash-merges <branch_name> into the main branch
  with commit message specified by the user
- pushes the main branch to the remote repository
- deletes <branch_name> from the local and remote repositories
- repeat until the stack has been merged

If you are using GitHub, this command can squash merge pull requests via the GitHub API. Setup:
1. Get a GitHub personal access token with the "repo" scope
2. Run 'git config git-town.github-token XXX' (optionally add the '--global' flag)
Now anytime you ship a branch with a pull request on GitHub, it will squash merge via the GitHub API.
It will also update the base branch for any pull requests against that branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := gitShipAllConfig(args)
		stepList := getShipAllStepList(config)
		runState := steps.NewRunState("ship", stepList)
		steps.Run(runState)
	},
	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return util.FirstError(
			git.ValidateIsRepository,
			validateIsConfigured,
		)
	},
}

func gitShipAllConfig(args []string) (result shipAllConfig) {
	result.InitialBranch = git.GetCurrentBranchName()
	var branchToShip string
	if len(args) == 0 {
		branchToShip = result.InitialBranch
	} else {
		branchToShip = args[0]
	}
	if branchToShip == result.InitialBranch {
		git.EnsureDoesNotHaveOpenChanges("Did you mean to commit them before shipping?")
	}
	if git.HasRemote("origin") && !git.IsOffline() {
		script.Fetch()
	}
	if branchToShip != result.InitialBranch {
		git.EnsureHasBranch(branchToShip)
	}
	git.EnsureIsFeatureBranch(branchToShip, "Only feature branches can be shipped.")
	prompt.EnsureKnowsParentBranches([]string{branchToShip})
	ancestors := git.GetAncestorBranches(branchToShip)
	result.BranchToMergeInto = ancestors[0]
	result.BranchesToShip = ancestors[1:]
	return
}

func getShipAllStepList(config shipAllConfig) steps.StepList {
	result := steps.StepList{}
	var isOffline = git.IsOffline()
	isShippingInitialBranch := config.BranchesToShip[len(config.BranchesToShip)-1] == config.InitialBranch
	branchToMergeInto := config.BranchToMergeInto
	for _, branchToShip := range config.BranchesToShip {
		result.AppendList(steps.GetSyncBranchSteps(branchToMergeInto, true))
		result.AppendList(steps.GetSyncBranchSteps(branchToShip, false))
		result.Append(&steps.EnsureHasShippableChangesStep{BranchName: branchToShip})
		result.Append(&steps.CheckoutBranchStep{BranchName: branchToMergeInto})
		canShipWithDriver, defaultCommitMessage := getCanShipAllWithDriver(branchToShip, branchToMergeInto)
		if canShipWithDriver {
			result.Append(&steps.PushBranchStep{BranchName: branchToShip})
			result.Append(&steps.DriverMergePullRequestStep{BranchName: branchToShip, CommitMessage: shipAllCommitMessage, DefaultCommitMessage: defaultCommitMessage})
			result.Append(&steps.PullBranchStep{})
		} else {
			result.Append(&steps.SquashMergeBranchStep{BranchName: branchToShip, CommitMessage: shipAllCommitMessage})
		}
		if git.HasRemote("origin") && !isOffline {
			result.Append(&steps.PushBranchStep{BranchName: branchToMergeInto, Undoable: true})
		}
		childBranches := git.GetChildBranches(branchToShip)
		if canShipWithDriver || (git.HasTrackingBranch(branchToShip) && len(childBranches) == 0 && !isOffline) {
			// I don't think I need this step. The Github API seems to be deleting remote branches.
			// result.Append(&steps.DeleteRemoteBranchStep{BranchName: branchToShip, IsTracking: true})
		}
		result.Append(&steps.DeleteLocalBranchStep{BranchName: branchToShip})
		result.Append(&steps.DeleteParentBranchStep{BranchName: branchToShip})
		for _, child := range childBranches {
			result.Append(&steps.SetParentBranchStep{BranchName: child, ParentBranchName: branchToMergeInto})
		}
	}
	if !isShippingInitialBranch {
		result.Append(&steps.CheckoutBranchStep{BranchName: config.InitialBranch})
	}
	result.Wrap(steps.WrapOptions{RunInGitRoot: true, StashOpenChanges: !isShippingInitialBranch})
	return result
}

func getCanShipAllWithDriver(branch, parentBranch string) (bool, string) {
	if !git.HasRemote("origin") {
		return false, ""
	}
	if git.IsOffline() {
		return false, ""
	}
	driver := drivers.GetActiveDriver()
	if driver == nil {
		return false, ""
	}
	canMerge, defaultCommitMessage, err := driver.CanMergePullRequest(branch, parentBranch)
	exit.If(err)
	return canMerge, defaultCommitMessage
}

func init() {
	shipAllCmd.Flags().StringVarP(&shipAllCommitMessage, "message", "m", "", "Specify the commit message for the squash commit")
	RootCmd.AddCommand(shipAllCmd)
}
