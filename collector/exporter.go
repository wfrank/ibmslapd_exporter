package collector

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	LdapURI *string
	BindDn  *string
	BindPw  *string
}

type Exporter struct {
	mutex  sync.RWMutex
	logger *slog.Logger

	ldapURI string
	bindDn  string
	bindPw  string

	ldapConn   *ldap.Conn
	collectors []prometheus.Collector

	up             *prometheus.Desc
	info           *prometheus.Desc
	scrapeFailures *prometheus.Desc
}

func NewExporter(logger *slog.Logger, config *Config) *Exporter {
	e := &Exporter{
		logger: logger,

		ldapURI: *config.LdapURI,
		bindDn:  *config.BindDn,
		bindPw:  *config.BindPw,

		up: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "up"),
			"Could the ibmslapd server be reached",
			nil,
			nil),
		info: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "info"),
			"Could the ibmslapd server be reached",
			[]string{"vendor", "version", "server_id"},
			nil),
		scrapeFailures: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "exporter", "scrape_failures_total"),
			"Number of errors while scraping ibmslapd.",
			nil,
			nil),
	}
	e.collectors = append(e.collectors, NewMonitorCollecter(e))
	e.collectors = append(e.collectors, NewReplicationCollecter(e))

	return e
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.info
	ch <- e.scrapeFailures

	for _, v := range e.collectors {
		v.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	l, err := ldap.DialURL(e.ldapURI)
	if err != nil {
		e.logger.Error("Error contacting LDAP server", "err", err)
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return
	}
	defer l.Close()
	l.SetTimeout(1 * time.Second)
	e.ldapConn = l

	q := ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{"*"}, nil,
	)
	p, err := l.Search(q)
	if err != nil {
		e.logger.Error("Error querying Root DSE", "err", err)
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	x := p.Entries[0]
	id := x.GetAttributeValue("ibm-serverId")
	vendor := x.GetAttributeValue("vendorname")
	version := x.GetAttributeValue("vendorversion")
	ch <- prometheus.MustNewConstMetric(e.info, prometheus.GaugeValue, 1, vendor, version, id)

	for _, v := range e.collectors {
		v.Collect(ch)
	}
}
