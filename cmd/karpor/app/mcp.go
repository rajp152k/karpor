// Copyright The Karpor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"fmt"
	"os"

	_ "github.com/KusionStack/karpor/pkg/infra/search/storage/elasticsearch"
	_ "github.com/KusionStack/karpor/pkg/mcp"
	_ "github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

type mcpOptions struct {
	SSEPort string
	// TODO update mcpOptions to use the generic storage interface
	// should be able handle accept multiple storage backends
	// as of now, this in alignment with whatever the syncer syncs
	ElasticSearchAddresses []string

	// Flags for health and metrics probes
	MetricsAddr string
	ProbeAddr   string
}

func NewMCPOptions() *mcpOptions {
	return &mcpOptions{}
}

func (o *mcpOptions) AddFlags(fs *pflag.FlagSet) {
	// TODO chart out how to handle multiple generic storage backends
	// as of now this is in alignment with whatever the syncer syncs
	fs.StringVar(&o.SSEPort, "mcp-sse-port", ":7999", "The address exposing the mcp server SSE endpoint.")
	fs.StringSliceVar(&o.ElasticSearchAddresses, "elastic-search-addresses", nil, "The elastic search address.")

	// Add flags for health and metrics probes
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
}

func NewMCPCommand(ctx context.Context) *cobra.Command {
	options := NewMCPOptions()
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "start a storage mcp server to enable natural language interaction capabilities with the storage backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mcpRun(ctx, options)
		},
	}
	options.AddFlags(cmd.Flags())
	return cmd
}

// TODO update mcpOptions to use the generic storage interface
//nolint:unparam
func mcpRun(ctx context.Context, options *mcpOptions) error {
	ctrl.SetLogger(klog.NewKlogr())
	log := ctrl.Log.WithName("mcp")

	log.Info("Starting MCP SSE server",
		"port", options.SSEPort,
		"metrics-bind-address", options.MetricsAddr,
		"health-probe-bind-address", options.ProbeAddr,
	)

	// Initialize controller-runtime manager for health/readyz probes and metrics
	// This requires a Kubernetes config, even if the MCP server doesn't directly use it yet.
	// TODO: Re-evaluate if a full manager is needed just for probes, or if a simpler http server is sufficient.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		MetricsBindAddress:     options.MetricsAddr,
		HealthProbeBindAddress: options.ProbeAddr,
		// We don't need leader election or controllers for just probes/metrics
		LeaderElection: false,
	})
	if err != nil {
		log.Error(err, "unable to create manager for probes")
		return fmt.Errorf("unable to create manager for probes: %w", err)
	}

	// Add health and readiness checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	// Start the manager's probe/metrics server in a goroutine
	go func() {
		log.Info("starting manager for probes and metrics")
		// Use a separate context for the manager's Start method if needed,
		// but the main ctx should work for graceful shutdown.
		if err := mgr.Start(ctx); err != nil {
			log.Error(err, "problem running manager for probes and metrics")
			// If the probe server fails, the main process should likely exit
			os.Exit(1)
		}
	}()


	//TODO update to use the generic storage interface for initialization
	//nolint:contextcheck
	// es, err := elasticsearch.NewStorage(esclient.Config{
	// 	Addresses: options.ElasticSearchAddresses,
	// })
	// if err != nil {
	// 	log.Error(err, "unable to init elasticsearch client")
	// 	return err
	// }
	// log.Info("Acquired elasticsearch storage backend", "esStorage", es)


	//TODO pickup syncer operations patterns for running the mcp server from app/syncer.go

	log.Info("TODO: yet to implement mcp functionality")
	log.Info("see /cmd/karpor/app/mcp.go for further directives")

	// TODO: Initialize and start the actual MCP SSE server here.
	// The MCP server's Serve() method is blocking, so this call would
	// typically be the last thing in this function, or managed alongside
	// the manager's Start using a run.Group or similar.
	// Example (placeholder):
	// mcpServer := mcp.NewMCPStorageServer(nil, "http://localhost"+options.SSEPort, options.SSEPort) // Pass actual storage
	// if err := mcpServer.Serve(); err != nil {
	// 	log.Error(err, "problem running MCP SSE server")
	// 	return err
	// }


	// This return will be reached immediately after starting the manager goroutine
	// and before the actual MCP server is started (which is still TODO).
	// The function should block until the main server (TODO) or context is done.
	// For now, we'll just return nil, but this needs correction when the MCP server is added.
	// A common pattern is to wait on the context's Done channel.
	<-ctx.Done()
	log.Info("MCP command shutting down")
	return nil
}
