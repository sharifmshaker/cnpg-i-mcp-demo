package reconciler

import (
	"context"
	"fmt"

	apiv1 "github.com/cloudnative-pg/api/pkg/api/v1"
	"github.com/cloudnative-pg/machinery/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ensureMCPService cr`eates or updates a dedicated MCP service for the cluster
func ensureMCPService(ctx context.Context, cli client.Client, cluster *apiv1.Cluster, replicaOnly bool) error {
	logger := log.FromContext(ctx).WithName("mcp_service_manager")

	serviceName := fmt.Sprintf("%s-mcp", cluster.Name)
	namespace := cluster.Namespace

	// Define the service selector
	selector := map[string]string{
		"cnpg.io/cluster": cluster.Name,
	}

	// If replica-only mode, only target replica instances
	if replicaOnly {
		selector["cnpg.io/instanceRole"] = "replica"
	}

	// Define the desired service
	desiredService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"cnpg.io/cluster":   cluster.Name,
				"app":               "postgres-mcp",
				"managed-by":        "postgres-mcp-plugin",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: cluster.APIVersion,
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
					Controller: ptr(true),
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "http-mcp",
					Port:       8888,
					TargetPort: intstr.FromInt(8888),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Try to get existing service
	var existingService corev1.Service
	serviceKey := client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}
	err := cli.Get(ctx, serviceKey, &existingService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Service doesn't exist, create it
			logger.Info("Creating MCP service", "serviceName", serviceName, "namespace", namespace)
			if err := cli.Create(ctx, desiredService); err != nil {
				return fmt.Errorf("failed to create MCP service: %w", err)
			}
			logger.Info("Successfully created MCP service", "serviceName", serviceName)
			return nil
		}
		return fmt.Errorf("failed to get existing service: %w", err)
	}

	// Service exists, check if it needs updating
	if !serviceMCPPortExists(&existingService) {
		logger.Info("Updating existing MCP service to add port", "serviceName", serviceName)
		// Copy the existing service and add/update the MCP port
		existingService.Spec.Ports = ensureMCPPortInPorts(existingService.Spec.Ports)
		existingService.Spec.Selector = selector

		if err := cli.Update(ctx, &existingService); err != nil {
			return fmt.Errorf("failed to update MCP service: %w", err)
		}
		logger.Info("Successfully updated MCP service", "serviceName", serviceName)
	} else {
		logger.Debug("MCP service already exists with correct configuration", "serviceName", serviceName)
	}

	return nil
}

// serviceMCPPortExists checks if the http-mcp port exists in the service
func serviceMCPPortExists(svc *corev1.Service) bool {
	for _, port := range svc.Spec.Ports {
		if port.Name == "http-mcp" {
			return true
		}
	}
	return false
}

// ensureMCPPortInPorts adds or updates the MCP port in the port list
func ensureMCPPortInPorts(ports []corev1.ServicePort) []corev1.ServicePort {
	// Check if port already exists
	for i, port := range ports {
		if port.Name == "http-mcp" {
			// Update existing port
			ports[i].Port = 8888
			ports[i].TargetPort = intstr.FromInt(8888)
			ports[i].Protocol = corev1.ProtocolTCP
			return ports
		}
	}

	// Add new port
	return append(ports, corev1.ServicePort{
		Name:       "http-mcp",
		Port:       80,
		TargetPort: intstr.FromInt(8888),
		Protocol:   corev1.ProtocolTCP,
	})
}

// ptr is a helper to get a pointer to a bool
func ptr(b bool) *bool {
	return &b
}

// deleteMCPService deletes the dedicated MCP service for a cluster
func deleteMCPService(ctx context.Context, cli client.Client, cluster *apiv1.Cluster) error {
	logger := log.FromContext(ctx).WithName("mcp_service_manager")

	serviceName := fmt.Sprintf("%s-mcp", cluster.Name)
	namespace := cluster.Namespace

	logger.Info("Deleting MCP service", "serviceName", serviceName, "namespace", namespace)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
	}

	err := cli.Delete(ctx, service)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Service doesn't exist, which is fine
			logger.Debug("MCP service not found, nothing to delete", "serviceName", serviceName)
			return nil
		}
		return fmt.Errorf("failed to delete MCP service: %w", err)
	}

	logger.Info("Successfully deleted MCP service", "serviceName", serviceName)
	return nil
}
