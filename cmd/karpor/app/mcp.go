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

	"github.com/KusionStack/karpor/pkg/infra/search/storage"
	"github.com/KusionStack/karpor/pkg/infra/search/storage/elasticsearch"
	"github.com/KusionStack/karpor/pkg/mcp"
	esclient "github.com/elastic/go-elasticsearch/v8"
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
	// TODO: Make readyz check dependent on storage backend health
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


	// Initialize storage backend (Elasticsearch for now)
	// TODO update to use the generic storage interface for initialization
	//nolint:contextcheck // Context is passed to storage methods, not NewStorage
	es, err := elasticsearch.NewStorage(esclient.Config{
		Addresses: options.ElasticSearchAddresses,
	})
	if err != nil {
		log.Error(err, "unable to init elasticsearch client")
		return fmt.Errorf("unable to init elasticsearch client: %w", err)
	}
	log.Info("Acquired elasticsearch storage backend", "esStorage", es)

	// TODO pickup syncer operations patterns for running the mcp server from app/syncer.go

	// Initialize and start the actual MCP SSE server here.
	// Pass the initialized storage backend(s) to the MCP server.
	// The MCP server's Serve() method is blocking, so this call would
	// typically be the last thing in this function, or managed alongside
	// the manager's Start using a run.Group or similar.
	// For now, we pass the ES storage as a single-element slice.
	mcpServer := mcp.NewMCPStorageServer([]storage.Storage{es}, "http://localhost"+options.SSEPort, options.SSEPort) // Pass actual storage

	log.Info("Starting MCP SSE server...")
	// The Serve method is blocking.
	if err := mcpServer.Serve(); err != nil {
		log.Error(err, "problem running MCP SSE server")
		return fmt.Errorf("problem running MCP SSE server: %w", err)
	}

	// This return will only be reached if mcpServer.Serve() returns without error,
	// or if the context is cancelled and Serve() respects it.
	// A common pattern is to wait on the context's Done channel *after* starting
	// all blocking components, or use a run.Group.
	// For now, the Serve() call is blocking, so the line below is unreachable
	// unless Serve() is modified to return on context cancellation.
	// <-ctx.Done() // This line is now effectively unreachable
	log.Info("MCP command shutting down")
	return nil // Or return the error from mcpServer.Serve()
}
