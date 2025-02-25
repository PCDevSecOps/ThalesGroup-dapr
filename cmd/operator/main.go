/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/dapr/dapr/cmd/operator/options"
	"github.com/dapr/dapr/pkg/buildinfo"
	"github.com/dapr/dapr/pkg/metrics"
	"github.com/dapr/dapr/pkg/operator"
	"github.com/dapr/dapr/pkg/operator/monitoring"
	"github.com/dapr/dapr/pkg/signals"
	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("dapr.operator")

func main() {
	log.Infof("starting Dapr Operator -- version %s -- commit %s", buildinfo.Version(), buildinfo.Commit())

	opts, err := options.New()
	if err != nil {
		log.Fatal(err)
	}

	metricsExporter := metrics.NewExporterWithOptions(metrics.DefaultMetricNamespace, opts.Metrics)

	// Apply options to all loggers
	if err = logger.ApplyOptionsToLoggers(&opts.Logger); err != nil {
		log.Fatal(err)
	}
	log.Infof("log level set to: %s", opts.Logger.OutputLevel)

	// Initialize dapr metrics exporter
	if err = metricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	if err = monitoring.InitMetrics(); err != nil {
		log.Fatal(err)
	}

	operatorOpts := operator.Options{
		Config:                              opts.Config,
		CertChainPath:                       opts.CertChainPath,
		LeaderElection:                      !opts.DisableLeaderElection,
		WatchdogMaxRestartsPerMin:           opts.MaxPodRestartsPerMinute,
		WatchNamespace:                      opts.WatchNamespace,
		ServiceReconcilerEnabled:            !opts.DisableServiceReconciler,
		ArgoRolloutServiceReconcilerEnabled: opts.EnableArgoRolloutServiceReconciler,
		WatchdogEnabled:                     opts.WatchdogEnabled,
		WatchdogInterval:                    opts.WatchdogInterval,
		WatchdogCanPatchPodLabels:           opts.WatchdogCanPatchPodLabels,
	}

	ctx := signals.Context()

	op, err := operator.NewOperator(ctx, operatorOpts)
	if err != nil {
		log.Fatalf("error creating operator: %v", err)
	}

	err = op.Run(ctx)
	if err != nil {
		log.Fatalf("error running operator: %v", err)
	}
	log.Info("operator shut down gracefully")
}
