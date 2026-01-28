// Package qwen provides Qwen model family specifications.
package qwen

import (
	"github.com/tsingmao/xw/internal/api"
	"github.com/tsingmao/xw/internal/models"
)

// Qwen2_Instruct_7B is the instruction-tuned version of Qwen2 7B
//
// This variant has been specifically fine-tuned to follow instructions
// and provide helpful, harmless, and honest responses. Recommended for
// chat and assistant applications.
var Qwen2_Instruct_7B = &models.ModelSpec{
	ID:              "qwen2.5-7b-instruct",
	SourceID:        "Qwen/Qwen2.5-7B-Instruct",
	DisplayName:     "Qwen2 7B Instruct",
	Family:          "qwen",
	Version:         "2.5",
	Description:     "Instruction-tuned Qwen2 7B for chat and assistant tasks",
	Parameters:      7.0,
	RequiredVRAM:    16,
	ContextLength:   32768,
	EmbeddingLength: 3584,
	Capabilities:    []string{"completion"},
	License:         "Apache-2.0",
	Homepage:        "https://github.com/QwenLM/Qwen2",
	Tag:             "", // Default full precision variant

	SupportedDevices: []api.DeviceType{
		api.DeviceTypeAscend,
		api.DeviceTypeKunlun,
	},

	// Backend configuration - only vLLM Docker is supported
	Backends: []models.BackendOption{
		{Type: models.BackendTypeVLLM, Mode: models.DeploymentModeDocker},
	},
}

func init() {
	models.RegisterModelSpec(Qwen2_Instruct_7B)
}

