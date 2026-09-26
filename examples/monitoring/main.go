// Read-only monitoring example; requires an explicitly enabled API environment.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	monitoringv1 "github.com/AES-Services/metalhost-sdk/gen/go/aes/monitoring/v1"
	"github.com/AES-Services/metalhost-sdk/gen/go/aes/monitoring/v1/monitoringv1connect"
	"github.com/AES-Services/metalhost-sdk/metalhost"
)

func main() {
	cfg := metalhost.Config{Endpoint: os.Getenv("METALHOST_ENDPOINT"), APIKey: os.Getenv("METALHOST_API_KEY")}
	vm := os.Getenv("METALHOST_VM")
	if cfg.Endpoint == "" || cfg.APIKey == "" || vm == "" {
		log.Fatal("set METALHOST_ENDPOINT, METALHOST_API_KEY and the full METALHOST_VM resource name")
	}
	cfg.HTTPClient = &http.Client{
		Transport: cfg.RoundTripper(http.DefaultTransport),
		Timeout:   30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	client := monitoringv1connect.NewMonitoringServiceClient(cfg.Client(), cfg.BaseURL())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	end := time.Now().Unix()
	result, err := client.QueryVMMonitoring(ctx, connect.NewRequest(&monitoringv1.QueryVMMonitoringRequest{
		Name: vm, MetricIds: []string{"cpu_utilization_pct", "memory_used_bytes"},
		StartTimeUnix: end - 3600, EndTimeUnix: end, StepSeconds: 60,
	}))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bounds=%d..%d step=%ds\n", result.Msg.StartTimeUnix, result.Msg.EndTimeUnix, result.Msg.StepSeconds)
	for _, quality := range result.Msg.Quality {
		fmt.Printf("%s: %s observed=%d\n", quality.MetricId, quality.Status, quality.ObservedAtUnix)
	}
	for _, series := range result.Msg.Series {
		fmt.Printf("%s dimensions=%v samples=%d\n", series.MetricId, series.Dimensions, len(series.Samples))
	}
}
