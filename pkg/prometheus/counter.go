package prometheus

import (
	"github.com/golang/glog"
	"github.com/nirmata/kyverno/pkg/client/listers/policy/v1alpha1"
	"github.com/nirmata/kyverno/pkg/webhooks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	labels "k8s.io/apimachinery/pkg/labels"
)

type MetricsManager struct {
	policyLister v1alpha1.PolicyLister
}
type PrometheusMetrics struct {
	Prefix      string
	AuditPolicy prometheus.Gauge
}

func InitPrometheusMetrics(prefix string) PrometheusMetrics {
	pm := PrometheusMetrics{
		Prefix: prefix,
		AuditPolicy: promauto.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_audit_policy",
			Help: "Number of audit policies",
		}),
	}

	return pm
}

func NewMetricsManager(policyLister v1alpha1.PolicyLister) MetricsManager {
	return MetricsManager{
		policyLister: policyLister,
	}
}

func (m MetricsManager) CountAuditPolicy() float64 {
	var counter float64

	policies, err := m.policyLister.List(labels.NewSelector())
	if err != nil {
		glog.Errorf("Failed to get policy list, err %v\n", err)
		return 0
	}

	for _, p := range policies {
		if p.Spec.ValidationFailureAction == webhooks.ReportViolation {
			// TODO: change to thread safe counter
			counter++
		}
	}

	glog.V(3).Infof("Found %v audit policies", counter)
	return counter
}
