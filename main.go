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
		log.Fatalf("Cannot convert PROMETHEUS_API_TIMEOUT into an int: %v\n", err)
	}
	prometheusApiTimeout := time.Duration(prometheusApiTimeout_i)

	// the alert name whose status we are interested in
	prometheusAlertName := os.Getenv("PROMETHEUS_ALERT_NAME")
	if prometheusAlertName == "" {
		log.Fatalln("Missing required parameter: PROMETHEUS_ALERT_NAME")
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
		log.Fatalf("Error creating Prometheus client: %v\n", err)
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

		// note that alertsResult.Alerts only contains active alerts, not all alerts.
		// but we're not interested in inactive alerts anyway.
		for _, alert := range alertsResult.Alerts {
			// scan until we reach the alert we're interested in
			if string(alert.Labels[model.AlertNameLabel]) != prometheusAlertName {
				continue
			}

			if string(alert.State) == "firing" {
				errMsg := fmt.Sprintf("The Prometheus alert is firing: %v", alert.Labels)
				log.Println("ERROR: " + errMsg)
				notReadyResponse(writer, request, errors.New(errMsg))
				return
			}
		}

		// if there are no issues, then report readiness
		readyResponse(writer, request)
	})

	log.Print("Starting HTTP listener...")
	log.Fatal(http.ListenAndServe(":"+kubeProbeListenPort, nil))
}
