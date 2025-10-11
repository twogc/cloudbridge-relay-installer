package main

import (
	"flag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	calicov1 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	controllers "github.com/twogc/cloudbridge-relay-installer/internal/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Регистрируем Calico CRDs
	utilruntime.Must(calicov1.AddToScheme(scheme))

	// Регистрируем наши CRDs (если есть)
	// utilruntime.Must(cloudbridgev1.AddToScheme(scheme))
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		watchNamespace       string
		pprofAddr            string
		concurrency          int
		resync               time.Duration
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "metrics endpoint bind address")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "health/ready endpoint bind address")
	flag.StringVar(&pprofAddr, "pprof-bind-address", "", "pprof bind address (empty to disable)")
	flag.StringVar(&watchNamespace, "watch-namespace", "", "if set, restricts the manager's cache to a single namespace")
	flag.IntVar(&concurrency, "concurrency", 2, "max concurrent reconciles for TenantPolicy controller")
	flag.DurationVar(&resync, "cache-resync", 0, "cache resync period (0 to disable)")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true, "enable leader election for controller manager")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgrOpts := ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "tenant-policy-controller.cloudbridge.io",
		Cache: cache.Options{
			// Ограничение видимости, если нужно
			ByObject: map[client.Object]cache.ByObject{},
			SyncPeriod: func() *time.Duration {
				if resync <= 0 {
					return nil
				}
				return &resync
			}(),
		},
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOpts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Индексатор по лейблу тенанта — ускоряет выборки и события
	// Примечание: controller-runtime индексирует по полям, а нам нужен индекс по label.
	// Оставляем как TODO, т.к. текущий контроллер уже использует LabelSelector.

	// Регистрируем контроллер
	reconciler := &controllers.TenantPolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("TenantPolicy"),
	}
	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TenantPolicy")
		os.Exit(1)
	}

	// Health/ready — без привязки к webhook server
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager",
		"metrics", metricsAddr,
		"health", probeAddr,
		"pprof", pprofAddr,
		"leaderElection", enableLeaderElection,
		"watchNamespace", watchNamespace,
	)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
