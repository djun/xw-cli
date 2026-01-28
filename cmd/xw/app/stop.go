package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// StopOptions holds options for the stop command
type StopOptions struct {
	*GlobalOptions

	// Alias is the instance alias to stop
	Alias string

	// Force forces stop even if instance is in use
	Force bool
}

// NewStopCommand creates the stop command.
//
// The stop command stops a running model instance.
//
// Usage:
//
//	xw stop ALIAS [OPTIONS]
//
// Examples:
//
//	# Stop an instance
//	xw stop my-model
//
//	# Force stop
//	xw stop test --force
//
// Parameters:
//   - globalOpts: Global options shared across commands
//
// Returns:
//   - A configured cobra.Command for stopping instances
func NewStopCommand(globalOpts *GlobalOptions) *cobra.Command {
	opts := &StopOptions{
		GlobalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:   "stop ALIAS",
		Short: "Stop a running model instance",
		Long: `Stop a running model instance by its alias.

The alias can be found using 'xw ps'. Stopping an instance will:
  - Stop the backend process/container
  - Keep the instance configuration
  - The instance can be viewed with 'xw ps --all'

To completely remove the instance, use 'xw rmi ALIAS' after stopping.
Use --force to stop an instance even if it's currently processing requests.`,
		Example: `  # Stop an instance
  xw stop my-model

  # Force stop
  xw stop test --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Alias = args[0]
			return runStop(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false,
		"force stop even if instance is in use")

	return cmd
}

// runStop executes the stop command logic
func runStop(opts *StopOptions) error {
	client := getClient(opts.GlobalOptions)

	// Stop the instance via server API (using alias)
	err := client.StopInstanceByAlias(opts.Alias, opts.Force)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	fmt.Printf("Stopped instance: %s\n", opts.Alias)

	return nil
}

