package collector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/prometheus/client_golang/prometheus"
)

// entryDN: cn=goldentooth02,cn=goldentooth01,ibm-replicaGroup=default,dc=avnet,d
//  c=com
// ibm-capabilitiessubentry: cn=ibm-capabilities,dc=avnet,dc=com
// ibm-replicationThisServerIsMaster: TRUE

// ibm-replicationIsQuiesced: FALSE

// ibm-replicationLastResult: N/A
// ibm-replicationLastResultAdditional: N/A
// ibm-replicationNextTime: N/A

type ReplicationCollecter struct {
	exporter                  *Exporter
	state                     *prometheus.Desc
	lastActivation            *prometheus.Desc
	lastFinish                *prometheus.Desc
	lastChangeId              *prometheus.Desc
	pendingChanges            *prometheus.Desc
	failedChanges             *prometheus.Desc
	perfQueueSizeLimit        *prometheus.Desc
	perfLastOperationId       *prometheus.Desc
	perfSendQueueSize         *prometheus.Desc
	perfDependentUpdates      *prometheus.Desc
	perfSendQueueLimitHits    *prometheus.Desc
	perfDependentUpdatesSent  *prometheus.Desc
	perfSendQueueWaited       *prometheus.Desc
	perfReceiveQueueLimitHits *prometheus.Desc
	perfUpdatesAcknowledged   *prometheus.Desc
	perfUpdatesSent           *prometheus.Desc
	perfErrorsReported        *prometheus.Desc
	perfSenderSessions        *prometheus.Desc
	perfReceiverSessions      *prometheus.Desc
}

func NewReplicationCollecter(e *Exporter) *ReplicationCollecter {
	return &ReplicationCollecter{
		exporter: e,
		state: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "state"),
			"The current state of replication with this consumer.",
			[]string{"consumer", "state"},
			nil),
		lastActivation: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "last_activation_seconds"),
			"The time that the last replication session started between this supplier and consumer.",
			[]string{"consumer"},
			nil),
		lastFinish: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "last_finish_seconds"),
			"The time that the last replication session finished between this supplier and consumer.",
			[]string{"consumer"},
			nil),
		lastChangeId: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "last_change_id"),
			"The change ID of the last update sent to this consumer.",
			[]string{"consumer"},
			nil),
		pendingChanges: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "pending_changes"),
			"The number of updates queued to be replicated to this consumer.",
			[]string{"consumer"},
			nil),
		failedChanges: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication", "failed_changes"),
			"The count of the failures logged for this replication agreement.",
			[]string{"consumer"},
			nil),

		perfQueueSizeLimit: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "queue_size_limit"),
			"This is the size limit for each queue.",
			[]string{"consumer", "connection"},
			nil),
		perfLastOperationId: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "last_operation_id"),
			"The replication ID of the last operation assigned to the send queue of the connection.",
			[]string{"consumer", "connection"},
			nil),
		perfSendQueueSize: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "send_queue_size"),
			"The current size (number of operations) of the send queue.",
			[]string{"consumer", "connection"},
			nil),
		perfDependentUpdates: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "dependent_updates"),
			"The count of dependent updates.",
			[]string{"consumer", "connection"},
			nil),
		perfSendQueueLimitHits: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "send_queue_limit_hits"),
			"The number of times the send queue hit the size limit.",
			[]string{"consumer", "connection"},
			nil),
		perfDependentUpdatesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "dependent_updates_sent"),
			"The number of dependent updates sent.",
			[]string{"consumer", "connection"},
			nil),
		perfSendQueueWaited: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "send_queue_waited"),
			"The number of times the send queue waited for a dependent update before sending additional updates.",
			[]string{"consumer", "connection"},
			nil),
		perfReceiveQueueLimitHits: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "receive_queue_limit_hits"),
			"The number of times the receive queue hit the size limit.",
			[]string{"consumer", "connection"},
			nil),
		perfUpdatesAcknowledged: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "updates_acknowledged"),
			"The number of updates where results have been received.",
			[]string{"consumer", "connection"},
			nil),
		perfUpdatesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "updates_sent"),
			"The number of updates sent to a consumer since start-up.",
			[]string{"consumer", "connection"},
			nil),
		perfErrorsReported: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "errors_reported"),
			"The number of replication errors reported by the consumer.",
			[]string{"consumer", "connection"},
			nil),
		perfSenderSessions: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "sender_sessions"),
			"The session count for the sender thread (incremented when the connection to the consumer is established).",
			[]string{"consumer", "connection"},
			nil),
		perfReceiverSessions: prometheus.NewDesc(
			prometheus.BuildFQName("ibmslapd", "replication_performance", "receiver_sessions"),
			"The session count for the receiver thread.",
			[]string{"consumer", "connection"},
			nil),
	}
}

