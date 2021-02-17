package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coralogix/prometheus-alerts-readiness/internal/config"
	"github.com/coralogix/prometheus-alerts-readiness/internal/responses"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"log"
	"net/http"
	"time"
)

func main() {
	c, err := config.New()
	if err != nil {
		log.Fatalf("Error reading Configuration: %v\n", err)
	}

	// initialize prometheus client
	client, err := api.NewClient(api.Config{
		Address: c.PrometheusEndpoint,
	})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v\n", err)
	}

	v1api := v1.NewAPI(client)

	// register at live path
	http.HandleFunc(c.KubernetesLivenessPath, func(writer http.ResponseWriter, request *http.Request) {
		responses.Ready(writer, request)
	})

	// register at ready path
	http.HandleFunc(c.KubernetesReadinessPath, func(writer http.ResponseWriter, request *http.Request) {
		// query the prometheus endpoint with the query
		ctx, cancel := context.WithTimeout(request.Context(), c.PrometheusApiTimeout*time.Second)
		defer cancel()

		alertsResult, err := v1api.Alerts(ctx)
		if err != nil {
			responses.NotReady(writer, request, err)
			return
		}

		// note that alertsResult.Alerts only contains active alerts, not all alerts.
		// but we're not interested in inactive alerts anyway.
		for _, alert := range alertsResult.Alerts {
			// scan until we reach the alert we're interested in
			severity := string(alert.Labels["severity"])
			severityIsRelevant := false
			for _, configuredSeverity := range c.PrometheusAlertSeverities {
				if severity == configuredSeverity {
					severityIsRelevant = true
					break
				}
			}

			if !severityIsRelevant {
				continue
			}

			if alert.State == v1.AlertStateFiring {
				errMsg := fmt.Sprintf("The Prometheus alert is firing: %v", alert.Labels)
				log.Println("ERROR: " + errMsg)
				responses.NotReady(writer, request, errors.New(errMsg))
				return
			}
		}

		// if there are no issues, then report readiness
		responses.Ready(writer, request)
	})

	log.Print("Starting HTTP listener...")
	log.Fatal(http.ListenAndServe(":"+c.KubeProbeListenPort, nil))
}
