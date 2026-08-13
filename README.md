# Metalhost SDK

Public Go SDK and versioned API contract snapshots for AES Metalhost.

The module is `github.com/AES-Services/metalhost-sdk`. Its first stable release is `v1.0.0`; consumers should select a concrete `v1` version in `go.mod`.

## What is included

```text
proto/aes/       Released protobuf API snapshot
gen/go/          Generated protobuf messages and Connect RPC clients
gen/openapi/     Generated and finalized OpenAPI document
metalhost/       Small hand-written HTTP configuration helper
docs/            API surface, consumer, and release notes
scripts/         API snapshot synchronization
```

The checked-in proto files are the source contract for a release. `buf generate` derives the Go and OpenAPI outputs from that snapshot. See [API surface](docs/API_SURFACE.md) for packages and a minimal client example.

## Go client setup

`metalhost.Config` contains the endpoint, API key, optional HTTP client, and user agent. `BaseURL` trims whitespace and a trailing slash. `Client` returns the supplied HTTP client or a new client with a 30-second timeout.

Authentication is installed explicitly as an HTTP transport:

```go
package main

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	healthv1 "github.com/AES-Services/metalhost-sdk/gen/go/aes/health/v1"
	"github.com/AES-Services/metalhost-sdk/gen/go/aes/health/v1/healthv1connect"
	"github.com/AES-Services/metalhost-sdk/metalhost"
)

func check(ctx context.Context) error {
	cfg := metalhost.Config{
		Endpoint: "https://api.metalhost.net",
		APIKey:   "aes_...",
		UserAgent: "my-integration/1.0",
	}
	cfg.HTTPClient = &http.Client{Transport: cfg.RoundTripper(http.DefaultTransport)}

	client := healthv1connect.NewHealthServiceClient(cfg.Client(), cfg.BaseURL())
	_, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))
	return err
}
```

`RoundTripper` clones each request, sets `Authorization: Bearer <API key>` when a key is present, and supplies the configured user agent only when the request does not already have one. `Config.Client()` does **not** install that transport automatically.

Generated service constructors and request/response types remain the primary SDK surface; the helper does not hide authorization, validation, pagination, operations, or billing behavior.

## Keep the API snapshot synchronized

Install pinned generation tools once:

```sh
make tools
```

To replace the public snapshot from the API source repository:

```sh
METALHOST_API_SOURCE="/path/to/api/source" ./scripts/sync-api-snapshot.sh
```

The script copies the allowlisted public `proto/aes` packages, regenerates Go and OpenAPI output, finalizes the OpenAPI metadata/filtering, runs `go mod tidy`, and tests the module. It requires `yq`.

For an existing snapshot, `make ci` runs proto lint, generation/OpenAPI finalization, and Go tests. CI also fails if regeneration changes `gen/`.

## Consumers and releases

- The public `metalhost` CLI uses the generated Connect clients and `metalhost.Config`.
- A Terraform provider is planned but is not published from this repository.
- Release and compatibility rules are in [Release process](docs/RELEASE_PROCESS.md).
- CLI and Terraform boundaries are in [CLI and Terraform consumers](docs/CLI_AND_TERRAFORM.md).
