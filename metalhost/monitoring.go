package metalhost

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// MonitoringURLs are project-scoped authenticated HTTP endpoints, not Connect RPCs.
type MonitoringURLs struct {
	Prometheus    string
	Metrics       string
	ScrapeTargets string
}

var projectComponent = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

// MonitoringEndpoints validates a canonical project name and an API origin.
// Plain HTTP is supported only on loopback for local development. A returned URL
// does not grant access: use an authorized project-scoped monitoring.read key.
func (c Config) MonitoringEndpoints(project string) (MonitoringURLs, error) {
	endpoint, err := url.Parse(c.BaseURL())
	if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.ForceQuery || endpoint.Fragment != "" || (endpoint.Path != "" && endpoint.Path != "/") {
		return MonitoringURLs{}, errors.New("monitoring requires an API origin without credentials, path, query, or fragment")
	}
	loopback := endpoint.Hostname() == "localhost" || endpoint.Hostname() == "127.0.0.1" || endpoint.Hostname() == "::1"
	if endpoint.Scheme != "https" && !(endpoint.Scheme == "http" && loopback) {
		return MonitoringURLs{}, errors.New("monitoring requires HTTPS, except for a local loopback API")
	}
	parts := strings.Split(project, "/")
	valid := len(parts) == 2 && parts[0] == "projects" && projectComponent.MatchString(parts[1])
	valid = valid || (len(parts) == 4 && parts[0] == "organizations" && parts[2] == "projects" && projectComponent.MatchString(parts[1]) && projectComponent.MatchString(parts[3]))
	if !valid {
		return MonitoringURLs{}, errors.New("canonical projects/{id} or organizations/{org}/projects/{id} name required")
	}
	endpoint.Path = "/v1/monitoring/" + project
	base := endpoint.String()
	return MonitoringURLs{Prometheus: base + "/prometheus", Metrics: base + "/metrics", ScrapeTargets: base + "/scrape-targets"}, nil
}
