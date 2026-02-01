// Package models - config_loader.go provides configuration-based model loading.
//
// This module bridges the configuration system with the model registry,
// converting YAML-based model definitions into ModelSpec instances.
package models

import (
	"fmt"
	
	"github.com/tsingmao/xw/internal/api"
	"github.com/tsingmao/xw/internal/config"
	"github.com/tsingmao/xw/internal/logger"
)

// LoadModelsFromConfig loads model specifications from the configuration file.
//
// This function reads the model configuration and converts it into ModelSpec
// format used by the model registry. It provides a bridge between the
// configuration system and the existing model registry logic.
//
// Parameters:
//   - configPath: Optional path to model configuration file (empty for default)
//
// Returns:
//   - Slice of ModelSpec instances
//   - Error if configuration cannot be loaded or is invalid
//
// Example:
//
//	specs, err := LoadModelsFromConfig("")
//	if err != nil {
//	    log.Fatalf("Failed to load models: %v", err)
//	}
//	registry := NewRegistry()
//	for _, spec := range specs {
//	    registry.RegisterModelSpec(&spec)
//	}
func LoadModelsFromConfig(configPath string) ([]ModelSpec, error) {
	// Load model configuration
	modConfig, err := config.LoadModelsConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load model configuration: %w", err)
	}
	
	// Convert configuration to ModelSpec format
	var specs []ModelSpec
	
	for _, model := range modConfig.Models {
		spec := ModelSpec{
			ID:              model.ModelID,
			SourceID:        model.Source.SourceID,
			DisplayName:     model.ModelName,
			Family:          model.Family,
			Version:         model.Version,
			Parameters:      model.Parameters,
			ContextLength:   model.ContextLength,
			RequiredVRAM:    model.RequiredVRAMGB,
			Tag:             model.Source.Tag,
		}
		
		// Convert supported devices
		for _, device := range model.SupportedDevices {
			spec.SupportedDevices = append(spec.SupportedDevices, api.DeviceType(device))
		}
		
		// Convert backends
		for _, backend := range model.Backends {
			backendOpt := BackendOption{
				Type: convertBackendType(backend.Type),
				Mode: convertDeploymentMode(backend.Mode),
			}
			spec.Backends = append(spec.Backends, backendOpt)
		}
		
		specs = append(specs, spec)
	}
	
	logger.Debug("Loaded %d model(s) from configuration", len(specs))
	return specs, nil
}

// convertBackendType converts config.BackendType to models.BackendType.
func convertBackendType(configBackend config.BackendType) BackendType {
	switch configBackend {
	case config.BackendVLLM:
		return BackendTypeVLLM
	case config.BackendMindIE:
		return BackendTypeMindIE
	case config.BackendMLGuider:
		return BackendTypeMLGuider
	default:
		logger.Warn("Unknown backend type: %s, defaulting to vllm", configBackend)
		return BackendTypeVLLM
	}
}

// convertDeploymentMode converts config.DeploymentMode to models.DeploymentMode.
func convertDeploymentMode(configMode config.DeploymentMode) DeploymentMode {
	switch configMode {
	case config.DeploymentModeDocker:
		return DeploymentModeDocker
	case config.DeploymentModeNative:
		return DeploymentModeNative
	default:
		logger.Warn("Unknown deployment mode: %s, defaulting to docker", configMode)
		return DeploymentModeDocker
	}
}

// LoadAndRegisterModelsFromConfig loads models from configuration and registers them.
//
// This function loads models from the configuration file and registers them
// with the global model registry. It should be called during application
// initialization to populate the registry with models from configuration.
//
// Parameters:
//   - configPath: Optional path to model configuration file (empty for default)
//
// Returns:
//   - Error if configuration loading fails
//
// Example:
//
//	if err := LoadAndRegisterModelsFromConfig(""); err != nil {
//	    log.Fatalf("Failed to load models: %v", err)
//	}
func LoadAndRegisterModelsFromConfig(configPath string) error {
	// Load model specs from configuration
	specs, err := LoadModelsFromConfig(configPath)
	if err != nil {
		return err
	}
	
	// Register all models with the global registry
	registeredCount := 0
	for i := range specs {
		RegisterModelSpec(&specs[i])
		registeredCount++
	}
	
	logger.Info("Loaded and registered %d model(s) from configuration", registeredCount)
	return nil
}

