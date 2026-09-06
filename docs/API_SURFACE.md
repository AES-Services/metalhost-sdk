# API surface

The SDK's public Go surface has two layers:

- `metalhost.Config` normalizes an endpoint, selects an HTTP client, and creates an authentication/user-agent `http.RoundTripper`.
- `gen/go/aes/<service>/v1` contains protobuf messages; the adjacent `<service>v1connect` package contains generated Connect client interfaces, constructors, and procedure constants.

## Generated services

The released snapshot includes:

- audit
- bare metal
- catalog
- compute and SSH keys
- health
- IAM
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
