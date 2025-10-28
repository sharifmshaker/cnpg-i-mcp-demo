package metadata

import "github.com/cloudnative-pg/cnpg-i/pkg/identity"

// PluginName is the name of the plugin
const PluginName = "postgres-mcp-cnpg-plugin"

// Data is the metadata of this plugin
var Data = identity.GetPluginMetadataResponse{
	Name:          PluginName,
	Version:       "0.0.1",
	DisplayName:   "Plugin for MCP Server",
	ProjectUrl:    "https://github.com/sharifmshaker/cnpg-i-mcp-demo",
	RepositoryUrl: "https://github.com/sharifmshaker/cnpg-i-mcp-demo",
	License:       "Proprietary",
	LicenseUrl:    "https://github.com/sharifmshaker/cnpg-i-mcp-demo/LICENSE",
	Maturity:      "alpha",
}
