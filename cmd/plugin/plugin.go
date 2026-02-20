package plugin

import (
	cnpgv1 "github.com/cloudnative-pg/api/pkg/api/v1"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/http"
	"github.com/cloudnative-pg/cnpg-i/pkg/lifecycle"
	"github.com/cloudnative-pg/cnpg-i/pkg/reconciler"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sharifmshaker/cnpg-i-mcp-demo/internal/identity"
	lifecycleImpl "github.com/sharifmshaker/cnpg-i-mcp-demo/internal/lifecycle"
	reconcilerImpl "github.com/sharifmshaker/cnpg-i-mcp-demo/internal/reconciler"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cnpgv1.AddToScheme(scheme))
}

var kubeClient client.Client

// initKubeClient initializes the Kubernetes client
func initKubeClient() error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	cli, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}

	kubeClient = cli
	return nil
}

// NewCmd creates the `plugin` command
func NewCmd() *cobra.Command {
	cmd := http.CreateMainCmd(identity.Implementation{}, func(server *grpc.Server) error {
		// Register the declared implementations
		lifecycle.RegisterOperatorLifecycleServer(server, lifecycleImpl.Implementation{})
		reconciler.RegisterReconcilerHooksServer(server, reconcilerImpl.Implementation{
			Client: kubeClient,
		})
		return nil
	})
	cmd.Use = "plugin"

	// Initialize Kubernetes client before starting the server
	cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initKubeClient()
	}

	return cmd
}
