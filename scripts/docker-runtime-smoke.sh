#!/usr/bin/env sh
set -eu

COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-deploy/docker/docker-compose.local.yml}"
PANEL_URL="${LENKER_PANEL_URL:-http://localhost:8080}"
AGENT_URL="${LENKER_AGENT_URL:-http://localhost:8090}"
ADMIN_EMAIL="${ADMIN_EMAIL:-owner@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-change-me-now}"
POLL_INTERVAL="${LENKER_AGENT_CONFIG_POLL_INTERVAL:-2s}"

log() {
  printf '[runtime-smoke] %s\n' "$*"
}

fail() {
  printf '[runtime-smoke] ERROR: %s\n' "$*" >&2
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

node_name="runtime-smoke-$(date -u +%Y%m%d%H%M%S)"

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

log "creating config revision"
revision_json="$(curl -fsS -X POST "$PANEL_URL/api/v1/nodes/$node_id/config-revisions" \
  -H "Authorization: Bearer $admin_token")"
revision_id="$(printf '%s' "$revision_json" | json_get data.id)"
revision_number="$(printf '%s' "$revision_json" | json_get data.revision_number)"

log "restarting node-agent with registered node identity"
LENKER_AGENT_NODE_ID="$node_id" \
LENKER_AGENT_NODE_TOKEN="$node_token" \
LENKER_AGENT_CONFIG_POLL_INTERVAL="$POLL_INTERVAL" \
docker compose -f "$COMPOSE_FILE" up -d node-agent >/dev/null
wait_for_url "$AGENT_URL/healthz" "node-agent" 45

log "waiting for node-agent apply/report"
i=1
while [ "$i" -le 45 ]; do
  detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id" -H "Authorization: Bearer $admin_token")"
  revision_detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id/config-revisions/$revision_id" -H "Authorization: Bearer $admin_token")"
  if DETAIL="$detail_json" REVISION="$revision_detail_json" ruby -rjson -e '
      detail = JSON.parse(ENV.fetch("DETAIL")).fetch("data")
      revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
      events = detail.fetch("runtime_events", [])
      ok = revision["status"] == "applied" &&
        detail["active_revision_id"].to_i == revision["revision_number"].to_i &&
        detail["last_validation_status"] == "applied" &&
        detail["runtime_state"] == "active_config_ready" &&
        events.any? { |event| event["type"] == "apply_success" && event["status"] == "applied" }
      exit(ok ? 0 : 1)
    '; then
    break
  fi
  sleep 1
  i=$((i + 1))
done

[ "$i" -le 45 ] || fail "node-agent did not apply and report revision $revision_number"

log "checking local active artifacts"
docker compose -f "$COMPOSE_FILE" exec -T node-agent sh -c \
  'test -s /var/lib/lenker/node-agent/active/config.json &&
   test -s /var/lib/lenker/node-agent/active/metadata.json &&
   test -s /var/lib/lenker/node-agent/state.json' >/dev/null

agent_status_json="$(curl -fsS "$AGENT_URL/status")"

DETAIL="$detail_json" REVISION="$revision_detail_json" STATUS="$agent_status_json" ruby -rjson -e '
  detail = JSON.parse(ENV.fetch("DETAIL")).fetch("data")
  revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
  status = JSON.parse(ENV.fetch("STATUS")).fetch("data")
  node_events = detail.fetch("runtime_events", [])
  agent_events = status.fetch("runtime_events", [])
  summary = {
    node_id: detail["id"],
    revision_number: revision["revision_number"],
    revision_status: revision["status"],
    active_revision_id: detail["active_revision_id"],
    runtime_state: detail["runtime_state"],
    last_validation_status: detail["last_validation_status"],
    persisted_runtime_events: node_events.length,
    agent_runtime_events: agent_events.length
  }
  puts JSON.pretty_generate(summary)
'

log "runtime apply/readiness/events smoke passed"
