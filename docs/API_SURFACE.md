# API surface

The SDK's public Go surface has two layers:

- `metalhost.Config` normalizes an endpoint, selects an HTTP client, and creates an authentication/user-agent `http.RoundTripper`.
- `gen/go/aes/<service>/v1` contains protobuf messages; the adjacent `<service>v1connect` package contains generated Connect client interfaces, constructors, and procedure constants.

## Generated services

This checkout's snapshot includes the services below. The September entries
remain coordinated-release previews; this list is not a claim that every
published SDK version or production endpoint exposes them.

- audit
- bare metal
- catalog
- compute and SSH keys
- health
- IAM
- scoped automation credentials, project service accounts, and GitHub workload identity
- monitoring metrics, alert rules, verified destinations, incidents, and delivery tests
- network
- operations
- projects and organizations
- quota
- storage
- support
- wallet and billing
- webhooks

Use the generated constructor for the service you need:

```go
cfg := metalhost.Config{
	Endpoint: "https://api.metalhost.net",
	APIKey:   os.Getenv("METALHOST_API_KEY"),
	UserAgent: "example/1.0",
}
cfg.HTTPClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: cfg.RoundTripper(http.DefaultTransport),
}

client := computev1connect.NewComputeServiceClient(cfg.Client(), cfg.BaseURL())
response, err := client.ListVirtualMachines(
	ctx,
	connect.NewRequest(&computev1.ListVirtualMachinesRequest{
		ProjectName: "projects/example",
	}),
)
```

Imports for that example are `connectrpc.com/connect`, `github.com/AES-Services/metalhost-sdk/gen/go/aes/compute/v1`, `github.com/AES-Services/metalhost-sdk/gen/go/aes/compute/v1/computev1connect`, and `github.com/AES-Services/metalhost-sdk/metalhost`, plus the standard library.

## Transport behavior

`Config.RoundTripper(base)` wraps `base`, or `http.DefaultTransport` when `base` is nil. It clones outgoing requests, overwrites `Authorization` with a Bearer API key when configured, and sets a user agent only if the request has none. The default user agent is `metalhost-sdk-go`.

`Config.Client()` returns `Config.HTTPClient` unchanged when supplied. Otherwise it creates an `http.Client` with a 30-second timeout; it does not attach `RoundTripper`. Callers therefore must install the wrapper as shown above when they need authentication.

## Other representations

- `proto/aes/**` is the source API snapshot.
- `gen/openapi/metalhost.openapi.yaml` is the generated HTTP/OpenAPI representation of the same public RPCs.

Generated code is part of each tagged release but should not be edited by hand.

## Backups (v1.1.0)

Compute exposes whole-VM backup capture and restore, including captured disk
manifests and consistency. Storage exposes single-disk backup CRUD, disk restore
via `from_snapshot`, and automatic backup schedules with keep-N retention.
See [Backups API](BACKUPS.md) for retry, billing, timing, and restore semantics.

## September observability (unreleased)

The new `AutomationService`, `MonitoringService`, and `AlertService` contracts
require the coordinated September backend release with automation and monitoring
enabled. Existing IAM keys and sessions are unchanged.

Creation/rotation requests accept a client UUID `request_id`. Retain the exact
request on an ambiguous retry. Credential retries return the same credential's
current metadata with `secret_unavailable=true`; no secret is stored for replay.
Revoke a recovered key whose original secret was lost before creating another.

Rule, destination, and incident mutations also carry a stable request ID and
optimistic revision where applicable. A successful save is not confirmation that
the native evaluator has applied it; inspect desired/applied revision and health.
Queued delivery is not provider acceptance, and provider acceptance is not proof
that a recipient read a notification. Missing data is never healthy zero usage.

Hosted PromQL and Prometheus scrape endpoints are ordinary authenticated HTTP
interfaces, not Connect RPCs. Use a project-scoped `monitoring.read` credential;
the server derives immutable tenant identity and does not trust tenant headers.

See [Observability](OBSERVABILITY.md) for runnable examples and explicit CLI/API
boundaries. Workflow identity does not provision GitHub runners.
