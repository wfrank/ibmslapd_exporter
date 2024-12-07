package collector

import (
	"strconv"

	"github.com/go-ldap/ldap/v3"
	"github.com/prometheus/client_golang/prometheus"
)

// cn=monitor
// version=IBM Security Verify Directory (SSL), Version 10.0.3
// directoryversion
//     The specific version number that indicates the fixpack level.

// maxconnections
//     The maximum number of active connections allowed.

// opsinitiated
//     The number of requests since the server was started.
// opscompleted
//     The number of completed requests since the server was started.

// transactionsrequested
//     The number of transaction requests initiated.
// transactionscompleted
//     The number of transaction operations completed.
// transactionpreparesrequested
//     The number of prepare transaction operations that are requested.
// transactionpreparescompleted
//     The number of prepare transaction operations that are completed.
// transactioncommitsrequested
//     The number of commit transaction operations requested.
// transactionscommitted
//     The number of transaction operations committed.
// transactionrollbacksrequested
//     The number of transaction operations that are requested for rollback.
// transactionsrolledback
//     The number of transaction operations that are rolled back.
// transactionspreparedwaitingoncommit
//     The number of transaction operations, which are prepared and waiting for commit/rollback.

// slapderrorlog_messages
//     The number of server error messages that are recorded since the server was started or since a reset was performed.
// slapdclierrors_messages
//     The number of DB2® error messages that are recorded since the server was started or since a reset was performed.
// auditlog_messages
//     The number of audit messages that are recorded since the server was started or since a reset was performed.
// auditlog_failedop_messages
//     The number of failed operation messages that are recorded since the server was started or since a reset was performed.

// filter_cache_size
//     The maximum number of filters that are allowed in the cache.
// filter_cache_current
//     The number of filters currently in the cache.
// filter_cache_hit
//     The number of filters that are found in the cache.
// filter_cache_miss
//     The number of search operations that attempted to use the filter cache, but did not find a matching operation in the cache.
// filter_cache_bypass_limit
//     Search filters that return more entries than this limit are not cached.

// entry_cache_size
//     The maximum number of entries that are allowed in the cache.
// entry_cache_current
//     The number of entries currently in the cache.
// entry_cache_hit
//     The number of entries that are found in the cache.
// entry_cache_miss
//     The number of entries that are not found in the cache.

// group_members_cache_size
//     The maximum number of groups whose members needs to be cached.
// group_members_cache_current
//     The number of groups whose members are currently cached.
// group_members_cache_hit
//     The number of groups whose members were requested and retrieved from the group members’ cache.
// group_members_cache_miss
//     The number of groups whose members were requested and found in the group members’ cache that needed to have the members that are retrieved from DB2.
// group_members_cache_bypass
//     The maximum number of members that are allowed in a group that is cached in the group members’ cache.

// acl_cache
//     A Boolean value that indicates the ACL cache is active (TRUE) or inactive (FALSE).
// acl_cache_size
//     The maximum number of entries in the ACL cache.

// maximum_operations_waiting
//     The maximum number of operations waiting in the deadlock detector at a time.

// cached_attribute_total_size
//     The amount of memory in kilobytes used by attribute caching.
// cached_attribute_configured_size
//     The amount of memory in kilobytes that can be used by attribute caching.
// cached_attribute_auto_adjust
//     Indicates if attribute cache auto adjusting is configured to be on or off.
// cached_attribute_auto_adjust_time
//     Indicates the configured time on which to start attribute cache auto adjusting.
// cached_attribute_auto_adjust_time_interval
//     Indicates the time interval after which to repeat attribute cache auto adjusting for the day.
// cached_attribute_hit
//     The number of times the attribute is used in a filter that could be processed by the changelog attribute cache. The value is reported as follows:

//     cached_attribute_hit=attrname:#####

// cached_attribute_size
//     The amount of memory that is used for this attribute in the changelog attribute cache. This value is reported in kilobytes as follows:

//     cached_attribute_size=attrname:######

// cached_attribute_candidate_hit
//     A list of up to ten most frequently used non-cached attributes that is used in a filter that is processed by the changelog attribute cache if all of the attributes that are used in the filter is cached. The value is reported as follows:

//     cached_attribute_candidate_hit=attrname:#####

