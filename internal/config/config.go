package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// where to query for the alert status
	PrometheusEndpoint string

	// timeout in case Prometheus does not respond quickly enough
	PrometheusApiTimeout time.Duration

	// for a given alert with the label key "severity", cause a NotReady response if the
	// label value is one of the values in this slice.
	PrometheusAlertSeverities []string

	// the path for the liveness check
	KubernetesLivenessPath string

	// the path for the readiness check
	KubernetesReadinessPath string

	// the port number on which the liveness/readiness paths will listen
	KubeProbeListenPort string
}

// Instantiates a Configuration native-Go struct from raw external configuration sources.
func New() (*Config, error) {
	c := new(Config)

	c.PrometheusEndpoint = os.Getenv("PROMETHEUS_ENDPOINT")
	if c.PrometheusEndpoint == "" {
		c.PrometheusEndpoint = "http://localhost:9090"
	}

	prometheusApiTimeoutS := os.Getenv("PROMETHEUS_API_TIMEOUT")
	if prometheusApiTimeoutS == "" {
		prometheusApiTimeoutS = "10"
	}
	prometheusApiTimeoutI, err := strconv.Atoi(prometheusApiTimeoutS)
	if err != nil {
		log.Printf("Cannot convert PROMETHEUS_API_TIMEOUT into an int: %v\n", err)
		return nil, err
	}
	c.PrometheusApiTimeout = time.Duration(prometheusApiTimeoutI)

	prometheusAlertSeveritiesCSV := os.Getenv("PROMETHEUS_ALERT_SEVERITIES")
	if prometheusAlertSeveritiesCSV == "" {
		prometheusAlertSeveritiesCSV = "critical,warning"
	}
	c.PrometheusAlertSeverities = strings.Split(prometheusAlertSeveritiesCSV, ",")

	c.KubernetesLivenessPath = os.Getenv("KUBE_LIVENESS_PATH")
	if c.KubernetesLivenessPath == "" {
		c.KubernetesLivenessPath = "/live"
	}

	c.KubernetesReadinessPath = os.Getenv("KUBE_READINESS_PATH")
	if c.KubernetesReadinessPath == "" {
		c.KubernetesReadinessPath = "/ready"
	}

	c.KubeProbeListenPort = os.Getenv("KUBE_PROBE_LISTEN_PORT")
	if c.KubeProbeListenPort == "" {
		c.KubeProbeListenPort = "8080"
	}

	return c, nil
}
