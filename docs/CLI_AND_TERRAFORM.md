# CLI and Terraform consumers

The released SDK is the shared contract for customer tools. Consumers import generated request/response types and Connect clients from `gen/go/aes/...`; `metalhost.Config` only supplies endpoint normalization and an opt-in authentication/user-agent transport.

## Public CLI

The current public CLI imports the generated clients directly. It builds an `http.Client` whose transport is `metalhost.Config.RoundTripper(http.DefaultTransport)`, then passes that client and `Config.BaseURL()` to each generated `New<Service>Client` constructor.

CLI behavior belongs in the CLI repository, including:

- profiles and environment/flag precedence;
- interactive email/password, OIDC, MFA, and API-key authentication;
- pagination and `--all`;
- operation polling and output formatting;
- prompts, scripting exit codes, and declarative `apply`.

These are not generic helpers in the SDK today. In particular, the SDK does not currently provide pagination, operation waiting, idempotency-key, typed-error, or presigned-transfer abstractions. Connect errors and generated messages remain visible to callers.

## Terraform provider (planned)

A Metalhost Terraform provider is planned; this repository does not currently contain or publish one.

When implemented, the provider should:

- depend on a tagged `v1` SDK release rather than copying proto types;
- construct generated clients through the same authenticated HTTP transport;
- keep Terraform schema, import IDs, diffing, timeouts, retries, and state upgrades inside the provider;
- model asynchronous operations through provider create/read/update/delete timeouts;
- pin and deliberately upgrade the SDK so provider behavior does not change when this repository's `main` branch moves.

The generated OpenAPI document can support documentation and non-Go tooling, but the Go provider should use the generated Connect clients as its wire contract.
