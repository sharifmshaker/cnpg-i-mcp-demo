package plugin

import (
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper/http"
	"github.com/cloudnative-pg/cnpg-i/pkg/lifecycle"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/sharifmshaker/cnpg-i-mcp-demo/internal/identity"
	lifecycleImpl "github.com/sharifmshaker/cnpg-i-mcp-demo/internal/lifecycle"
)

// NewCmd creates the `plugin` command
func NewCmd() *cobra.Command {
	cmd := http.CreateMainCmd(identity.Implementation{}, func(server *grpc.Server) error {
		// Register the declared implementations
		lifecycle.RegisterOperatorLifecycleServer(server, lifecycleImpl.Implementation{})
		return nil
	})
	cmd.Use = "plugin"
	return cmd
}
