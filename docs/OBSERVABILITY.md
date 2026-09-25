# September observability (unreleased)

These contracts require the coordinated September backend release, enabled
automation/monitoring features and provisioned telemetry storage. A generated
client or successful save does not prove the deployment has enabled a feature.
The APIs preserve existing browser sessions, legacy keys and billing behavior.

## Credentials and GitHub Actions

Use `aes.iam.v1.AutomationService` for project service accounts, explicit
capabilities, expiring keys and GitHub trusts. Shared automation should use a
service account; personal keys are appropriate for individual local tooling.
`monitoring.read` permits metric reads, not VM changes or alert management.
Trust permissions are intersected with the service account's current grants.

Retain a UUID `request_id` and the exact request through ambiguous creation or
rotation failures. `secret_unavailable` means the request already committed;
the API returns metadata, never the original secret. Inspect/revoke that key
before creating another. Overlap rotation is explicit and bounded to 24 hours.
Legacy IAM rotation retains its original immediate-revocation semantics.

For keyless GitHub Actions, use the customer portal's signed workflow-discovery
flow and approve the observed immutable repository/owner and workflow identity.
Do not trust a repository name alone. The workflow needs `id-token: write`;
assertions must match the deployment's configured audience and fixed GitHub
issuer. Pull-request events are not permitted. The exchanged token lasts 15
minutes, has no refresh token, and can only use its configured project/grants.
The September CLI branch implements `auth github` and `monitoring`; older
CLI releases do not. Alternatively, use the
[complete HTTP workflow example](https://metalhost.net/docs/developers/guides/github-actions)
from the coordinated docs release. Exchange through
`AutomationService.ExchangeGitHubToken` with `trust_name` and `assertion`; use
the returned `access_token` as Bearer, scoped to the returned `project_name`.
Mask both tokens and keep them in one step, not a config file, `GITHUB_ENV` or
artifact. Long jobs need a fresh assertion/exchange. The trust's exact workflow,
branch, event and environment must match verification. This is not GitHub sign-in
and does not provision runners. Managed runners are planned separately.

## Metrics, Grafana and Prometheus

`Config.MonitoringEndpoints(project)` validates the API origin and canonical
project name and returns:

- `Prometheus`: Grafana datasource base; append `/api/v1/query` or
  `/api/v1/query_range` for direct PromQL requests.
- `ScrapeTargets`: authenticated Prometheus HTTP service discovery.
- `Metrics`: one scrape shard; prefer discovery for automatic project sizing.

Use HTTPS with `Authorization: Bearer <project monitoring.read credential>`.
Never supply `X-Scope-OrgID`; the gateway derives tenant identity from current
authorization. Endpoint construction itself does not authorize access.

Grafana uses a Prometheus datasource (Mimir 3.2.1 compatibility), 30-second
minimum interval and 10-second query timeout. Store the key in
`secureJsonData.httpHeaderValue1`, not dashboard JSON. Download the Metalhost VM
dashboard from the portal's Monitoring → Integrations page. Data-source-managed
rule administration is not exposed; create managed alerts through Metalhost's
structured rule API. Customer Grafana-managed alerts query the read API and
remain subject to its budgets.

Prometheus should use authenticated HTTP discovery and scraping, a 60-second
interval, 10-second timeout, and `honor_timestamps: true`. Both discovery and
scraping need `authorization.credentials_file`. Keep the credential file mode
0600. The exporter omits expired samples instead of refreshing old observations.

The initial hosted limit is seven days, minimum 30-second query steps, 10,000
points per series, a 16-KiB public expression cap and a 4-MiB response
cap. Concurrency is per project per gateway replica: four in earlier previews,
twelve in the follow-up backend implementation, also subject to a global budget.
Narrow queries on 422; back off on 429; show an unavailable state on service
failure. Empty, unsupported and stale metrics are not zero or proof of health.
Use `MonitoringService` for typed metric descriptors, VM summaries and curated
chart responses with per-family quality. Metric samples are never invoices.

### Runnable read-only example

The [monitoring example](../examples/monitoring/main.go) uses this checkout's
generated client and explicitly installs `Config.RoundTripper`; `Config.Client`
does not add Bearer authentication on its own. Set `METALHOST_ENDPOINT` to an
enabled HTTPS API origin, `METALHOST_VM` to a full resource name, and load
`METALHOST_API_KEY` from your secret manager with `monitoring.read`, then run:

```sh
go run ./examples/monitoring
```

It queries one hour of CPU/memory, prints quality and sample counts, and changes
no resources. Seven-day queries need a coarser step (for example 1800 seconds),
not 60 seconds. Use returned bounds/step and retain quality alongside values.
List calls use `next_page_token`/`page_token`; one page is not a whole inventory.

Optional guest pause/resume is a newer follow-up API not present in this SDK
snapshot. Do not assume generated clients track an unreleased backend checkout.
The portal and verified collector archive remain the installation entry points;
revocation is permanent, unlike pause, and does not uninstall guest software.

## Alerts and destinations

`AlertService` manages structured rules, current/future VM selections, previews,
incidents, verified destinations and explicit test deliveries. Rules reconcile
asynchronously: check desired/applied version and evaluation health after save.
Preview defaults to the current sample. Set `history_seconds` to 3600, 21600 or
86400 for a bounded historical estimate on up to 50 matching VMs. Inspect
`history_status` and each resource's evaluated/expected sample counts: missing
retained running intent or fresh telemetry produces gaps, not healthy samples.
The episode estimate applies the proposed threshold and sustained duration; it
does not replay routing, delivery or predict future notifications.
Enhanced templates accept only their allowlisted exact `dimensions`; obtain
mount/service/GPU names from observations rather than constructing PromQL.
Acknowledging or snoozing an incident does not resolve it. A missing metric or
collector outage must not manufacture recovery.
`OPEN` means recovery has not been confirmed, not that the retained incident
value is live. `RECOVERED` requires recovery evidence; `RETIRED` records a changed
rule/resource context rather than fabricated recovery. Rule saving is a separate
mutation from preview; use stable UUIDs and expected versions.

Select existing project-member or webhook rows as destinations. Email needs
recipient verification; destination pause stops future sending without changing
incident history. `TestAlertDestination` is an explicit external send with a
stable request UUID; poll `GetDestinationTest`. The latest retained result is
also returned by `ListAlertDestinations`. `SENT` means provider/endpoint
acceptance, not that a person read it.
Slack, Discord and Teams destinations take a write-only provider webhook URL
at creation. Changing that endpoint requires a new destination. Check the
deployment's channel availability before asking recipients to verify or test.
Scoped automation is an RPC allowlist: `monitoring.write` is not blanket access
to every AlertService method. In particular, verification, destination tests
and incident acknowledgements use the authorized human dashboard flow in the
current implementation.

## Verify and rotate webhook signatures

The payload is at-least-once. Deduplicate its stable event IDs durably. Grouped
payloads contain at most 100 detailed events but enumerate every included event
ID; follow incident IDs for the full state rather than assuming details are
complete. The stable delivery ID is reused for identical group retries; a
different attempt UUID is generated on each HTTP request.

Use `metalhost.VerifyWebhook(headers, rawBody, secret, time.Now())` before
decoding the JSON. Limit bodies to 1 MiB before reading them, reject invalid
signatures and keep a trustworthy clock. The verifier authenticates
`timestamp.deliveryID.attemptID.rawBody` with HMAC-SHA256 and checks a five-minute
clock/replay window. It does not implement durable deduplication for you. It
never falls back to the legacy untimestamped header.

See the compilable [receiver example](../examples/webhookreceiver/receiver.go).
Supply an `Inbox` implementation that atomically stores the verified raw body
with a unique delivery ID before returning 2xx. An in-memory map is insufficient
across restarts/replicas. Decode/process from that durable inbox; deduplicate
individual event IDs as well when grouped deliveries overlap.

Check examples without making API requests:

```sh
go test ./metalhost ./examples/...
```

`WebhooksService.RotateSubscriptionSecret` requires the current secret version
and a stable request UUID. A successful original call returns one new secret;
retry recovery returns metadata plus `secret_unavailable`. Select immediate
rotation to invalidate both older keys if a secret was lost or compromised.

With a nonzero overlap (maximum 24 hours), `X-Metalhost-Signature-V2` contains
`t=<unix>,v2=<new HMAC>,v2=<old HMAC>`. Either secret can verify it during that
window. The original `X-Metalhost-Signature: sha256=<HMAC of raw body>` header
continues using the old key until overlap expires, then uses the new key.
Upgrade receivers to V2 before rotating; a legacy receiver cannot validate the
new key through the original header until overlap ends. A second overlapping
rotation is rejected while the first is active; immediate rotation is always
available with the current version. Expired keys never sign new attempts.

Webhook endpoints must be public HTTPS; redirects/private-address destinations
are rejected. Test receivers must be explicitly approved fixtures. Never log
the signing secret, credential response or raw support/customer payload.
