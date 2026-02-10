package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ServerConfFileName is the name of the server configuration file
	ServerConfFileName = "server.conf"
	
	// ServerNameLength is the length of the server name
	ServerNameLength = 6
)

// ServerIdentity represents the server's unique identity and configuration state.
// This struct holds persistent server-specific settings that are stored in server.conf.
type ServerIdentity struct {
	// Name is the unique identifier for this server instance.
	// Generated randomly on first start if not present.
	Name string `json:"name"`
	
	// Registry is the URL to the configuration package registry.
	// This registry maintains available configuration versions.
	Registry string `json:"registry"`
	
	// ConfigVersion is the currently active configuration version.
	// Defaults to the binary version (main.Version) if not specified.
	// Format: vX.Y.Z (e.g., "v0.0.1")
	ConfigVersion string `json:"config_version"`
}

// GenerateServerName generates a random 6-character server name
// consisting of uppercase, lowercase letters and numbers
func GenerateServerName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, ServerNameLength)
	rand.Read(b)
	
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	
	return string(b)
}

// GetOrCreateServerIdentity retrieves the server identity from server.conf
// or creates a new one if it doesn't exist.
//
// This method ensures that all required fields (name, registry, config_version)
// are present and populated with appropriate default values if missing.
// The config_version defaults to the binary version if not specified.
//
// Returns the server identity or an error if reading/writing fails.
func (c *Config) GetOrCreateServerIdentity() (*ServerIdentity, error) {
	confPath := filepath.Join(c.Storage.DataDir, ServerConfFileName)
	
	// Check if server.conf exists
	if _, err := os.Stat(confPath); err == nil {
		// File exists, read it
		identity, err := c.readServerIdentity(confPath)
		if err != nil {
			return nil, err
		}
		
		// Check for missing fields and add defaults
		needsUpdate := false
		
		if identity.Registry == "" {
			identity.Registry = DefaultRegistry
			needsUpdate = true
		}
		
		if identity.ConfigVersion == "" {
			identity.ConfigVersion = c.getBinaryVersion()
			needsUpdate = true
		}
		
		// Update file if any field was missing
		if needsUpdate {
			if err := c.writeServerIdentity(confPath, identity); err != nil {
				return nil, fmt.Errorf("failed to update server identity: %w", err)
			}
		}
		
		return identity, nil
	}
	
	// File doesn't exist, create new identity with all defaults
	identity := &ServerIdentity{
		Name:          GenerateServerName(),
		Registry:      DefaultRegistry,
		ConfigVersion: c.getBinaryVersion(),
	}
	
	// Write to file
	if err := c.writeServerIdentity(confPath, identity); err != nil {
		return nil, fmt.Errorf("failed to write server identity: %w", err)
	}
	
	return identity, nil
}

// getBinaryVersion returns the binary version for use as default config version.
// This is stored in the Config during initialization from main.Version.
func (c *Config) getBinaryVersion() string {
	if c.BinaryVersion != "" {
		return c.BinaryVersion
	}
	// Fallback to v0.0.1 if not set (should not happen in normal operation)
	return "v0.0.1"
}

// readServerIdentity reads server identity from file
func (c *Config) readServerIdentity(path string) (*ServerIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read server.conf: %w", err)
	}
	
	// Parse simple key=value format
	lines := strings.Split(string(data), "\n")
	identity := &ServerIdentity{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "name":
			identity.Name = value
		case "registry":
			identity.Registry = value
		case "config_version":
			identity.ConfigVersion = value
		}
	}
	
	if identity.Name == "" {
		return nil, fmt.Errorf("server.conf does not contain 'name' field")
	}
	
	return identity, nil
}

// writeServerIdentity writes server identity to file in key=value format.
// The file includes helpful comments for each configuration field.
func (c *Config) writeServerIdentity(path string, identity *ServerIdentity) error {
	content := fmt.Sprintf(`# XW Server Configuration
# Do not modify this file unless you know what you are doing

# Server instance unique identifier
name=%s

# Configuration package registry URL
registry=%s

# Configuration version currently in use
config_version=%s
`, identity.Name, identity.Registry, identity.ConfigVersion)
	
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadServerConfig loads server configuration from server.conf
func (c *Config) LoadServerConfig() error {
	identity, err := c.GetOrCreateServerIdentity()
	if err != nil {
		return err
	}
	
	c.Server.Name = identity.Name
	c.Server.Registry = identity.Registry
	return nil
}

// SaveServerConfig saves current server configuration to server.conf
func (c *Config) SaveServerConfig() error {
	confPath := filepath.Join(c.Storage.DataDir, ServerConfFileName)
	identity := &ServerIdentity{
		Name:     c.Server.Name,
		Registry: c.Server.Registry,
	}
	return c.writeServerIdentity(confPath, identity)
}

