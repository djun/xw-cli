// Package config provides configuration management for the xw CLI tool.
//
// This file defines extended sandbox configuration structures, which allow
// users to define device sandbox configurations for additional chip types
// without modifying the source code. This is particularly useful for supporting
// niche accelerator chips in a flexible and maintainable way.
//
// Extended sandbox configurations are defined within devices.yaml under each
// chip model's ext_sandboxes field.
package config

import "fmt"

// ExtSandboxesConfig contains both common configuration and engine-specific configs.
//
// This structure allows defining shared settings (devices, volumes, runtime) that apply
// to all engines, along with engine-specific configurations that define unique behavior.
//
// Example YAML:
//
//	ext_sandboxes:
//	  # Common configuration (shared by all engines)
//	  devices:
//	    - /dev/xpu0
//	    - /dev/xpu1
//	  volumes:
//	    - /usr/local/xpu:/usr/local/xpu
//	  runtime: runc
//	  # Engine-specific configurations
//	  vllm:
//	    device_env: XPU_VISIBLE_DEVICES
//	    privileged: true
//	  mindie:
//	    device_env: XPU_DEVICE_IDS
//	    privileged: true
type ExtSandboxesConfig struct {
	// Common configuration shared by all engines
	Devices []string `yaml:"devices,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
	Runtime string   `yaml:"runtime,omitempty"`
	
	// Engine-specific configurations  
	// Populated during YAML unmarshaling with all non-common keys
	Engines map[string]*ExtSandboxConfig `yaml:"-"`
}

// UnmarshalYAML implements custom YAML unmarshaling for ExtSandboxesConfig.
// It separates common fields (devices, volumes, runtime) from engine-specific configs.
func (c *ExtSandboxesConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First, unmarshal into a raw map
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	
	// Initialize engines map
	c.Engines = make(map[string]*ExtSandboxConfig)
	
	// Process each key
	for key, value := range raw {
		switch key {
		case "devices":
			if devices, ok := value.([]interface{}); ok {
				c.Devices = make([]string, len(devices))
				for i, d := range devices {
					if s, ok := d.(string); ok {
						c.Devices[i] = s
					}
				}
			}
		case "volumes":
			if volumes, ok := value.([]interface{}); ok {
				c.Volumes = make([]string, len(volumes))
				for i, v := range volumes {
					if s, ok := v.(string); ok {
						c.Volumes[i] = s
					}
				}
			}
		case "runtime":
			if s, ok := value.(string); ok {
				c.Runtime = s
			}
		default:
			// This is an engine configuration
			// Re-marshal and unmarshal into ExtSandboxConfig
			if engineMap, ok := value.(map[string]interface{}); ok {
				cfg := &ExtSandboxConfig{}
				// Convert map to YAML-compatible format and unmarshal
				if err := remarshalYAML(engineMap, cfg); err != nil {
					return fmt.Errorf("failed to parse engine %s: %w", key, err)
				}
				c.Engines[key] = cfg
			}
		}
	}
	
	return nil
}

// remarshalYAML is a helper to convert a map back to a struct
func remarshalYAML(from interface{}, to interface{}) error {
	// This is a simple approach - convert through map
	// In production, you might want to use a more robust method
	if m, ok := from.(map[string]interface{}); ok {
		if devEnv, ok := m["device_env"].(string); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.DeviceEnv = devEnv
			}
		}
		if env, ok := m["environment"].(map[string]interface{}); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Environment = make(map[string]string)
				for k, v := range env {
					if s, ok := v.(string); ok {
						cfg.Environment[k] = s
					}
				}
			}
		}
		if priv, ok := m["privileged"].(bool); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Privileged = priv
			}
		}
		if shm, ok := m["shm_size_gb"].(int); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.ShmSizeGB = shm
			}
		}
		if runtime, ok := m["runtime"].(string); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Runtime = runtime
			}
		}
		if caps, ok := m["capabilities"].([]interface{}); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Capabilities = make([]string, len(caps))
				for i, cap := range caps {
					if s, ok := cap.(string); ok {
						cfg.Capabilities[i] = s
					}
				}
			}
		}
		if devices, ok := m["devices"].([]interface{}); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Devices = make([]string, len(devices))
				for i, d := range devices {
					if s, ok := d.(string); ok {
						cfg.Devices[i] = s
					}
				}
			}
		}
		if volumes, ok := m["volumes"].([]interface{}); ok {
			if cfg, ok := to.(*ExtSandboxConfig); ok {
				cfg.Volumes = make([]string, len(volumes))
				for i, v := range volumes {
					if s, ok := v.(string); ok {
						cfg.Volumes[i] = s
					}
				}
			}
		}
	}
	return nil
}

// ExtSandboxConfig defines the configuration for an extended device sandbox.
//
// This configuration maps directly to the DeviceSandbox interface methods,
// providing a declarative way to define device-specific container requirements
// for AI accelerators.
//
// The configuration is designed to support common device patterns while
// maintaining simplicity. Complex scenarios (e.g., custom device initialization,
// dynamic configuration) should be implemented in code-based sandboxes.
//
// Configuration Structure:
//
// Extended sandboxes support chip-level common configuration to reduce duplication.
// Common settings (devices, volumes, runtime) are defined at chip model level and
// automatically merged with engine-specific configurations.
//
// Example configuration:
//
//	chip_models:
//	  - config_key: kunlun-r200
//	    ext_sandboxes:
//	      # Common configuration (shared by all engines)
//	      devices:
//	        - /dev/xpu0
//	        - /dev/xpu1
//	        - /dev/xpu_ctl
//	      volumes:
//	        - /usr/local/xpu:/usr/local/xpu
//	      runtime: runc
//	      # Engine-specific configurations (device_env, privileged, etc.)
//	      vllm:
//	        device_env: XPU_VISIBLE_DEVICES
//	        environment:
//	          XPU_WORKER_MULTIPROC_METHOD: spawn
//	        privileged: true
//	        shm_size_gb: 50
//	        capabilities:
//	          - SYS_ADMIN
//	        # Optional: engines can add extra volumes
//	        volumes:
//	          - /root/.cache:/root/.cache
//	      mindie:
//	        device_env: XPU_DEVICE_IDS
//	        privileged: true
//	        shm_size_gb: 100
//
//
// Merge Strategy:
//   - Devices: common.devices + engine.devices (appended)
//   - Volumes: common.volumes + engine.volumes (appended)
//   - Runtime: engine.runtime overrides common.runtime if specified
//   - Other fields: managed by engine config only
type ExtSandboxConfig struct {
	// DeviceEnv specifies the environment variable name for device visibility.
	// This variable will be automatically set to a comma-separated list of
	// device indices (e.g., "0,1,2") when the container is created.
	//
	// Common examples:
	//   - Ascend NPU: ASCEND_RT_VISIBLE_DEVICES
	//   - NVIDIA GPU: CUDA_VISIBLE_DEVICES
	//   - Kunlun XPU: XPU_VISIBLE_DEVICES
	//   - Cambricon MLU: MLU_VISIBLE_DEVICES
	DeviceEnv string `yaml:"device_env"`

	// Environment contains additional static environment variables to set
	// in the container. These are set as-is without any template processing.
	//
	// Example:
	//   environment:
	//     LOG_LEVEL: "3"
	//     WORKER_METHOD: spawn
	Environment map[string]string `yaml:"environment"`

	// Devices lists all device node paths that may need to be mounted.
	//
	// Device mounting behavior:
	//   - Paths ending with a digit (e.g., /dev/xpu0, /dev/xpu5):
	//     Mounted only when the corresponding device index is allocated.
	//   - Paths not ending with a digit (e.g., /dev/xpu_ctl):
	//     Mounted as shared devices for all containers.
	//
	// Example:
	//   devices:
	//     - /dev/xpu0          # Mounted only when device 0 is allocated
	//     - /dev/xpu1          # Mounted only when device 1 is allocated
	//     - /dev/xpu_ctl       # Always mounted (shared control device)
	Devices []string `yaml:"devices"`

	// Volumes defines host-to-container volume mounts in "host:container" format.
	// If no container path is specified (no colon), the same path is used for both.
	//
	// Example:
	//   volumes:
	//     - /usr/local/xpu:/usr/local/xpu
	//     - /root/.cache:/root/.cache
	Volumes []string `yaml:"volumes"`

	// Privileged indicates whether the container requires privileged mode.
	//
	// Privileged mode grants the container extended permissions, including
	// access to all devices and the ability to modify system settings.
	// While less secure, some accelerators require it for proper operation.
	Privileged bool `yaml:"privileged"`

	// Runtime specifies the Docker runtime to use (e.g., "runc", "nvidia").
	// Defaults to "runc" if not specified.
	Runtime string `yaml:"runtime"`

	// ShmSizeGB specifies the shared memory size in gigabytes.
	// If not specified or zero, defaults to 16GB.
	//
	// Shared memory is required for:
	//   - Inter-process communication in distributed inference
	//   - PyTorch DataLoader workers
	//   - Model tensor sharing
	ShmSizeGB int `yaml:"shm_size_gb,omitempty"`

	// Capabilities lists Linux capabilities required by the container.
	//
	// Common capabilities for AI accelerators:
	//   - SYS_ADMIN: System administration operations
	//   - SYS_RAWIO: Direct device I/O access
	//   - IPC_LOCK: Memory locking for device buffers
	//   - SYS_RESOURCE: Resource limit adjustments
	Capabilities []string `yaml:"capabilities"`
}

// LoadExtSandboxesFromDevices extracts extended sandbox configurations from device config.
//
// This function scans all loaded device configurations and collects extended sandbox
// definitions for the specified engine. It returns a map where keys are device
// config_keys and values are the corresponding sandbox configurations.
//
// The function expects device configuration to already be loaded via LoadDevicesConfig()
// or LoadDevicesConfigFrom().
//
// Parameters:
//   - engineName: Engine name to extract configs for ("vllm", "mindie", "mlguider")
//
// Returns:
//   - Map of device config_key -> ExtSandboxConfig for the specified engine
//   - Empty map if device config is not loaded or no extended sandboxes are defined
//
// Example usage:
//
//	// After loading device config
//	extConfigs := config.LoadExtSandboxesFromDevices("vllm")
//	for deviceType, cfg := range extConfigs {
//	    sandbox := runtime.NewExtSandbox(deviceType, cfg)
//	    // Register sandbox...
//	}
func LoadExtSandboxesFromDevices(engineName string) map[string]*ExtSandboxConfig {
	globalConfig, err := GetDevicesConfig()
	if err != nil || globalConfig == nil {
		return make(map[string]*ExtSandboxConfig)
	}

	result := make(map[string]*ExtSandboxConfig)

	// Scan all vendors and their chip models
	for _, vendor := range globalConfig.Vendors {
		for _, chipModel := range vendor.ChipModels {
			// Check if this chip model has extended sandbox configurations
			if chipModel.ExtSandboxes == nil {
				continue
			}

			// Check if there's a configuration for the specified engine
			if engineConfig, ok := chipModel.ExtSandboxes.Engines[engineName]; ok {
				// Merge common config with engine-specific config
				merged := mergeExtSandboxes(chipModel.ExtSandboxes, engineConfig)
				result[chipModel.ConfigKey] = merged
			}
		}
	}

	return result
}

// mergeExtSandboxes merges common sandbox configuration with engine-specific configuration.
//
// Merge strategy:
//   - Devices: common.devices + engine.devices (appended)
//   - Volumes: common.volumes + engine.volumes (appended)
//   - Runtime: engine.runtime overrides common.runtime if set
//   - Other fields (device_env, environment, privileged, shm_size_gb, capabilities):
//     Managed by engine-specific config only
//
// Parameters:
//   - common: ExtSandboxesConfig containing shared settings
//   - engine: Engine-specific configuration
//
// Returns:
//   - Merged configuration with common + engine-specific settings
func mergeExtSandboxes(common *ExtSandboxesConfig, engine *ExtSandboxConfig) *ExtSandboxConfig {
	// Create merged config starting with engine config
	merged := &ExtSandboxConfig{
		DeviceEnv:    engine.DeviceEnv,
		Environment:  engine.Environment,
		Privileged:   engine.Privileged,
		ShmSizeGB:    engine.ShmSizeGB,
		Capabilities: engine.Capabilities,
	}

	// Merge devices: common + engine-specific
	if len(common.Devices) > 0 {
		merged.Devices = append([]string{}, common.Devices...)
		merged.Devices = append(merged.Devices, engine.Devices...)
	} else {
		merged.Devices = engine.Devices
	}

	// Merge volumes: common + engine-specific
	if len(common.Volumes) > 0 {
		merged.Volumes = append([]string{}, common.Volumes...)
		merged.Volumes = append(merged.Volumes, engine.Volumes...)
	} else {
		merged.Volumes = engine.Volumes
	}

	// Runtime: Engine overrides common if set
	if engine.Runtime != "" {
		merged.Runtime = engine.Runtime
	} else if common.Runtime != "" {
		merged.Runtime = common.Runtime
	}

	return merged
}
