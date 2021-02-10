package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	// where to query for the alert status
	prometheusEndpoint := os.Getenv("PROMETHEUS_ENDPOINT")
	if prometheusEndpoint == "" {
		prometheusEndpoint = "http://localhost:9090"
	}

	// timeout in case Prometheus does not respond quickly enough
	prometheusApiTimeout_s := os.Getenv("PROMETHEUS_API_TIMEOUT")
	if prometheusApiTimeout_s == "" {
		prometheusApiTimeout_s = "10"
	}
	prometheusApiTimeout_i, err := strconv.Atoi(prometheusApiTimeout_s)
	if err != nil {
		fmt.Printf("Cannot convert PROMETHEUS_API_TIMEOUT into an int: %v\n", err)
		os.Exit(1)
	}
	prometheusApiTimeout := time.Duration(prometheusApiTimeout_i)

	// the alert name whose status we are interested in
	prometheusAlertName := os.Getenv("PROMETHEUS_ALERT_NAME")
	if prometheusAlertName == "" {
		fmt.Printf("Missing required parameter: PROMETHEUS_ALERT_NAME")
		os.Exit(1)
	}

	// the path for the liveness check
	kubernetesLivenessPath := os.Getenv("KUBE_LIVENESS_PATH")
	if kubernetesLivenessPath == "" {
		kubernetesLivenessPath = "/live"
	}

	// the path for the readiness check
	kubernetesReadinessPath := os.Getenv("KUBE_READINESS_PATH")
	if kubernetesReadinessPath == "" {
		kubernetesReadinessPath = "/ready"
	}

	// the port number on which the liveness/readiness paths will listen
	kubeProbeListenPort := os.Getenv("KUBE_PROBE_LISTEN_PORT")
	if kubeProbeListenPort == "" {
		kubeProbeListenPort = "8080"
	}

	// initialize prometheus client
	client, err := api.NewClient(api.Config{
		Address: prometheusEndpoint,
	})
	if err != nil {
		fmt.Printf("Error creating Prometheus client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)

	readyResponse := func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok\n"))
	}

	notReadyResponse := func(writer http.ResponseWriter, request *http.Request, err error) {
		writer.Header().Add("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusServiceUnavailable)
		writer.Write([]byte(fmt.Sprintf("not ok, err:\n%v\n", err)))
	}

	// register at live path
	http.HandleFunc(kubernetesLivenessPath, func(writer http.ResponseWriter, request *http.Request) {
		readyResponse(writer, request)
	})

	// register at ready path
	http.HandleFunc(kubernetesReadinessPath, func(writer http.ResponseWriter, request *http.Request) {
		// query the prometheus endpoint with the query
		ctx, cancel := context.WithTimeout(request.Context(), prometheusApiTimeout*time.Second)
		defer cancel()

		alertsResult, err := v1api.Alerts(ctx)
		if err != nil {
			notReadyResponse(writer, request, err)
			return
		}
		alertFound := false
		for _, alert := range alertsResult.Alerts {
			// scan until we reach the alert we're interested in
			if string(alert.Labels[model.AlertNameLabel]) != prometheusAlertName {
				continue
			}
			alertFound = true

			if string(alert.State) == "firing" {
				notReadyResponse(writer, request, errors.New("The Prometheus alert is firing!"))
				return
			}
		}

		if !alertFound {
			panic("Configured Prometheus alert was not found on the Prometheus endpoint!")
		}

		// if there are no issues, then report readiness
		readyResponse(writer, request)
	})

	log.Fatal(http.ListenAndServe(":"+kubeProbeListenPort, nil))
}