//     You can use this list to help you decide which attributes you want to cache. Typically, you want to put a limited number of attributes into the attribute cache because of memory constraints.
// currenttime
//     The current time on the server. The current time is in the format:

//     year-month-day hour:minutes:seconds GMT

// starttime
//     The time the server was started. The start time is in the format:

//     year-month-day hour:minutes:seconds GMT

// trace_enabled
//     The current trace value for the server. TRUE, if you are collecting trace data, and FALSE, if you are not collecting trace data. See the ldaptrace command information in the Command Reference for information about enabling and starting the trace function.
// trace_message_level
//     The current ldap_debug value for the server. The value is in hexadecimal form, for example:

//     0x0=0
//     0xffff=65535

//     For more information, see the section on Debugging levels in the Command Reference.
// trace_message_log
//     The current LDAP_DEBUG_FILE environment variable setting for the server.
// auditinfo
//     Contains the current audit configuration. This attribute is displayed only if the monitor search is initiated by an administrator.

// en_currentregs
//     The current number of client registrations for event notification.
// en_notificationssent
//     The total number of event notifications sent to clients since the server was started.

// currentpersistentsearches
//     Indicates number of active persistent search connections.
// persistentsearchpendingchanges
//     Indicates the number of new updates in the queue that are yet to be processed by the persistent search thread.
// persistentsearchprocessedchanges
//     Indicates number of changes that are processed by persistent search process.
// lostpersistentsearchconns
//     Indicates the number of lost persistent search connections.

// bypass_deref_aliases
//     The server runtime value that indicates if alias processing can be bypassed. It displays true, if no alias object exists in the directory, and false, if at least one alias object exists in the directory.

// largest_workqueue_size
//     The largest size that the work queue.

type MonitorCollecter struct {
	exporter *Exporter

	entriesSent              *prometheus.Desc
	currentConnections       *prometheus.Desc
	currentWorkQueueDepth    *prometheus.Desc
	idleConnectionsClosed    *prometheus.Desc
	autoConnectionCleanerRun *prometheus.Desc
	workerThreads            *prometheus.Desc
	totalConnections         *prometheus.Desc
	operationsRequested      *prometheus.Desc
	operationsCompleted      *prometheus.Desc
	operationsFromSuppliers  *prometheus.Desc
	operationsWaiting        *prometheus.Desc
	operationsRetried        *prometheus.Desc
	operationsDeadlocked     *prometheus.Desc
}

func NewMonitorCollecter(e *Exporter) *MonitorCollecter {
	return &MonitorCollecter{
		exporter: e,
		entriesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "entries_sent_total"),
			"The number of entries that are sent by the server since the server was started.",
			nil,
			nil),
		currentConnections: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "current_connections"),
			"The number of active connections.",
			nil,
			nil),
		currentWorkQueueDepth: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "current_work_queue_depth"),
			"The current depth of the work queue.",
			nil,
			nil),
		idleConnectionsClosed: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "idle_connections_closed"),
			"The number of idle connections closed by the Automatic Connection Cleaner.",
			nil,
			nil),
		autoConnectionCleanerRun: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "auto_connection_cleaner_run"),
			"The number of times that the Automatic Connection Cleaner is run.",
			nil,
			nil),
		totalConnections: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "connections_total"),
			"The total number of connections of different kinds(tcp, ssl, tls) since the server was started.",
			[]string{"connection"},
			nil),
		workerThreads: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "worker_threads"),
			"The number of threads in different states(read, write, live, idle). read: reading data from the client; write: sending data back to the client; live: used by the server; idle: available for work.",
			[]string{"state"},
			nil),
		operationsRequested: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_requested_total"),
			"The number of requested operations of different kinds(search, bind, unbind, add, delete, modrdn, modify, compare, abandon, extop, unknownop) since the server was started.",
			[]string{"operation"},
			nil),
		operationsCompleted: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_completed_total"),
			"The number of completed operations of different kinds(search, bind, unbind, add, delete, modrdn, modify, compare, abandon, extop, unknownop) since the server was started.",
			[]string{"operation"},
			nil),
		operationsFromSuppliers: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_from_suppliers_total"),
			"The number of operations of different kinds(add, delete, modrdn, modify) that are received from replication supplier.",
			[]string{"operation"},
			nil),
		operationsWaiting: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_waiting"),
			"The number of operations that are waiting in the deadlock detector.",
			nil,
			nil),
		operationsRetried: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_retried_total"),
			"The number of operations retired due to deadlocks.",
			nil,
			nil),
		operationsDeadlocked: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "", "operations_deadlocked"),
			"The number of operations in deadlock.",
			nil,
			nil),
	}
}

