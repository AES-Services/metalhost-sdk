#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export PATH="$(go env GOPATH)/bin:${PATH}"

finalize_openapi() {
  local spec="${ROOT}/gen/openapi/metalhost.openapi.yaml"
  local tmp
  tmp="$(mktemp)"

  command -v yq >/dev/null 2>&1 || {
    echo "yq is required (brew install yq or https://github.com/mikefarah/yq)" >&2
    exit 1
  }

  yq '
    .info = {
      "title": "Metalhost API",
      "version": "v1",
      "description": "Customer-facing HTTP API for Metalhost. Every dashboard action maps to one of the RPCs below. Authenticate with a Bearer API key on every request — mint one in the dashboard under Developers → API keys.\n\nEvery RPC is POST application/json with a Bearer header. Field names use snake_case (proto JSON).",
      "contact": {"name": "Metalhost support", "email": "support@metalhost.net", "url": "https://metalhost.net/docs"}
    }
    | .servers = [{"url": "https://api.metalhost.net", "description": "Production"}]
    | (.paths |= with_entries(select(.key | test("Admin[A-Z]") | not)))
    | (.tags  |= map(select(.name | test("Admin[A-Z]") | not)))
  ' "$spec" > "$tmp"
  mv "$tmp" "$spec"

  echo "finalized ${spec} ($(yq '.paths | length' "$spec") paths)"
}

if [[ "${1:-}" == "--finalize-openapi-only" ]]; then
  finalize_openapi
  exit 0
fi

API_SOURCE="${METALHOST_API_SOURCE:-}"

if [[ -z "${API_SOURCE}" ]]; then
  echo "METALHOST_API_SOURCE is required" >&2
  exit 2
fi

if [[ ! -d "${API_SOURCE}/proto/aes" ]]; then
  echo "missing proto tree: ${API_SOURCE}/proto/aes" >&2
  exit 2
fi

rm -rf "${ROOT}/proto/aes" "${ROOT}/gen/go" "${ROOT}/gen/openapi"
mkdir -p "${ROOT}/proto/aes" "${ROOT}/gen"

for pkg in \
  audit \
  baremetal \
  catalog \
  compute \
  health \
  iam \
  network \
  ops \
  project \
  quota \
  storage \
  support \
  wallet \
  webhooks
do
  if [[ -d "${API_SOURCE}/proto/aes/${pkg}" ]]; then
    cp -R "${API_SOURCE}/proto/aes/${pkg}" "${ROOT}/proto/aes/${pkg}"
  fi
done

cd "${ROOT}"
buf lint
buf generate
finalize_openapi
go mod tidy
go test ./...
