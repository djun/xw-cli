package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ReloadOptions holds options for the reload command
type ReloadOptions struct {
	*GlobalOptions
}

// NewReloadCommand creates the reload command.
//
// The reload command triggers the server to reload all configuration files
// (devices.yaml, models.yaml, runtime_params.yaml) without restarting.
// This is useful after updating configuration versions.
//
// Usage:
//
//	xw reload
//
// Examples:
//
//	# Reload configuration after update
//	xw reload
//
// Parameters:
//   - globalOpts: Global options shared across commands
//
// Returns:
//   - A configured cobra.Command for reloading configuration
func NewReloadCommand(globalOpts *GlobalOptions) *cobra.Command {
	opts := &ReloadOptions{
		GlobalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload server configuration",
		Long: `Reload all configuration files without restarting the server.

This command triggers the server to reload:
  - devices.yaml: Device and runtime images configuration
  - models.yaml: Model definitions and deployment configurations
  - runtime_params.yaml: Runtime parameter templates

This is useful after updating configuration versions with 'xw update'.
The server will reload the configuration from the current version directory
without requiring a restart, allowing changes to take effect immediately.`,
		Example: `  # Reload configuration after update
  xw reload`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReload(opts)
		},
	}

	return cmd
}

// runReload executes the reload command logic.
//
// This function calls the server API to trigger configuration reload.
//
// Parameters:
//   - opts: Reload command options
//
// Returns:
//   - nil on success
//   - error if API call fails or server is unreachable
func runReload(opts *ReloadOptions) error {
	client := getClient(opts.GlobalOptions)

	if err := client.ReloadConfig(); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	fmt.Println("✓ Configuration reloaded successfully")

	return nil
}

