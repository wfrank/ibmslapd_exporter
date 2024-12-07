package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/wfrank/ibmslapd_exporter/collector"
)

var (
	metricsEndpoint = kingpin.Flag("telemetry.endpoint", "Path under which to expose metrics.").Default("/metrics").String()
	toolkitFlags    = kingpinflag.AddFlags(kingpin.CommandLine, ":9981")
)

func main() {
	exporterConfig := &collector.Config{
		LdapURI: kingpin.Flag("ldap_uri", "URI referring to the ldap server, only the protocol/host/port fields are allowed.").Default("ldap://localhost:389").String(),
		BindDn:  kingpin.Flag("bind_dn", "Binding DN to authenticate the LDAP connections.").Default("cn=root").String(),
		BindPw:  kingpin.Flag("bind_pw", "Password of the Binding DN.").String(),
	}

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print("ibmslapd_exporter"))
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	exporter := collector.NewExporter(logger, exporterConfig)
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(versioncollector.NewCollector("ibmslapd_exporter"))

	logger.Info("Starting ibmslapd_exporter", "version", version.Info())
	logger.Info("Build context", "build", version.BuildContext())
	logger.Info("Collect metrics from", "ldap_uri", *exporterConfig.LdapURI)

	landingConfig := web.LandingConfig{
		Name:        "ibmslapd exporter",
		Description: "Prometheus exporter for IBM Security Verify Directory metrics",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address: *metricsEndpoint,
				Text:    "Metrics",
			},
		},
	}
	landingPage, err := web.NewLandingPage(landingConfig)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	http.Handle("/", landingPage)
	http.Handle(*metricsEndpoint, promhttp.Handler())
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
