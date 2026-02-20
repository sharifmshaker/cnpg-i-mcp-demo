package reconciler

import (
	"context"
	"fmt"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/common"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/decoder"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/object"
	"github.com/cloudnative-pg/cnpg-i/pkg/reconciler"
	"github.com/cloudnative-pg/machinery/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sharifmshaker/cnpg-i-mcp-demo/internal/config"
	"github.com/sharifmshaker/cnpg-i-mcp-demo/pkg/metadata"
)

// Implementation is the implementation of the reconciler hooks
type Implementation struct {
	Client client.Client
	reconciler.UnimplementedReconcilerHooksServer
}

// GetCapabilities declares which reconcilers this plugin hooks into
func (impl Implementation) GetCapabilities(
	ctx context.Context,
	_ *reconciler.ReconcilerHooksCapabilitiesRequest,
) (*reconciler.ReconcilerHooksCapabilitiesResult, error) {
	result := &reconciler.ReconcilerHooksCapabilitiesResult{
		ReconcilerCapabilities: []*reconciler.ReconcilerHooksCapability{
			{
				Kind: reconciler.ReconcilerHooksCapability_KIND_CLUSTER,
			},
		},
	}
	return result, nil
}

// Pre is called before the operator executes the reconciliation loop
// This is where we create/update the MCP service
func (impl Implementation) Pre(
	ctx context.Context,
	request *reconciler.ReconcilerHooksRequest,
) (*reconciler.ReconcilerHooksResult, error) {
	logger := log.FromContext(ctx).WithName("cnpg_i_mcp_reconciler_pre")

	// Check what kind of resource is being reconciled
	reconciledKind, err := object.GetKind(request.GetResourceDefinition())
	if err != nil {
		return nil, err
	}

	// We only handle Cluster reconciliation
	if reconciledKind != "Cluster" {
		logger.Debug("Skipping non-Cluster resource", "kind", reconciledKind)
		return &reconciler.ReconcilerHooksResult{
			Behavior: reconciler.ReconcilerHooksResult_BEHAVIOR_CONTINUE,
		}, nil
	}

	logger.Info("Parsing cluster definition")
	cluster, err := decoder.DecodeClusterLenient(request.GetResourceDefinition())
	if err != nil {
		logger.Error(err, "Failed to decode cluster definition")
		return nil, err
	}

	logger.Info("Pre-reconciliation hook called", "cluster", cluster.Name, "namespace", cluster.Namespace)

	// Check if cluster is being deleted
	if !cluster.DeletionTimestamp.IsZero() {
		logger.Info("Cluster is being deleted, cleaning up MCP service", "cluster", cluster.Name)
		if err := deleteMCPService(ctx, impl.Client, cluster); err != nil {
			logger.Error(err, "Failed to delete MCP service")
			return nil, fmt.Errorf("failed to delete MCP service: %w", err)
		}
		logger.Info("MCP service deleted successfully", "cluster", cluster.Name)
		// Return BEHAVIOR_CONTINUE to allow deletion to proceed
		return &reconciler.ReconcilerHooksResult{
			Behavior: reconciler.ReconcilerHooksResult_BEHAVIOR_CONTINUE,
		}, nil
	}

	helper := common.NewPlugin(
		*cluster,
		metadata.PluginName,
	)

	configuration, valErrs := config.FromParameters(helper)
	if len(valErrs) > 0 {
		return nil, valErrs[0]
	}

	// Ensure dedicated MCP service exists for this cluster
	logger.Info("Ensuring MCP service exists", "cluster", cluster.Name, "namespace", cluster.Namespace, "replicaOnly", configuration.ReplicaOnly)
	if err := ensureMCPService(ctx, impl.Client, cluster, configuration.ReplicaOnly); err != nil {
		logger.Error(err, "Failed to ensure MCP service exists")
		// Return error to prevent reconciliation
		return nil, fmt.Errorf("failed to ensure MCP service exists: %w", err)
	}

	logger.Info("MCP service ensured successfully", "cluster", cluster.Name)

	// Return BEHAVIOR_CONTINUE - let the normal reconciliation proceed
	return &reconciler.ReconcilerHooksResult{
		Behavior: reconciler.ReconcilerHooksResult_BEHAVIOR_CONTINUE,
	}, nil
}

// Post is called after the operator executes the reconciliation loop
// We use this to verify the service is still correct
func (impl Implementation) Post(
	ctx context.Context,
	request *reconciler.ReconcilerHooksRequest,
) (*reconciler.ReconcilerHooksResult, error) {
	cluster, err := decoder.DecodeClusterLenient(request.GetClusterDefinition())
	if err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).WithName("cnpg_i_mcp_reconciler_post")
	logger.Debug("Post-reconciliation hook called", "cluster", cluster.Name, "namespace", cluster.Namespace)

	// Return BEHAVIOR_CONTINUE - reconciliation is complete
	return &reconciler.ReconcilerHooksResult{
		Behavior: reconciler.ReconcilerHooksResult_BEHAVIOR_CONTINUE,
	}, nil
}
