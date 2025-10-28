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

	// If you want to provide your own logr.Logger here, inject it into a context.Context
	// with logr.NewContext(ctx, logger) and pass it to cmd.SetContext(ctx)

	// Additional custom behaviour can be added by wrapping cmd.PersistentPreRun or cmd.Run

	cmd.Use = "plugin"

	return cmd
}
