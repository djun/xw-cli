// Package config - model_config.go implements model configuration loading and management.
//
// This module provides a flexible configuration system for AI model definitions.
// Model configurations are loaded from YAML files, allowing easy addition of new
// models without code changes.
//
// The configuration system supports model metadata, hardware compatibility,
// runtime backends, and quantization strategies.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
	
	"github.com/tsingmao/xw/internal/logger"
)

// BackendType represents the inference engine type.
type BackendType string

const (
	BackendVLLM     BackendType = "vllm"     // vLLM (fast LLM inference)
	BackendMindIE   BackendType = "mindie"   // MindIE (Huawei Ascend)
	BackendMLGuider BackendType = "mlguider" // MLGuider
)

// DeploymentMode represents how a backend can be deployed.
type DeploymentMode string

const (
	DeploymentModeDocker DeploymentMode = "docker" // Docker container deployment
	DeploymentModeNative DeploymentMode = "native" // Native installation
)

// ModelSourceType represents where the model comes from.
type ModelSourceType string

const (
	SourceTypeHuggingFace ModelSourceType = "huggingface" // HuggingFace Hub
	SourceTypeModelScope  ModelSourceType = "modelscope"  // ModelScope (阿里巴巴)
	SourceTypeLocal       ModelSourceType = "local"       // Local filesystem
	SourceTypeGit         ModelSourceType = "git"         // Git repository
	SourceTypeHTTP        ModelSourceType = "http"        // HTTP/HTTPS URL
)

// ModelSource defines where and how to obtain a model.
type ModelSource struct {
	// SourceType specifies the source platform/method
	SourceType ModelSourceType `yaml:"source_type"`
	
	// SourceID is the identifier within the source platform
	// Examples: "Qwen/Qwen2-7B" (HuggingFace), "qwen/Qwen2-7B" (ModelScope)
	SourceID string `yaml:"source_id"`
	
	// Tag specifies version/variant (optional)
	// Examples: "main", "v1.0"
	Tag string `yaml:"tag,omitempty"`
}

// BackendConfig defines a backend option for running a model.
type BackendConfig struct {
	// Type is the backend engine type (vllm, mindie, mlguider)
	Type BackendType `yaml:"type"`
	
	// Mode is the deployment mode (docker or native)
	Mode DeploymentMode `yaml:"mode"`
}

// ModelConfig defines configuration for an AI model.
//
// This structure represents a model that can be deployed and served by xw.
type ModelConfig struct {
	// ModelID is the unique identifier for this model
	// Convention: lowercase, hyphen-separated (e.g., "qwen2-7b")
	ModelID string `yaml:"model_id"`
	
	// ModelName is the human-readable display name
	// Example: "Qwen2 7B"
	ModelName string `yaml:"model_name"`
	
	// Family groups related models (e.g., "qwen", "llama")
	Family string `yaml:"family,omitempty"`
	
	// Version is the model version string (optional)
	// Example: "2.0", "2.5"
	Version string `yaml:"version,omitempty"`
	
	// Source defines where to obtain the model
	Source ModelSource `yaml:"source"`
	
	// Parameters is the model size in billions of parameters
	// Example: 7.0 for 7B model
	Parameters float64 `yaml:"parameters,omitempty"`
	
	// ContextLength is the maximum context window size in tokens
	ContextLength int `yaml:"context_length,omitempty"`
	
	// RequiredVRAMGB is the minimum VRAM in GB needed to run this model
	RequiredVRAMGB int `yaml:"required_vram_gb,omitempty"`
	
	// SupportedDevices lists compatible device config keys
	// Must match config_key from device configuration
	// Example: ["ascend-910b", "ascend-310p"]
	SupportedDevices []string `yaml:"supported_devices"`
	
	// Backends lists available backend options in priority order
	// The first available backend will be used by default
	Backends []BackendConfig `yaml:"backends"`
}

// ModelsConfig is the root configuration structure for model definitions.
//
// This structure maps to the YAML configuration file and contains all
// model definitions available in the system.
type ModelsConfig struct {
	// Version specifies the configuration schema version
	// Used for compatibility checking and migration
	Version string `yaml:"version"`
	
	// Models contains all available model configurations
	Models []ModelConfig `yaml:"models"`
	
	// ModelGroups provides logical grouping of related models (optional)
	// Example: {"qwen2": ["qwen2-0.5b", "qwen2-7b", "qwen2-72b"]}
	ModelGroups map[string][]string `yaml:"model_groups,omitempty"`
}

