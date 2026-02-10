package client

import (
	"github.com/tsingmaoai/xw-cli/internal/config"
)

// ListVersionsResponse represents the response from the list versions endpoint.
type ListVersionsResponse struct {
	CurrentXwVersion     string           `json:"current_xw_version"`
	CurrentConfigVersion string           `json:"current_config_version"`
	CompatibleVersions   []config.Package `json:"compatible_versions"`
	IncompatibleVersions []config.Package `json:"incompatible_versions"`
	InstalledVersions    []string         `json:"installed_versions"`
}

// CurrentVersionResponse represents the current version information.
type CurrentVersionResponse struct {
	ConfigVersion string `json:"config_version"`
	Installed     bool   `json:"installed"`
}

// UpdateRequest represents the update request body.
type UpdateRequest struct {
	Version string `json:"version"`
}

// UpdateResponse represents the update response.
type UpdateResponse struct {
	Success         bool   `json:"success"`
	PreviousVersion string `json:"previous_version"`
	NewVersion      string `json:"new_version"`
	Downloaded      bool   `json:"downloaded"`
	Message         string `json:"message"`
	RestartRequired bool   `json:"restart_required"`
}

// ListVersions retrieves all available configuration versions from the server.
//
// Returns:
//   - ListVersionsResponse with version information
//   - error if the request fails
func (c *Client) ListVersions() (*ListVersionsResponse, error) {
	var resp ListVersionsResponse
	if err := c.doRequest("GET", "/api/update/list", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetCurrentVersion retrieves the currently active configuration version.
//
// Returns:
//   - CurrentVersionResponse with current version info
//   - error if the request fails
func (c *Client) GetCurrentVersion() (*CurrentVersionResponse, error) {
	var resp CurrentVersionResponse
	if err := c.doRequest("GET", "/api/update/current", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Update updates the configuration to the specified version.
// If version is empty, updates to the latest compatible version.
//
// Parameters:
//   - version: Target version string (e.g., "v0.0.2"), or empty for latest
//
// Returns:
//   - UpdateResponse with update results
//   - error if the update fails
func (c *Client) Update(version string) (*UpdateResponse, error) {
	req := UpdateRequest{
		Version: version,
	}

	var resp UpdateResponse
	if err := c.doRequest("POST", "/api/update", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

