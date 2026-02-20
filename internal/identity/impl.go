package identity

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i/pkg/identity"
	"github.com/cloudnative-pg/machinery/pkg/log"

	"github.com/sharifmshaker/cnpg-i-mcp-demo/pkg/metadata"
)

// Implementation is the implementation of the identity service
type Implementation struct {
	identity.IdentityServer
}

// GetPluginMetadata implements the IdentityServer interface
func (Implementation) GetPluginMetadata(
	context.Context,
	*identity.GetPluginMetadataRequest,
) (*identity.GetPluginMetadataResponse, error) {
	return &metadata.Data, nil
}

// GetPluginCapabilities implements the IdentityServer interface
func (Implementation) GetPluginCapabilities(
	ctx context.Context,
	_ *identity.GetPluginCapabilitiesRequest,
) (*identity.GetPluginCapabilitiesResponse, error) {
	logger := log.FromContext(ctx).WithName("cnpg_i_mcp_identity")
	logger.Info("=== GetPluginCapabilities called on Identity service ===")

	response := &identity.GetPluginCapabilitiesResponse{
		Capabilities: []*identity.PluginCapability{
			{
				Type: &identity.PluginCapability_Service_{
					Service: &identity.PluginCapability_Service{
						Type: identity.PluginCapability_Service_TYPE_LIFECYCLE_SERVICE,
					},
				},
			},
			{
				Type: &identity.PluginCapability_Service_{
					Service: &identity.PluginCapability_Service{
						Type: identity.PluginCapability_Service_TYPE_RECONCILER_HOOKS,
					},
				},
			},
		},
	}

	logger.Info("=== Returning Identity capabilities ===",
		"numCapabilities", len(response.Capabilities),
		"hasLifecycle", true,
		"hasReconcilerHooks", true)

	return response, nil
}

// Probe implements the IdentityServer interface
func (Implementation) Probe(context.Context, *identity.ProbeRequest) (*identity.ProbeResponse, error) {
	return &identity.ProbeResponse{
		Ready: true,
	}, nil
}
