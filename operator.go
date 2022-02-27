/*
Copyright 2022.

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
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"git.earthnet.ch/simon.beck/kopia-k8s/controllers"
	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

type operator struct {
	log    logr.Logger
	cliCtx *cli.Context
	uuid   string
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func newOperator(cliCtx *cli.Context) *operator {
	log := logger.AppLogger(cliCtx.Context).WithName("operator")

	return &operator{
		cliCtx: cliCtx,
		log:    log,
		uuid:   cliCtx.String("uuid"),
	}
}

func (o *operator) runOperator() {
	mgr := o.initManager()

	o.registerController(mgr)

	o.startManager(mgr)
}

func (o *operator) registerController(mgr manager.Manager) {

	for name, reconciler := range map[string]controllers.ReconcilerSetup{
		"Job": &controllers.JobReconciler{},
	} {
		if err := reconciler.SetupWithManager(mgr, o.log.WithName("controllers").WithName(name), o.uuid); err != nil {
			o.log.Error(err, "unable to initialize operator mode", "step", "controller", "controller", name)
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		o.log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		o.log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
}

func (o *operator) startManager(mgr manager.Manager) {
	o.log.Info("starting manager")
	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			o.log.Error(err, "problem running manager")
			os.Exit(1)
		}
	}()
	time.Sleep(1 * time.Second)
}

func (o *operator) initManager() manager.Manager {
	ctrl.SetLogger(o.log)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     o.cliCtx.String("metricsAddr"),
		Port:                   9443,
		HealthProbeBindAddress: o.cliCtx.String("probeAddr"),
		LeaderElection:         o.cliCtx.Bool("enableLeaderElection"),
		LeaderElectionID:       "68072c41.earthnet.ch",
	})
	if err != nil {
		o.log.Error(err, "unable to start manager")
		os.Exit(1)
	}
	return mgr
}
