#!/usr/bin/env sh
set -eu

COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-deploy/docker/docker-compose.local.yml}"
PANEL_URL="${LENKER_PANEL_URL:-http://localhost:8080}"
AGENT_URL="${LENKER_AGENT_URL:-http://localhost:8090}"
ADMIN_EMAIL="${ADMIN_EMAIL:-owner@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-change-me-now}"
POLL_INTERVAL="${LENKER_AGENT_CONFIG_POLL_INTERVAL:-2s}"
MISSING_XRAY_BIN="${LENKER_MISSING_XRAY_BIN:-/opt/lenker/xray/missing-xray-for-smoke}"

log() {
  printf '[runtime-failure-smoke] %s\n' "$*"
}

fail() {
  printf '[runtime-failure-smoke] ERROR: %s\n' "$*" >&2
  exit 1
}

need_command() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

json_get() {
  ruby -rjson -e '
    value = JSON.parse(STDIN.read)
    ARGV.fetch(0).split(".").each { |key| value = value.fetch(key) }
    puts value
  ' "$1"
}

wait_for_url() {
  url="$1"
  label="$2"
  tries="${3:-30}"
  i=1
  while [ "$i" -le "$tries" ]; do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  fail "$label did not become ready at $url"
}

wait_for_revision_state() {
  node_id="$1"
  revision_id="$2"
  admin_token="$3"
  expected_status="$4"
  tries="${5:-45}"
  i=1
  while [ "$i" -le "$tries" ]; do
    detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id" -H "Authorization: Bearer $admin_token")"
    revision_detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id/config-revisions/$revision_id" -H "Authorization: Bearer $admin_token")"
    if DETAIL="$detail_json" REVISION="$revision_detail_json" EXPECTED="$expected_status" ruby -rjson -e '
        detail = JSON.parse(ENV.fetch("DETAIL")).fetch("data")
        revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
        expected = ENV.fetch("EXPECTED")
        events = detail.fetch("runtime_events", [])
        ok = revision["status"] == expected
        if expected == "applied"
          ok &&= detail["active_revision_id"].to_i == revision["revision_number"].to_i
          ok &&= detail["last_validation_status"] == "applied"
          ok &&= detail["runtime_state"] == "active_config_ready"
          ok &&= events.any? { |event| event["type"] == "apply_success" && event["status"] == "applied" }
        else
          ok &&= detail["last_validation_status"] == "failed"
          ok &&= detail["runtime_state"] == "validation_failed"
          ok &&= detail["last_validation_error"].to_s.start_with?("xray_dry_run_failed:xray_binary_not_found")
          ok &&= events.any? { |event| event["type"] == "dry_run_failure" && event["status"] == "failed" }
        end
        exit(ok ? 0 : 1)
      '; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  return 1
}

need_command docker
need_command curl
need_command ruby

if ! docker compose version >/dev/null 2>&1; then
  fail "Docker daemon or docker compose is unavailable"
fi

log "starting local Docker stack"
docker compose -f "$COMPOSE_FILE" up -d postgres migrate panel-api node-agent >/dev/null
wait_for_url "$PANEL_URL/healthz" "panel-api" 45
wait_for_url "$AGENT_URL/healthz" "node-agent" 45

log "bootstrapping admin"
docker compose -f "$COMPOSE_FILE" --profile setup run --rm bootstrap-admin >/dev/null

log "logging in as admin"
login_json="$(curl -fsS "$PANEL_URL/api/v1/auth/admin/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")"
admin_token="$(printf '%s' "$login_json" | json_get data.session.token)"

node_name="runtime-failure-smoke-$(date -u +%Y%m%d%H%M%S)"

log "creating bootstrap token"
bootstrap_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/bootstrap-token" \
  -H "Authorization: Bearer $admin_token" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$node_name\",\"region\":\"eu\",\"country_code\":\"FI\",\"hostname\":\"$node_name\",\"expires_in_minutes\":30}")"
node_id="$(printf '%s' "$bootstrap_json" | json_get data.node_id)"
bootstrap_token="$(printf '%s' "$bootstrap_json" | json_get data.bootstrap_token)"

log "registering node"
register_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/register" \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$node_id\",\"bootstrap_token\":\"$bootstrap_token\",\"agent_version\":\"0.1.0-dev\",\"hostname\":\"$node_name\"}")"
node_token="$(printf '%s' "$register_json" | json_get data.node_token)"

log "creating baseline config revision"
revision_a_json="$(curl -fsS -X POST "$PANEL_URL/api/v1/nodes/$node_id/config-revisions" \
  -H "Authorization: Bearer $admin_token")"
