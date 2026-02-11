package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tsingmaoai/xw-cli/internal/config"
	"github.com/tsingmaoai/xw-cli/internal/logger"
)

// ListVersionsResponse represents the response for listing available configuration versions.
//
// This response includes current version information, compatible versions,
// and versions that require a newer xw binary.
type ListVersionsResponse struct {
	// CurrentXwVersion is the current xw binary version (e.g., "v0.0.1").
	CurrentXwVersion string `json:"current_xw_version"`

	// CurrentConfigVersion is the currently active configuration version.
	CurrentConfigVersion string `json:"current_config_version"`

	// CompatibleVersions lists all configuration versions compatible with the current binary.
	// These versions can be updated to without upgrading the xw binary.
	CompatibleVersions []config.Package `json:"compatible_versions"`

	// IncompatibleVersions lists versions that require a newer xw binary.
	IncompatibleVersions []config.Package `json:"incompatible_versions"`

	// InstalledVersions lists configuration versions already downloaded locally.
	InstalledVersions []string `json:"installed_versions"`
}

// UpdateRequest represents the request body for updating configuration version.
//
// If Version is empty, the system will update to the latest compatible version.
type UpdateRequest struct {
	// Version specifies the target version (e.g., "v0.0.2").
	// If empty, updates to the latest compatible version.
	Version string `json:"version"`
}

// UpdateResponse represents the response after a configuration update operation.
//
// This response indicates whether the update was successful and provides
// details about the version change.
type UpdateResponse struct {
	// Success indicates whether the update completed successfully.
	Success bool `json:"success"`

	// PreviousVersion is the configuration version before the update.
	PreviousVersion string `json:"previous_version"`

	// NewVersion is the configuration version after the update.
	NewVersion string `json:"new_version"`

	// Downloaded indicates whether the configuration package was downloaded.
	// False if it was already present locally.
	Downloaded bool `json:"downloaded"`

	// Message provides a human-readable status message.
	Message string `json:"message"`

	// RestartRequired indicates if the server needs to be restarted
	// for the new configuration to take effect.
	RestartRequired bool `json:"restart_required"`
}

// CurrentVersionResponse represents the current configuration version information.
//
// This response provides details about the active configuration version.
type CurrentVersionResponse struct {
	// ConfigVersion is the currently active configuration version.
	ConfigVersion string `json:"config_version"`

	// Installed indicates whether the configuration files are present locally.
	Installed bool `json:"installed"`
}