// ModelConfigLoader handles loading and caching of model configurations.
//
// The loader implements singleton pattern with lazy initialization.
// Configurations are loaded once and cached for the lifetime of the application.
//
// Thread Safety: All methods are safe for concurrent use.
type ModelConfigLoader struct {
	mu     sync.RWMutex
	config *ModelsConfig
	loaded bool
}

var (
	// modelConfigLoader is the global singleton instance
	modelConfigLoader = &ModelConfigLoader{}
	
	// defaultModelConfigPath is the default location for model configuration
	defaultModelConfigPath = "/etc/xw/models.yaml"
)

// LoadModelsConfig loads model configuration from the specified file path.
//
// This method reads and parses the YAML configuration file, validating the
// structure and content. If no path is provided, it uses the default location.
//
// The configuration is cached after first load. Subsequent calls return the
// cached configuration without re-reading the file.
//
// Configuration File Location Priority:
//   1. Provided configPath parameter
//   2. XW_MODEL_CONFIG environment variable
//   3. Default: /etc/xw/models.yaml
//
// Parameters:
//   - configPath: Optional path to configuration file (empty string for default)
//
// Returns:
//   - Pointer to loaded ModelsConfig
//   - Error if file cannot be read, parsed, or validated
//
// Example:
//
//	config, err := LoadModelsConfig("")
//	if err != nil {
//	    log.Fatalf("Failed to load model config: %v", err)
//	}
//	for _, model := range config.Models {
//	    fmt.Printf("Loaded model: %s (%s)\n", model.ModelName, model.ModelID)
//	}
func LoadModelsConfig(configPath string) (*ModelsConfig, error) {
	modelConfigLoader.mu.Lock()
	defer modelConfigLoader.mu.Unlock()
	
	// Return cached config if already loaded
	if modelConfigLoader.loaded {
		logger.Debug("Using cached model configuration")
		return modelConfigLoader.config, nil
	}
	
	// Determine config file path
	path := configPath
	if path == "" {
		// Check environment variable
		if envPath := os.Getenv("XW_MODEL_CONFIG"); envPath != "" {
			path = envPath
			logger.Debug("Using model config from XW_MODEL_CONFIG: %s", path)
		} else {
			path = defaultModelConfigPath
			logger.Debug("Using default model config path: %s", path)
		}
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("model configuration file not found: %s", path)
	}
	
	// Read configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read model config file %s: %w", path, err)
	}
	
	// Parse YAML
	var config ModelsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse model config YAML: %w", err)
	}
	
	// Validate configuration
	if err := validateModelsConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid model configuration: %w", err)
	}
	
	// Cache the loaded configuration
	modelConfigLoader.config = &config
	modelConfigLoader.loaded = true
	
	logger.Info("Loaded model configuration: %d model(s)", len(config.Models))
	
	return &config, nil
}

// GetModelsConfig returns the cached model configuration.
//
// This method provides access to the previously loaded configuration without
// re-reading the file. If configuration hasn't been loaded yet, it loads it
// from the default location.
//
// Returns:
//   - Pointer to ModelsConfig
//   - Error if configuration not loaded and loading fails
func GetModelsConfig() (*ModelsConfig, error) {
	modelConfigLoader.mu.RLock()
	if modelConfigLoader.loaded {
		config := modelConfigLoader.config
		modelConfigLoader.mu.RUnlock()
		return config, nil
	}
	modelConfigLoader.mu.RUnlock()
	
	// Not loaded yet, load with default path
	return LoadModelsConfig("")
}

// ReloadModelsConfig forces a reload of the model configuration.
//
// This method clears the cache and re-reads the configuration file.
// Useful for applying configuration changes without restarting the application.
//
// Parameters:
//   - configPath: Optional path to configuration file (empty for default)
//
// Returns:
//   - Pointer to reloaded ModelsConfig
//   - Error if reload fails
func ReloadModelsConfig(configPath string) (*ModelsConfig, error) {
	modelConfigLoader.mu.Lock()
	modelConfigLoader.loaded = false
	modelConfigLoader.config = nil
	modelConfigLoader.mu.Unlock()
	
	logger.Info("Reloading model configuration")
	return LoadModelsConfig(configPath)
}

