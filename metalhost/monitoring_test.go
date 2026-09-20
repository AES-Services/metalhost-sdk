package metalhost

import "testing"

func TestMonitoringEndpoints(t *testing.T) {
	for _, project := range []string{"projects/example", "organizations/example/projects/default"} {
		urls, err := (Config{Endpoint: " https://api.metalhost.net/ "}).MonitoringEndpoints(project)
		if err != nil || urls.Prometheus != "https://api.metalhost.net/v1/monitoring/"+project+"/prometheus" || urls.Metrics != "https://api.metalhost.net/v1/monitoring/"+project+"/metrics" || urls.ScrapeTargets != "https://api.metalhost.net/v1/monitoring/"+project+"/scrape-targets" {
			t.Fatalf("URLs: %+v %v", urls, err)
		}
	}
	for _, project := range []string{"example", "projects/..", "projects/%2e%2e", "projects/a/b", "projects/a?tenant=b", "organizations/a/projects/", "projects/a\n"} {
		if _, err := (Config{Endpoint: "https://api.metalhost.net"}).MonitoringEndpoints(project); err == nil {
			t.Fatalf("accepted invalid project %q", project)
		}
	}
	for _, endpoint := range []string{"http://api.metalhost.net", "https://key@api.metalhost.net", "https://api.metalhost.net?key=x", "https://api.metalhost.net#token", "https://api.metalhost.net/path", "ftp://localhost", "https://api.metalhost.net?"} {
		if _, err := (Config{Endpoint: endpoint}).MonitoringEndpoints("projects/a"); err == nil {
			t.Fatalf("accepted invalid API origin %q", endpoint)
		}
	}
	for _, endpoint := range []string{"http://127.0.0.1:8080", "http://[::1]:8080", "http://localhost:8080"} {
		if _, err := (Config{Endpoint: endpoint}).MonitoringEndpoints("projects/a"); err != nil {
			t.Fatal(err)
		}
	}
}
