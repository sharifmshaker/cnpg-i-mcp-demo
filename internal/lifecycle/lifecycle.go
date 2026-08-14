package lifecycle

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/common"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/decoder"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/object"
	"github.com/cloudnative-pg/cnpg-i/pkg/lifecycle"
	"github.com/cloudnative-pg/machinery/pkg/log"
	corev1 "k8s.io/api/core/v1"

	"github.com/sharifmshaker/cnpg-i-mcp-demo/internal/config"
	"github.com/sharifmshaker/cnpg-i-mcp-demo/internal/utils"
	"github.com/sharifmshaker/cnpg-i-mcp-demo/pkg/metadata"
)

// Implementation is the implementation of the lifecycle handler
type Implementation struct {
	lifecycle.UnimplementedOperatorLifecycleServer
}

// GetCapabilities exposes the lifecycle capabilities
func (impl Implementation) GetCapabilities(
	_ context.Context,
	_ *lifecycle.OperatorLifecycleCapabilitiesRequest,
) (*lifecycle.OperatorLifecycleCapabilitiesResponse, error) {
	return &lifecycle.OperatorLifecycleCapabilitiesResponse{
		LifecycleCapabilities: []*lifecycle.OperatorLifecycleCapabilities{
			{
				Group: "",
				Kind:  "Pod",
				OperationTypes: []*lifecycle.OperatorOperationType{
					{
						Type: lifecycle.OperatorOperationType_TYPE_CREATE,
					},
					{
						Type: lifecycle.OperatorOperationType_TYPE_EVALUATE,
					},
				},
			},
		},
	}, nil
}

// LifecycleHook is called when creating Kubernetes services
func (impl Implementation) LifecycleHook(
	ctx context.Context,
	request *lifecycle.OperatorLifecycleRequest,
) (*lifecycle.OperatorLifecycleResponse, error) {
	kind, err := utils.GetKind(request.GetObjectDefinition())
	if err != nil {
		return nil, err
	}
	operation := request.GetOperationType().GetType().Enum()
	if operation == nil {
		return nil, errors.New("no operation set")
	}

	//nolint: gocritic
	switch kind {
	case "Pod":
		switch *operation {
		case lifecycle.OperatorOperationType_TYPE_CREATE, lifecycle.OperatorOperationType_TYPE_EVALUATE:
			return impl.reconcilePod(ctx, request)
		}
	}

	return &lifecycle.OperatorLifecycleResponse{}, nil
}

// reconcilePod is called when creating an instance Pod
func (impl Implementation) reconcilePod(
	ctx context.Context,
	request *lifecycle.OperatorLifecycleRequest,
) (*lifecycle.OperatorLifecycleResponse, error) {
	cluster, err := decoder.DecodeClusterLenient(request.GetClusterDefinition())
	if err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).WithName("cnpg_i_mcp_pod_lifecycle")
	helper := common.NewPlugin(
		*cluster,
		metadata.PluginName,
	)

	configuration, valErrs := config.FromParameters(helper)
	if len(valErrs) > 0 {
		return nil, valErrs[0]
	}

	pod, err := decoder.DecodePodJSON(request.GetObjectDefinition())
	if err != nil {
		return nil, err
	}

	if configuration.ReplicaOnly && (pod.Labels == nil || pod.Labels["cnpg.io/instanceRole"] != "replica") {
		logger.Debug("Skipping injection of MCP sidecar on non-replica pod")
		return &lifecycle.OperatorLifecycleResponse{}, nil
	}

	mutatedPod := pod.DeepCopy()
	bootstrap := cluster.Spec.Bootstrap
	var dbName string
	switch {
	case bootstrap.InitDB != nil:
		dbName = bootstrap.InitDB.Database
	case bootstrap.PgBaseBackup != nil:
		dbName = bootstrap.PgBaseBackup.Database
	case bootstrap.Recovery != nil:
		dbName = bootstrap.Recovery.Database
	}

	err = object.InjectPluginInitContainerSidecarSpec(&mutatedPod.Spec, &corev1.Container{
		Name:  "postgres-mcp",
		Image: fmt.Sprintf("crystaldba/postgres-mcp:%s", configuration.MCPImageTag),
		Command: []string{
			"postgres-mcp",
		},
		Args: []string{
			"--transport=sse",
			"--access-mode=unrestricted",
			"--sse-port=8888",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http-mcp",
				ContainerPort: 8888,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "DATABASE_URI",
				Value: fmt.Sprintf("postgresql://postgres@/%s?host=/postgresql/run", dbName),
			},
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "scratch-data",
			ReadOnly:  false,
			MountPath: "/postgresql",
		}},
	}, false)
	if err != nil {
		return nil, err
	}

	patch, err := object.CreatePatch(mutatedPod, pod)
	if err != nil {
		return nil, err
	}

	logger.Debug("generated pod patch", "content", string(patch), "configuration", configuration)

	return &lifecycle.OperatorLifecycleResponse{
		JsonPatch: patch,
	}, nil
}