// ListVersions handles GET /api/update/list requests.
//
// This endpoint returns all available configuration versions from the registry,
// categorized by compatibility with the current xw binary version.
//
// HTTP Method: GET
// Path: /api/update/list
//
// Response: 200 OK
//
//	{
//	  "current_xw_version": "v0.0.1",
//	  "current_config_version": "v0.0.1",
//	  "compatible_versions": [
//	    {
//	      "version": "v0.0.1",
//	      "name": "v0.0.1",
//	      "description": "Initial release",
//	      "release_date": "2026-02-10",
//	      "download_url": "https://xw.tsingmao.com/packages/v0.0.1.tar.gz",
//	      "sha256": "abc123...",
//	      "min_xw_version": "0.0.1"
//	    }
//	  ],
//	  "incompatible_versions": [],
//	  "installed_versions": ["v0.0.1"]
//	}
//
// Error Responses:
//   - 405 Method Not Allowed: Request method is not GET
//   - 500 Internal Server Error: Failed to fetch or parse registry
//
// Example:
//
//	curl http://localhost:11581/api/update/list
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.WriteError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create version manager
	vm := config.NewVersionManager(h.config)

	// Get current versions (always available locally)
	currentConfig, _ := vm.GetCurrentVersion()
	binaryVersion := h.config.BinaryVersion

	// Get locally installed versions (always available)
	installed, err := vm.ListInstalledVersions()
	if err != nil {
		logger.Warn("Failed to list installed versions: %v", err)
		installed = []string{}
	}

	// Try to fetch registry (may fail if offline or registry unavailable)
	logger.Debug("Fetching package registry...")
	registry, err := vm.FetchRegistry()
	if err != nil {
		logger.Warn("Failed to fetch registry: %v", err)
		
		// Return local information only
		response := ListVersionsResponse{
			CurrentXwVersion:     binaryVersion,
			CurrentConfigVersion: currentConfig,
			CompatibleVersions:   []config.Package{},
			IncompatibleVersions: []config.Package{},
			InstalledVersions:    installed,
		}
		
		h.WriteJSON(w, response, http.StatusOK)
		return
	}

	// Registry available, get compatible versions
	compatible, err := vm.GetCompatibleVersions(binaryVersion)
	if err != nil {
		logger.Error("Failed to get compatible versions: %v", err)
		h.WriteError(w, fmt.Sprintf("failed to get compatible versions: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Find incompatible versions
	compatibleMap := make(map[string]bool)
	for _, pkg := range compatible {
		compatibleMap[pkg.Version] = true
	}

	var incompatible []config.Package
	for _, pkg := range registry.Packages {
		if !compatibleMap[pkg.Version] {
			incompatible = append(incompatible, pkg)
		}
	}

	response := ListVersionsResponse{
		CurrentXwVersion:     binaryVersion,
		CurrentConfigVersion: currentConfig,
		CompatibleVersions:   compatible,
		IncompatibleVersions: incompatible,
		InstalledVersions:    installed,
	}

	h.WriteJSON(w, response, http.StatusOK)
}

// GetCurrentVersion handles GET /api/update/current requests.
//
// This endpoint returns information about the currently active configuration version.
//
// HTTP Method: GET
// Path: /api/update/current
//
// Response: 200 OK
//
//	{
//	  "config_version": "v0.0.1",
//	  "installed": true
//	}
//
// Error Responses:
//   - 405 Method Not Allowed: Request method is not GET
//   - 500 Internal Server Error: Failed to retrieve current version
//
// Example:
//
//	curl http://localhost:11581/api/update/current
func (h *Handler) GetCurrentVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.WriteError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vm := config.NewVersionManager(h.config)

	current, err := vm.GetCurrentVersion()
	if err != nil {
		h.WriteError(w, fmt.Sprintf("failed to get current version: %v", err),
			http.StatusInternalServerError)
		return
	}

	response := CurrentVersionResponse{
		ConfigVersion: current,
		Installed:     vm.IsVersionInstalled(current),
	}

	h.WriteJSON(w, response, http.StatusOK)
}

// Update handles POST /api/update requests.
//
// This endpoint updates the configuration to a specific version or the latest
// compatible version. The update process includes downloading the configuration
// package if needed, verifying checksums, and switching the active version.
//
// HTTP Method: POST
// Path: /api/update
//
// Request Body:
//
//	{
//	  "version": "v0.0.2"  // Optional: specific version, or empty for latest
//	}
//
// Response: 200 OK
//
//	{
//	  "success": true,
//	  "previous_version": "v0.0.1",
//	  "new_version": "v0.0.2",
//	  "downloaded": true,
//	  "message": "Updated to v0.0.2",
//	  "restart_required": true
//	}
//
// Error Responses:
//   - 400 Bad Request: Invalid request body or incompatible version
//   - 404 Not Found: Requested version does not exist
//   - 405 Method Not Allowed: Request method is not POST
//   - 500 Internal Server Error: Failed to download or switch version
//
// Example:
//
//	# Update to latest compatible version
//	curl -X POST http://localhost:11581/api/update -d '{}'
//
//	# Update to specific version
//	curl -X POST http://localhost:11581/api/update -d '{"version":"v0.0.2"}'
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.WriteError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	vm := config.NewVersionManager(h.config)

	// Get current version
	currentVersion, err := vm.GetCurrentVersion()
	if err != nil {
		h.WriteError(w, fmt.Sprintf("failed to get current version: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Fetch registry
	logger.Info("Fetching package registry...")
	_, err = vm.FetchRegistry()
	if err != nil {
		h.WriteError(w, fmt.Sprintf("failed to fetch registry: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Determine target version
	var targetVersion string
	var pkg *config.Package

	if req.Version == "" {
		// Update to latest compatible version
		logger.Info("Finding latest compatible version...")
		latest, err := vm.GetLatestCompatibleVersion(h.config.BinaryVersion)
		if err != nil {
			h.WriteError(w, fmt.Sprintf("no compatible versions found: %v", err),
				http.StatusBadRequest)
			return
		}
		targetVersion = latest.Version
		pkg = latest
	} else {
		// Update to specific version
		logger.Info("Finding version %s...", req.Version)
		foundPkg, err := vm.FindPackage(req.Version)
		if err != nil {
			h.WriteError(w, fmt.Sprintf("version not found: %v", err),
				http.StatusNotFound)
			return
		}

		// Check compatibility
		compatible, err := vm.IsVersionCompatible(foundPkg, h.config.BinaryVersion)
		if err != nil {
			h.WriteError(w, fmt.Sprintf("failed to check compatibility: %v", err),
				http.StatusInternalServerError)
			return
		}

		if !compatible {
			errMsg := fmt.Sprintf("configuration %s requires xw >= %s (current: %s)",
				foundPkg.Version, foundPkg.MinXwVersion, h.config.BinaryVersion)
			h.WriteError(w, errMsg, http.StatusBadRequest)
			return
		}

		targetVersion = req.Version
		pkg = foundPkg
	}

	// Check if already at target version (normalize versions by removing "v" prefix)
	normalizedCurrent := strings.TrimPrefix(currentVersion, "v")
	normalizedTarget := strings.TrimPrefix(targetVersion, "v")
	if normalizedCurrent == normalizedTarget && vm.IsVersionInstalled(targetVersion) {
		response := UpdateResponse{
			Success:         true,
			PreviousVersion: currentVersion,
			NewVersion:      targetVersion,
			Downloaded:      false,
			Message:         fmt.Sprintf("already at version %s", targetVersion),
			RestartRequired: false,
		}

		h.WriteJSON(w, response, http.StatusOK)
		return
	}

	// Download if not installed
	downloaded := false
	if !vm.IsVersionInstalled(targetVersion) {
		logger.Info("Downloading configuration %s...", targetVersion)
		if err := vm.DownloadPackage(pkg); err != nil {
			h.WriteError(w, fmt.Sprintf("failed to download package: %v", err),
				http.StatusInternalServerError)
			return
		}
		downloaded = true
	}

	// Switch version
	if err := vm.SwitchVersion(targetVersion); err != nil {
		h.WriteError(w, fmt.Sprintf("failed to switch version: %v", err),
			http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully updated from %s to %s", currentVersion, targetVersion)

	response := UpdateResponse{
		Success:         true,
		PreviousVersion: currentVersion,
		NewVersion:      targetVersion,
		Downloaded:      downloaded,
		Message:         fmt.Sprintf("updated to %s", targetVersion),
		RestartRequired: true,
	}

	h.WriteJSON(w, response, http.StatusOK)
}