// validateModelsConfig performs validation on the loaded configuration.
//
// Validation checks:
//   - Version field is present
//   - At least one model is defined
//   - Each model has required fields
//   - No duplicate model IDs
//   - Runtime configs are valid
//   - Hardware requirements are reasonable
//
// Parameters:
//   - config: Configuration to validate
//
// Returns:
//   - nil if valid
//   - Error describing validation failure
func validateModelsConfig(config *ModelsConfig) error {
	if config.Version == "" {
		return fmt.Errorf("configuration version is required")
	}
	
	if len(config.Models) == 0 {
		return fmt.Errorf("at least one model must be defined")
	}
	
	// Track model IDs to detect duplicates
	modelIDs := make(map[string]bool)
	
	for i, model := range config.Models {
		if model.ModelID == "" {
			return fmt.Errorf("model[%d]: model_id is required", i)
		}
		
		if model.ModelName == "" {
			return fmt.Errorf("model %s: model_name is required", model.ModelID)
		}
		
		// Check for duplicate model IDs
		if modelIDs[model.ModelID] {
			return fmt.Errorf("duplicate model_id: %s", model.ModelID)
		}
		modelIDs[model.ModelID] = true
		
		// Validate source
		if model.Source.SourceType == "" {
			return fmt.Errorf("model %s: source.source_type is required", model.ModelID)
		}
		if model.Source.SourceID == "" {
			return fmt.Errorf("model %s: source.source_id is required", model.ModelID)
		}
		
		// Validate supported devices
		if len(model.SupportedDevices) == 0 {
			return fmt.Errorf("model %s: at least one supported device is required", model.ModelID)
		}
		
		// Validate backends
		if len(model.Backends) == 0 {
			return fmt.Errorf("model %s: at least one backend is required", model.ModelID)
		}
		
		for j, backend := range model.Backends {
			if backend.Type == "" {
				return fmt.Errorf("model %s, backend[%d]: type is required", model.ModelID, j)
			}
			if backend.Mode == "" {
				return fmt.Errorf("model %s, backend[%d]: mode is required", model.ModelID, j)
			}
		}
	}
	
	return nil
}

// FindModelByID searches for a model by its ID.
//
// This is the primary method for looking up model configuration.
//
// Parameters:
//   - config: ModelsConfig to search
//   - modelID: The model_id to find
//
// Returns:
//   - Pointer to ModelConfig if found
//   - nil if not found
func FindModelByID(config *ModelsConfig, modelID string) *ModelConfig {
	for i := range config.Models {
		if config.Models[i].ModelID == modelID {
			return &config.Models[i]
		}
	}
	return nil
}

// FindModelsByDeviceType returns all models compatible with a device type.
//
// This method is used to filter models based on hardware availability.
//
// Parameters:
//   - config: ModelsConfig to search
//   - deviceConfigKey: Device config key to filter by
//
// Returns:
//   - Slice of pointers to compatible ModelConfig objects
func FindModelsByDeviceType(config *ModelsConfig, deviceConfigKey string) []*ModelConfig {
	var models []*ModelConfig
	for i := range config.Models {
		for _, supportedDevice := range config.Models[i].SupportedDevices {
			if supportedDevice == deviceConfigKey {
				models = append(models, &config.Models[i])
				break
			}
		}
	}
	return models
}

// GetDefaultBackend returns the first (highest priority) backend for a model.
//
// Parameters:
//   - model: ModelConfig to query
//
// Returns:
//   - Pointer to default BackendConfig
//   - nil if no backends are configured (should not happen after validation)
func GetDefaultBackend(model *ModelConfig) *BackendConfig {
	if len(model.Backends) == 0 {
		return nil
	}
	return &model.Backends[0]
}

// FindBackendByType searches for a backend configuration by backend type.
//
// Parameters:
//   - model: ModelConfig to search
//   - backendType: BackendType to find
//
// Returns:
//   - Pointer to BackendConfig if found
//   - nil if not found
func FindBackendByType(model *ModelConfig, backendType BackendType) *BackendConfig {
	for i := range model.Backends {
		if model.Backends[i].Type == backendType {
			return &model.Backends[i]
		}
	}
	return nil
}

// GetAllModelIDs returns a list of all model IDs defined in the configuration.
//
// Useful for validation and displaying available models.
//
// Parameters:
//   - config: ModelsConfig to extract IDs from
//
// Returns:
//   - Slice of model ID strings
func GetAllModelIDs(config *ModelsConfig) []string {
	ids := make([]string, len(config.Models))
	for i, model := range config.Models {
		ids[i] = model.ModelID
	}
	return ids
}

// SaveModelsConfig writes a ModelsConfig to a YAML file.
//
// This method is primarily used for:
//   - Generating template configuration files
//   - Exporting modified configurations
//   - Testing and validation
//
// Parameters:
//   - config: Configuration to save
//   - path: File path to write to
//
// Returns:
//   - Error if file cannot be written
func SaveModelsConfig(config *ModelsConfig, path string) error {
	// Validate before saving
	if err := validateModelsConfig(config); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}
	
	// Marshal to YAML with nice formatting
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	
	logger.Info("Saved model configuration to %s", path)
	return nil
}

