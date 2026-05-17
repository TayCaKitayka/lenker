#!/usr/bin/env sh
set -eu

COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-deploy/docker/docker-compose.local.yml}"
PANEL_URL="${LENKER_PANEL_URL:-http://localhost:8080}"
AGENT_URL="${LENKER_AGENT_URL:-http://localhost:8090}"
ADMIN_EMAIL="${ADMIN_EMAIL:-owner@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-change-me-now}"
POLL_INTERVAL="${LENKER_AGENT_CONFIG_POLL_INTERVAL:-2s}"

log() {
  printf '[subscription-access-smoke] %s\n' "$*"
}

fail() {
  printf '[subscription-access-smoke] ERROR: %s\n' "$*" >&2
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

suffix="$(date -u +%Y%m%d%H%M%S)"
node_name="access-smoke-$suffix"
node_region="access-$suffix"
user_email="access-smoke-$suffix@example.test"

log "creating user, plan, and node"
user_json="$(curl -fsS "$PANEL_URL/api/v1/users" \
  -H "Authorization: Bearer $admin_token" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$user_email\",\"display_name\":\"Access Smoke $suffix\"}")"
user_id="$(printf '%s' "$user_json" | json_get data.id)"

plan_json="$(curl -fsS "$PANEL_URL/api/v1/plans" \
  -H "Authorization: Bearer $admin_token" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"Access Smoke $suffix\",\"duration_days\":30,\"device_limit\":2}")"
plan_id="$(printf '%s' "$plan_json" | json_get data.id)"

bootstrap_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/bootstrap-token" \
  -H "Authorization: Bearer $admin_token" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$node_name\",\"region\":\"$node_region\",\"country_code\":\"FI\",\"hostname\":\"$node_name.example.test\",\"expires_in_minutes\":30}")"
node_id="$(printf '%s' "$bootstrap_json" | json_get data.node_id)"
bootstrap_token="$(printf '%s' "$bootstrap_json" | json_get data.bootstrap_token)"

log "registering node"
register_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/register" \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$node_id\",\"bootstrap_token\":\"$bootstrap_token\",\"agent_version\":\"0.1.0-dev\",\"hostname\":\"$node_name.example.test\"}")"
node_token="$(printf '%s' "$register_json" | json_get data.node_token)"

log "creating active subscription"
subscription_json="$(curl -fsS "$PANEL_URL/api/v1/subscriptions" \
  -H "Authorization: Bearer $admin_token" \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$user_id\",\"plan_id\":\"$plan_id\",\"preferred_region\":\"$node_region\"}")"
subscription_id="$(printf '%s' "$subscription_json" | json_get data.id)"

log "creating config revision with subscription-aware payload"
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

log "waiting for apply/report"
i=1
while [ "$i" -le 45 ]; do
  detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id" -H "Authorization: Bearer $admin_token")"
  revision_detail_json="$(curl -fsS "$PANEL_URL/api/v1/nodes/$node_id/config-revisions/$revision_id" -H "Authorization: Bearer $admin_token")"
  if DETAIL="$detail_json" REVISION="$revision_detail_json" ruby -rjson -e '
      detail = JSON.parse(ENV.fetch("DETAIL")).fetch("data")
      revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
      ok = revision["status"] == "applied" &&
        detail["active_revision_id"].to_i == revision["revision_number"].to_i &&
        detail["last_validation_status"] == "applied"
      exit(ok ? 0 : 1)
    '; then
    break
  fi
  sleep 1
  i=$((i + 1))
done

[ "$i" -le 45 ] || fail "node-agent did not apply and report revision $revision_number"

log "fetching subscription access export"
access_json="$(curl -fsS "$PANEL_URL/api/v1/subscriptions/$subscription_id/access" \
  -H "Authorization: Bearer $admin_token")"
active_config_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/config.json)"
active_metadata_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/metadata.json)"

ACCESS="$access_json" REVISION="$revision_detail_json" CONFIG="$active_config_json" METADATA="$active_metadata_json" NODE_ID="$node_id" SUBSCRIPTION_ID="$subscription_id" ruby -rjson -ruri -e '
  access = JSON.parse(ENV.fetch("ACCESS")).fetch("data")
  revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
  config = JSON.parse(ENV.fetch("CONFIG"))
  metadata = JSON.parse(ENV.fetch("METADATA"))
  node_id = ENV.fetch("NODE_ID")
  subscription_id = ENV.fetch("SUBSCRIPTION_ID")

  abort("access endpoint selected wrong node") unless access.dig("node", "id") == node_id
  abort("access subscription mismatch") unless access["subscription_id"] == subscription_id
  abort("access client id mismatch") unless access.dig("client", "id") == subscription_id
  abort("access endpoint address mismatch") unless access.dig("endpoint", "address") == access.dig("node", "hostname")
  abort("active metadata revision mismatch") unless metadata["revision_number"].to_i == revision["revision_number"].to_i

  entries = revision.dig("bundle", "access_entries") || []
  entry = entries.find { |item| item["subscription_id"] == subscription_id }
  abort("revision bundle missing access entry") unless entry
  abort("bundle access client mismatch") unless entry["vless_client_id"] == access.dig("client", "id")

  inbound = config.fetch("inbounds").fetch(0)
  client = inbound.fetch("settings").fetch("clients").find { |item| item["id"] == access.dig("client", "id") }
  abort("active config missing exported client") unless client
  abort("active config client flow mismatch") unless client["flow"] == access.dig("client", "flow")
  abort("active config port mismatch") unless inbound["port"].to_i == access.dig("endpoint", "port").to_i
  abort("active config network mismatch") unless inbound.dig("streamSettings", "network") == access.dig("endpoint", "network")
  abort("active config security mismatch") unless inbound.dig("streamSettings", "security") == access.dig("endpoint", "security")
  abort("active config sni mismatch") unless inbound.dig("streamSettings", "realitySettings", "serverNames").include?(access.dig("endpoint", "sni"))
  abort("active config short id mismatch") unless inbound.dig("streamSettings", "realitySettings", "shortIds").include?(access.dig("endpoint", "short_id"))

  uri = URI.parse(access.fetch("uri"))
  query = URI.decode_www_form(uri.query || "").to_h
  abort("uri scheme mismatch") unless uri.scheme == "vless"
  abort("uri client mismatch") unless uri.user == access.dig("client", "id")
  abort("uri host mismatch") unless uri.host == access.dig("endpoint", "address")
  abort("uri port mismatch") unless uri.port == access.dig("endpoint", "port").to_i
  abort("uri flow mismatch") unless query["flow"] == access.dig("client", "flow")
  abort("uri security mismatch") unless query["security"] == access.dig("endpoint", "security")
  abort("uri network mismatch") unless query["type"] == access.dig("endpoint", "network")
  abort("uri sni mismatch") unless query["sni"] == access.dig("endpoint", "sni")
  abort("uri short id mismatch") unless query["sid"] == access.dig("endpoint", "short_id")
  abort("uri public key mismatch") unless query["pbk"] == access.dig("endpoint", "public_key")

  summary = {
    subscription_id: subscription_id,
    node_id: node_id,
    revision_number: revision["revision_number"],
    access_protocol: access["protocol"],
    access_address: access.dig("endpoint", "address"),
    access_client_id: access.dig("client", "id"),
    active_config_clients: inbound.fetch("settings").fetch("clients").length
  }
  puts JSON.pretty_generate(summary)
'

log "subscription access export smoke passed"