revision_a_id="$(printf '%s' "$revision_a_json" | json_get data.id)"
revision_a_number="$(printf '%s' "$revision_a_json" | json_get data.revision_number)"

log "applying baseline revision without external dry-run"
LENKER_AGENT_NODE_ID="$node_id" \
LENKER_AGENT_NODE_TOKEN="$node_token" \
LENKER_AGENT_CONFIG_POLL_INTERVAL="$POLL_INTERVAL" \
LENKER_AGENT_XRAY_BIN="" \
docker compose -f "$COMPOSE_FILE" up -d node-agent >/dev/null
wait_for_url "$AGENT_URL/healthz" "node-agent" 45
if ! wait_for_revision_state "$node_id" "$revision_a_id" "$admin_token" "applied" 45; then
  fail "baseline revision $revision_a_number was not applied"
fi

baseline_metadata_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/metadata.json)"
baseline_revision_number="$(printf '%s' "$baseline_metadata_json" | json_get revision_number)"
[ "$baseline_revision_number" = "$revision_a_number" ] || fail "active artifact does not match baseline revision"

log "creating revision expected to fail dry-run"
revision_b_json="$(curl -fsS -X POST "$PANEL_URL/api/v1/nodes/$node_id/config-revisions" \
  -H "Authorization: Bearer $admin_token")"
revision_b_id="$(printf '%s' "$revision_b_json" | json_get data.id)"
revision_b_number="$(printf '%s' "$revision_b_json" | json_get data.revision_number)"

log "forcing optional Xray dry-run failure with missing binary"
LENKER_AGENT_NODE_ID="$node_id" \
LENKER_AGENT_NODE_TOKEN="$node_token" \
LENKER_AGENT_CONFIG_POLL_INTERVAL="$POLL_INTERVAL" \
LENKER_AGENT_XRAY_BIN="$MISSING_XRAY_BIN" \
docker compose -f "$COMPOSE_FILE" up -d node-agent >/dev/null
wait_for_url "$AGENT_URL/healthz" "node-agent" 45
if ! wait_for_revision_state "$node_id" "$revision_b_id" "$admin_token" "failed" 45; then
  fail "revision $revision_b_number did not fail through dry-run boundary"
fi

after_metadata_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/metadata.json)"
after_revision_number="$(printf '%s' "$after_metadata_json" | json_get revision_number)"
[ "$after_revision_number" = "$baseline_revision_number" ] || fail "failed dry-run changed active artifact"

detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id" -H "Authorization: Bearer $admin_token")"
revision_b_detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id/config-revisions/$revision_b_id" -H "Authorization: Bearer $admin_token")"
agent_status_json="$(curl -fsS "$AGENT_URL/status")"

DETAIL="$detail_json" REVISION="$revision_b_detail_json" STATUS="$agent_status_json" BASELINE="$baseline_revision_number" ARTIFACT="$after_revision_number" ruby -rjson -e '
  detail = JSON.parse(ENV.fetch("DETAIL")).fetch("data")
  revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
  status = JSON.parse(ENV.fetch("STATUS")).fetch("data")
  baseline = ENV.fetch("BASELINE").to_i
  artifact = ENV.fetch("ARTIFACT").to_i
  node_events = detail.fetch("runtime_events", [])
  dry_run_events = node_events.select { |event| event["type"] == "dry_run_failure" }
  summary = {
    node_id: detail["id"],
    baseline_active_revision: baseline,
    failed_revision_number: revision["revision_number"],
    failed_revision_status: revision["status"],
    active_revision_id_after_failure: detail["active_revision_id"],
    runtime_state: detail["runtime_state"],
    last_validation_status: detail["last_validation_status"],
    last_validation_error: detail["last_validation_error"],
    persisted_runtime_events: node_events.length,
    dry_run_failure_events: dry_run_events.length,
    active_artifact_revision: artifact,
    agent_runtime_state: status["runtime_state"]
  }
  puts JSON.pretty_generate(summary)
  abort("revision falsely applied") unless revision["status"] == "failed"
  abort("active revision moved after failure") unless detail["active_revision_id"].to_i == baseline
  abort("runtime status did not reflect failure") unless detail["last_validation_status"] == "failed" && detail["runtime_state"] == "validation_failed"
  abort("missing compact dry-run error") unless detail["last_validation_error"].to_s.start_with?("xray_dry_run_failed:xray_binary_not_found")
  abort("missing dry_run_failure runtime event") if dry_run_events.empty?
'

log "runtime dry-run failure smoke passed"
