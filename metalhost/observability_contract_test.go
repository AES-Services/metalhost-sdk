package metalhost

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	iamv1 "github.com/AES-Services/metalhost-sdk/gen/go/aes/iam/v1"
	monitoringv1 "github.com/AES-Services/metalhost-sdk/gen/go/aes/monitoring/v1"
	"github.com/AES-Services/metalhost-sdk/gen/go/aes/monitoring/v1/monitoringv1connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestSeptemberFollowupContracts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request proto.Message
		field   string
		want    any
	}{
		{"pause", &monitoringv1.SetEnhancedMonitoringPausedRequest{Name: "virtual-machines/example", InstallationId: "installation", Paused: true, RequestId: "request"}, "paused", true},
		{"project credentials", &iamv1.ListCredentialsRequest{ProjectName: "projects/example", AllProjectCredentials: true}, "allProjectCredentials", true},
		{"chat destination", &monitoringv1.SaveAlertDestinationRequest{WebhookUrl: "https://example.invalid/test-only"}, "webhookUrl", "https://example.invalid/test-only"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := protojson.Marshal(tc.request)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]any
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if fields[tc.field] != tc.want {
				t.Fatalf("missing or changed %s", tc.field)
			}
			roundtrip := tc.request.ProtoReflect().New().Interface()
			if err := protojson.Unmarshal(raw, roundtrip); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(tc.request, roundtrip) {
				t.Fatal("request did not round-trip")
			}
		})
	}
	const procedure = "/aes.monitoring.v1.MonitoringService/SetEnhancedMonitoringPaused"
	if monitoringv1connect.MonitoringServiceSetEnhancedMonitoringPausedProcedure != procedure {
		t.Fatal("procedure changed")
	}
	spec, err := os.ReadFile("../gen/openapi/metalhost.openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{procedure, "allProjectCredentials", "webhookUrl"} {
		if !strings.Contains(string(spec), field) {
			t.Fatalf("OpenAPI missing %s", field)
		}
	}
}