func (c *ReplicationCollecter) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.state
	ch <- c.lastActivation
	ch <- c.lastFinish
	ch <- c.lastChangeId
	ch <- c.pendingChanges
	ch <- c.failedChanges

	ch <- c.perfQueueSizeLimit
	ch <- c.perfLastOperationId
	ch <- c.perfSendQueueSize
	ch <- c.perfDependentUpdates
	ch <- c.perfSendQueueLimitHits
	ch <- c.perfDependentUpdatesSent
	ch <- c.perfSendQueueWaited
	ch <- c.perfReceiveQueueLimitHits
	ch <- c.perfUpdatesAcknowledged
	ch <- c.perfUpdatesSent
	ch <- c.perfErrorsReported
	ch <- c.perfSenderSessions
	ch <- c.perfReceiverSessions
}

func (c *ReplicationCollecter) Collect(ch chan<- prometheus.Metric) {
	q := ldap.NewSearchRequest(
		"",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=ibm-replicationAgreement)", []string{"cn", "++ibmrepl"}, nil,
	)
	p, err := c.exporter.ldapConn.Search(q)
	if err != nil {
		c.exporter.logger.Error("Error querying replication agreements", "err", err)
		return
	}
	for _, x := range p.Entries {
		state := x.GetAttributeValue("ibm-replicationState")
		if state == "" {
			continue
		}
		consumer := x.GetAttributeValue("cn")
		for _, v := range [...]string{"active", "ready", "retrying", "waiting", "binding", "connecting", "on hold", "error log full"} {
			if v == state {
				ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 1, consumer, v)
			} else {
				ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 0, consumer, v)
			}
		}
		var error int
		if n, _ := fmt.Sscanf(state, "error %d", &error); n == 1 {
			ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 1, consumer, "error")
		} else {
			ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 0, consumer, "error")
		}

		if t, err := time.Parse("20060102150405Z", x.GetAttributeValue("ibm-replicationLastActivationTime")); err == nil {
			ch <- prometheus.MustNewConstMetric(c.lastActivation, prometheus.CounterValue, float64(t.Unix()), consumer)
		}
		if t, err := time.Parse("20060102150405Z", x.GetAttributeValue("ibm-replicationLastFinishTime")); err == nil {
			ch <- prometheus.MustNewConstMetric(c.lastFinish, prometheus.CounterValue, float64(t.Unix()), consumer)
		}
		ch <- prometheus.MustNewConstMetric(c.lastChangeId, prometheus.CounterValue, attr(x, "ibm-replicationLastChangeId"), consumer)
		ch <- prometheus.MustNewConstMetric(c.pendingChanges, prometheus.GaugeValue, attr(x, "ibm-replicationPendingChangeCount"), consumer)
		ch <- prometheus.MustNewConstMetric(c.failedChanges, prometheus.GaugeValue, attr(x, "ibm-replicationFailedChangeCount"), consumer)

		// [c=0,l=10,op=3056,q=438,d=7,ws=0,s=438,ds=7,wd=0,wr=0,r=438,e=16,ss=1,rs=1]
		for _, p := range x.GetAttributeValues("ibm-replicationperformance") {
			var n int
			var l, op, q, d, ws, s, ds, wd, wr, r, e, ss, rs float64
			fmt.Sscanf(p,
				"[c=%v,l=%v,op=%v,q=%v,d=%v,ws=%v,s=%v,ds=%v,wd=%v,wr=%v,r=%v,e=%v,ss=%v,rs=%v]",
				&n, &l, &op, &q, &d, &ws, &s, &ds, &wd, &wr, &r, &e, &ss, &rs)
			ch <- prometheus.MustNewConstMetric(c.perfQueueSizeLimit, prometheus.GaugeValue,
				l, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfLastOperationId, prometheus.GaugeValue,
				op, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfSendQueueSize, prometheus.GaugeValue,
				q, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfDependentUpdates, prometheus.GaugeValue,
				d, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfSendQueueLimitHits, prometheus.GaugeValue,
				ws, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfDependentUpdatesSent, prometheus.GaugeValue,
				s, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfSendQueueWaited, prometheus.GaugeValue,
				ds, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfReceiveQueueLimitHits, prometheus.GaugeValue,
				wd, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfUpdatesAcknowledged, prometheus.GaugeValue,
				wr, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfUpdatesSent, prometheus.GaugeValue,
				r, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfErrorsReported, prometheus.GaugeValue,
				e, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfSenderSessions, prometheus.GaugeValue,
				ss, consumer, strconv.Itoa(n))
			ch <- prometheus.MustNewConstMetric(c.perfReceiverSessions, prometheus.GaugeValue,
				rs, consumer, strconv.Itoa(n))
		}
	}
}
