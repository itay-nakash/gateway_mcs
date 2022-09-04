package multicluster_gw

import (
	"flag"
	"net"
	"os"
	"sync"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	mcsv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

const pluginName = "multicluster_gw"

var (
	gateway_ip4 net.IP
	gateway_ip6 net.IP
	scheme      = runtime.NewScheme()
	setupLog    = ctrl.Log.WithName("setup")
	Mcgw        MulticlusterGw
)

// init registers this plugin.
func init() {
	log.Debug("Started init function")
	plugin.Register(pluginName, Mcgw.setup)

}

func (Mcgw *MulticlusterGw) setup(c *caddy.Controller) error {
	log.Info("Started setup function")
	Mcgw.SISet.mutex = new(sync.RWMutex)
	err := ParseStanza(c, Mcgw)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	// TODO: check about the chanells that its the right way to do so:
	initializeController()
	log.Info("Finished initialize Controllere function")

	//block until chanell gets a value in 'initializeController':

	log.Info("Started to register the Plugin")
	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		Mcgw.Next = next
		return Mcgw
	})
	log.Info("Finished adding the Plugin")

	// All OK, return a nil error.
	return nil
}

// ParseStanza parses a kubernetes stanza
func ParseStanza(c *caddy.Controller, mcgw *MulticlusterGw) error {
	c.Next() // Skip pluginName label
	log.Info("Started to parse Stanza")

	zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
	mcgw.New(zones)

	for c.NextBlock() {
		switch c.Val() {
		case "kubeconfig":
			args := c.RemainingArgs()
			if len(args) != 1 && len(args) != 2 {
				return c.ArgErr()
			}
			overrides := &clientcmd.ConfigOverrides{}
			if len(args) == 2 {
				overrides.CurrentContext = args[1]
			}
			config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: args[0]},
				overrides,
			)
			mcgw.ClientConfig = config

		case "fallthrough":
			mcgw.Fall.SetZonesFromArgs(c.RemainingArgs())

		case "gateway_ip":
			mcgw.gatewayIp4, mcgw.gatewayIp6 = parseIp(c)

		default:
			return c.Errf("unknown property '%s'", c.Val())
		}
	}
	log.Info("Finish to parse Stanza")

	return nil
}

// parse the Ip given as caddy.Controller arg, as a string, to ipv4 and ipv6 format
func parseIp(c *caddy.Controller) (net.IP, net.IP) {
	ipAsString := c.RemainingArgs()[0]
	ip := net.ParseIP(ipAsString)
	if ip == nil {
		//The ip was given as string, not a number
		return nil, nil
	} else {
		return ip.To4(), ip.To16()
	}
}

func initializeController() {
	log.Info("Started to initialize Controller")
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(mcsv1a1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8081", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8082", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ecaf1259.my.domain",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&ServiceImportReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceImportController")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder
	/* don't need to healthCheck, already happens in coredns
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	*/
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	go activateManager(mgr)
}

func activateManager(mgr manager.Manager) {
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
