package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RmiOptions holds options for the rmi command
type RmiOptions struct {
	*GlobalOptions

	// Alias is the instance alias to remove
	Alias string

	// Force forces removal even if instance is running
	Force bool
}

// NewRmiCommand creates the rmi command.
//
// The rmi command removes a stopped model instance.
//
// Usage:
//
//	xw rmi ALIAS [OPTIONS]
//
// Examples:
//
//	# Remove a stopped instance
//	xw rmi my-model
//
//	# Force remove a running instance
//	xw rmi test --force
//
// Parameters:
//   - globalOpts: Global options shared across commands
//
// Returns:
//   - A configured cobra.Command for removing instances
func NewRmiCommand(globalOpts *GlobalOptions) *cobra.Command {
	opts := &RmiOptions{
		GlobalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:   "rmi ALIAS",
		Short: "Remove a model instance",
		Long: `Remove a model instance by its alias.

The alias can be found using 'xw ps --all'. Removing an instance will:
  - Remove the backend container
  - Free up system resources
  - Permanently delete the instance

The instance must be stopped first (use 'xw stop' first), unless using --force.`,
		Example: `  # Remove a stopped instance
  xw rmi my-model

  # Force remove a running instance
  xw rmi test --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Alias = args[0]
			return runRmi(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false,
		"force remove even if instance is running")

	return cmd
}

// runRmi executes the rmi command logic
func runRmi(opts *RmiOptions) error {
	client := getClient(opts.GlobalOptions)

	// Remove the instance via server API (using alias)
	err := client.RemoveInstanceByAlias(opts.Alias, opts.Force)
	if err != nil {
		return fmt.Errorf("failed to remove instance: %w", err)
	}

	fmt.Printf("Removed instance: %s\n", opts.Alias)

	return nil
}

