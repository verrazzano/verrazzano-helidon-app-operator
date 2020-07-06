// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"runtime"
	"strconv"

	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/verrazzano/verrazzano-helidon-app-operator/pkg/apis"
	"github.com/verrazzano/verrazzano-helidon-app-operator/pkg/controller"
	"github.com/verrazzano/verrazzano-helidon-app-operator/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)


func printVersion() {
	// Initialize logger for version output
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "HelidonAppOperator").Str("name", "HelidonInit").Logger()
	logger.Info().Msg(fmt.Sprintf("Operator Version: %s", version.Version))
	logger.Info().Msg(fmt.Sprintf("Go Version: %s", runtime.Version()))
	logger.Info().Msg(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	logger.Info().Msg(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Initialize structured logging
	InitLogs()

	// Create log instance for function usage
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "HelidonAppOperator").Str("name", "HelidonInit").Logger()

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logger.Error().Msgf("Failed to get watch namespace: %s", err)
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "helidon-app-lock")
	if err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}

	logger.Info().Msg("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}

	if err = serveCRMetrics(cfg); err != nil {
		logger.Info().Msgf("Could not generate and serve custom resource metrics, error: %s", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}
	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		logger.Info().Msgf("Could not create metrics Service, error: %s", err.Error())
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	services := []*v1.Service{service}
	_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
	if err != nil {
		logger.Info().Msgf("Could not create ServiceMonitor object, error: %s", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			logger.Info().Msgf("Install prometheus-operator in your cluster to create ServiceMonitor objects, error %s", err.Error())
		}
	}

	logger.Info().Msg("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error().Msgf("Manager exited non-zero: %s", err.Error())
		os.Exit(1)
	}
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}

// Initialize logs with Time and Global Level of Logs set at Info
func InitLogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Log levels are outlined as follows:
	// Panic: 5
	// Fatal: 4
	// Error: 3
	// Warn: 2
	// Info: 1
	// Debug: 0
	// Trace: -1
	// more info can be found at https://github.com/rs/zerolog#leveled-logging

	envLog := os.Getenv("LOG_LEVEL")
	if val, err := strconv.Atoi(envLog); envLog != "" && err == nil && val >= -1 && val <= 5 {
		zerolog.SetGlobalLevel(zerolog.Level(val))
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}