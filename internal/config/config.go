package config

import (
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/common"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/validation"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
)

const (
	replicaOnlyParameter = "replica-only"
	mcpImageTagParameter = "mcp-image-tag"
)

// Configuration represents the plugin configuration parameters
type Configuration struct {
	ReplicaOnly bool
	MCPImageTag string
}

// FromParameters builds a plugin configuration from the configuration parameters
func FromParameters(
	helper *common.Plugin,
) (*Configuration, []*operator.ValidationError) {
	validationErrors := make([]*operator.ValidationError, 0)
	configuration := &Configuration{
		ReplicaOnly: false,
	}
	if ro := helper.Parameters[replicaOnlyParameter]; ro == "true" {
		configuration.ReplicaOnly = true
	} else if ro != "" && ro != "false" {
		validationErrors = append(
			validationErrors,
			validation.BuildErrorForParameter(helper, replicaOnlyParameter, "must be true or false if set"),
		)
	}
	if mt := helper.Parameters[mcpImageTagParameter]; mt != "" {
		configuration.MCPImageTag = mt
	} else {
		configuration.MCPImageTag = "latest"
	}
	return configuration, validationErrors
}

// ToParameters serialize the configuration to a map of plugin parameters
func (config *Configuration) ToParameters() (map[string]string, error) {
	result := make(map[string]string)
	if config.ReplicaOnly {
		result[replicaOnlyParameter] = "true"
	}
	if config.MCPImageTag != "" {
		result[mcpImageTagParameter] = config.MCPImageTag
	}
	return result, nil
}