func (c *MonitorCollecter) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.entriesSent
	ch <- c.currentConnections
	ch <- c.currentWorkQueueDepth
	ch <- c.idleConnectionsClosed
	ch <- c.autoConnectionCleanerRun
	ch <- c.totalConnections
	ch <- c.workerThreads
	ch <- c.operationsRequested
	ch <- c.operationsCompleted
	ch <- c.operationsFromSuppliers
	ch <- c.operationsWaiting
	ch <- c.operationsRetried
	ch <- c.operationsDeadlocked
}

func (c *MonitorCollecter) Collect(ch chan<- prometheus.Metric) {
	q := ldap.NewSearchRequest(
		"cn=monitor",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{"*"}, nil,
	)
	p, err := c.exporter.ldapConn.Search(q)
	if err != nil {
		c.exporter.logger.Error("Error querying cn=monitor", "err", err)
		return
	}
	x := p.Entries[0]

	ch <- prometheus.MustNewConstMetric(c.entriesSent, prometheus.CounterValue, attr(x, "entriessent"))
	ch <- prometheus.MustNewConstMetric(c.currentConnections, prometheus.GaugeValue, attr(x, "currentconnections"))
	ch <- prometheus.MustNewConstMetric(c.currentWorkQueueDepth, prometheus.GaugeValue, attr(x, "current_workqueue_size"))
	ch <- prometheus.MustNewConstMetric(c.idleConnectionsClosed, prometheus.GaugeValue, attr(x, "idle_connections_closed"))
	ch <- prometheus.MustNewConstMetric(c.autoConnectionCleanerRun, prometheus.GaugeValue, attr(x, "auto_connection_cleaner_run"))

	s := attr(x, "total_ssl_connections")
	t := attr(x, "total_tls_connections")
	v := attr(x, "totalconnections") - s - t
	ch <- prometheus.MustNewConstMetric(c.totalConnections, prometheus.CounterValue, s, "ssl")
	ch <- prometheus.MustNewConstMetric(c.totalConnections, prometheus.CounterValue, t, "tls")
	ch <- prometheus.MustNewConstMetric(c.totalConnections, prometheus.CounterValue, v, "tcp")

	for k, v := range map[string]string{
		"write": "writewaiters",
		"read":  "readwaiters",
		"live":  "livethreads",
		"idle":  "available_workers",
	} {
		ch <- prometheus.MustNewConstMetric(c.workerThreads, prometheus.GaugeValue, attr(x, v), k)
	}
	for _, o := range []string{"search", "bind", "unbind", "add", "delete", "modrdn", "modify", "compare", "abandon", "extop", "unknownop"} {
		v = attr(x, plural(o)+"requested")
		ch <- prometheus.MustNewConstMetric(c.operationsRequested, prometheus.CounterValue, v, o)
		v = attr(x, plural(o)+"completed")
		ch <- prometheus.MustNewConstMetric(c.operationsCompleted, prometheus.CounterValue, v, o)
	}
	for _, o := range []string{"add", "delete", "modrdn", "modify"} {
		v = attr(x, plural(o)+"fromsuppliers")
		ch <- prometheus.MustNewConstMetric(c.operationsFromSuppliers, prometheus.CounterValue, v, o)
	}
	ch <- prometheus.MustNewConstMetric(c.operationsWaiting, prometheus.GaugeValue, attr(x, "operations_waiting"))
	ch <- prometheus.MustNewConstMetric(c.operationsRetried, prometheus.CounterValue, attr(x, "operations_retried"))
	ch <- prometheus.MustNewConstMetric(c.operationsDeadlocked, prometheus.GaugeValue, attr(x, "operations_deadlocked"))
}

func plural(o string) string {
	switch o {
	case "search":
		{
			return "searches"
		}
	case "modify":
		{
			return "modifies"
		}
	default:
		{
			return o + "s"
		}
	}
}

func attr(x *ldap.Entry, a string) float64 {
	v := x.GetAttributeValue(a)
	if f, err := strconv.ParseFloat(v, 64); err != nil {
		return 0
	} else {
		return f
	}
}
