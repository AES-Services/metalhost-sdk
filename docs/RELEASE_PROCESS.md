# Release process

Metalhost SDK releases are Go module versions containing a reviewed public proto snapshot, reproducible generated clients, and the matching OpenAPI document. The current major line is `v1`; the first stable tag is `v1.0.0`.

There is no separate publish workflow or binary artifact. Pushing a semantic-version Git tag makes that revision available to Go module consumers through the normal module proxy.

## Release checklist

1. Start from the intended API source revision and sync the snapshot:

   ```sh
   METALHOST_API_SOURCE="/path/to/api/source" ./scripts/sync-api-snapshot.sh
   ```

   The script replaces only the allowlisted public packages under `proto/aes`, runs `buf lint` and `buf generate`, finalizes `gen/openapi/metalhost.openapi.yaml`, runs `go mod tidy`, and runs `go test ./...`.

2. Review the proto and generated changes together:
   - `proto/aes/**` is the released source contract.
   - `gen/go/**` must match the proto snapshot.
   - `gen/openapi/metalhost.openapi.yaml` must describe the same RPCs and report API version `v1`.
   - Confirm the OpenAPI finalizer did not expose admin RPCs.

3. Document additions, behavior changes, deprecations, and migrations. Breaking public API changes require a new Go module major version; do not put an incompatible contract behind a later `v1.x.y` tag.

4. Re-run the same checks CI uses:

   ```sh
   make ci
   git diff --exit-code gen/
   ```

   `make tools` installs the pinned versions of Buf and the Go, Connect, and Connect-OpenAPI plugins when needed. OpenAPI finalization also requires `yq`.

5. Merge the reviewed snapshot, ensure CI passes on the release commit, and create the next tag. For example:

   ```sh
   VERSION=v1.0.1
   git tag "$VERSION"
   git push origin "$VERSION"
   ```

6. Verify module resolution from a clean consumer:

   ```sh
   go list -m github.com/AES-Services/metalhost-sdk@v1.0.1
   ```

## Versioning and compatibility

- Patch releases fix implementation, documentation, or generation issues without intentionally changing compatible API behavior.
- Minor releases may add RPCs, messages, fields, and enum values. Consumers should tolerate fields and enum values added by later `v1` releases.
- Removing or renaming public RPCs/fields, changing field meaning incompatibly, or changing the Go module API incompatibly requires a new major version.
- `metalhost.Config` and other hand-written APIs should remain source compatible throughout `v1`.
- The service and SDK are released independently; release notes must state any minimum service version or rollout dependency.
