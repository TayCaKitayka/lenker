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

assert_token_status() {
  payload="$1"
  expected_status="$2"
  expected_issued="$3"
  expected_generation="$4"
  STATUS_PAYLOAD="$payload" EXPECTED_STATUS="$expected_status" EXPECTED_ISSUED="$expected_issued" EXPECTED_GENERATION="$expected_generation" ruby -rjson -e '
    data = JSON.parse(ENV.fetch("STATUS_PAYLOAD")).fetch("data")
    expected_status = ENV.fetch("EXPECTED_STATUS")
    expected_issued = ENV.fetch("EXPECTED_ISSUED") == "true"
    expected_generation = ENV.fetch("EXPECTED_GENERATION").to_i
    abort("token status mismatch: #{data.inspect}") unless data["status"] == expected_status
    abort("token issued mismatch: #{data.inspect}") unless data["issued"] == expected_issued
    abort("token generation mismatch: #{data.inspect}") unless data["generation"].to_i == expected_generation
    abort("token status leaked plaintext: #{data.inspect}") if data.key?("access_token")
    if expected_status == "never_issued"
      abort("never_issued should not have issued_at: #{data.inspect}") if data.key?("issued_at")
      abort("never_issued should not have revoked_at: #{data.inspect}") if data.key?("revoked_at")
    end
    if expected_status == "active"
      abort("active token missing issued_at: #{data.inspect}") unless data["issued_at"].to_s != ""
      abort("active token should not have revoked_at: #{data.inspect}") if data.key?("revoked_at")
    end
    if expected_status == "revoked"
      abort("revoked token missing issued_at: #{data.inspect}") unless data["issued_at"].to_s != ""
      abort("revoked token missing revoked_at: #{data.inspect}") unless data["revoked_at"].to_s != ""
    end
  '
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

log "checking initial token lifecycle status"
initial_token_status_json="$(curl -fsS "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
assert_token_status "$initial_token_status_json" "never_issued" "false" "0"

log "issuing subscription access token"
access_token_json="$(curl -fsS -X POST "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
access_token="$(printf '%s' "$access_token_json" | json_get data.access_token)"
issued_token_status_json="$(curl -fsS "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
assert_token_status "$issued_token_status_json" "active" "true" "1"

log "checking missing and invalid client access tokens"
missing_status="$(curl -s -o /dev/null -w '%{http_code}' "$PANEL_URL/api/v1/client/subscription-access")"
[ "$missing_status" = "401" ] || fail "missing client access token returned $missing_status, expected 401"
invalid_status="$(curl -s -o /dev/null -w '%{http_code}' "$PANEL_URL/api/v1/client/subscription-access" \
  -H "Authorization: Bearer invalid-subscription-access-token")"
[ "$invalid_status" = "401" ] || fail "invalid client access token returned $invalid_status, expected 401"

log "reading client subscription access"
client_access_json="$(curl -fsS "$PANEL_URL/api/v1/client/subscription-access" \
  -H "Authorization: Bearer $access_token")"

log "rotating subscription access token"
rotated_token_json="$(curl -fsS -X POST "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token/rotate" \
  -H "Authorization: Bearer $admin_token")"
rotated_token="$(printf '%s' "$rotated_token_json" | json_get data.access_token)"
rotated_token_status_json="$(curl -fsS "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
assert_token_status "$rotated_token_status_json" "active" "true" "2"

old_token_status="$(curl -s -o /dev/null -w '%{http_code}' "$PANEL_URL/api/v1/client/subscription-access" \
  -H "Authorization: Bearer $access_token")"
[ "$old_token_status" = "401" ] || fail "rotated old token returned $old_token_status, expected 401"

rotated_client_access_json="$(curl -fsS "$PANEL_URL/api/v1/client/subscription-access" \
  -H "Authorization: Bearer $rotated_token")"

log "revoking rotated subscription access token"
revoked_token_status_json="$(curl -fsS -X DELETE "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
assert_token_status "$revoked_token_status_json" "revoked" "true" "2"

revoked_token_status="$(curl -s -o /dev/null -w '%{http_code}' "$PANEL_URL/api/v1/client/subscription-access" \
  -H "Authorization: Bearer $rotated_token")"
[ "$revoked_token_status" = "401" ] || fail "revoked token returned $revoked_token_status, expected 401"

log "checking repeated revoke remains safe"
repeated_revoke_status_json="$(curl -fsS -X DELETE "$PANEL_URL/api/v1/subscriptions/$subscription_id/access-token" \
  -H "Authorization: Bearer $admin_token")"
assert_token_status "$repeated_revoke_status_json" "revoked" "true" "2"

active_config_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/config.json)"
active_metadata_json="$(docker compose -f "$COMPOSE_FILE" exec -T node-agent cat /var/lib/lenker/node-agent/active/metadata.json)"

ACCESS="$access_json" CLIENT_ACCESS="$client_access_json" ROTATED_CLIENT_ACCESS="$rotated_client_access_json" REVISION="$revision_detail_json" CONFIG="$active_config_json" METADATA="$active_metadata_json" NODE_ID="$node_id" SUBSCRIPTION_ID="$subscription_id" INITIAL_STATUS="$initial_token_status_json" ISSUED_STATUS="$issued_token_status_json" ROTATED_STATUS="$rotated_token_status_json" REVOKED_STATUS="$revoked_token_status_json" REPEATED_REVOKE_STATUS="$repeated_revoke_status_json" ruby -rjson -ruri -e '
  access = JSON.parse(ENV.fetch("ACCESS")).fetch("data")
  client_access = JSON.parse(ENV.fetch("CLIENT_ACCESS")).fetch("data")
  rotated_client_access = JSON.parse(ENV.fetch("ROTATED_CLIENT_ACCESS")).fetch("data")
  revision = JSON.parse(ENV.fetch("REVISION")).fetch("data")
  config = JSON.parse(ENV.fetch("CONFIG"))
  metadata = JSON.parse(ENV.fetch("METADATA"))
  initial_status = JSON.parse(ENV.fetch("INITIAL_STATUS")).fetch("data")
  issued_status = JSON.parse(ENV.fetch("ISSUED_STATUS")).fetch("data")
  rotated_status = JSON.parse(ENV.fetch("ROTATED_STATUS")).fetch("data")
  revoked_status = JSON.parse(ENV.fetch("REVOKED_STATUS")).fetch("data")
  repeated_revoke_status = JSON.parse(ENV.fetch("REPEATED_REVOKE_STATUS")).fetch("data")
  node_id = ENV.fetch("NODE_ID")
  subscription_id = ENV.fetch("SUBSCRIPTION_ID")

  abort("access endpoint selected wrong node") unless access.dig("node", "id") == node_id
  abort("access subscription mismatch") unless access["subscription_id"] == subscription_id
  abort("access client id mismatch") unless access.dig("client", "id") == subscription_id
  abort("access endpoint address mismatch") unless access.dig("endpoint", "address") == access.dig("node", "hostname")
  abort("active metadata revision mismatch") unless metadata["revision_number"].to_i == revision["revision_number"].to_i
  abort("client access subscription mismatch") unless client_access["subscription_id"] == subscription_id
  abort("client access leaked user id") if client_access.key?("user_id") || client_access.key?("user_label")
  abort("client access leaked plan id") if client_access.key?("plan_id") || client_access.dig("client", "plan_id")
  abort("client access uri mismatch") unless client_access["uri"] == access["uri"]
  abort("client access endpoint mismatch") unless client_access["endpoint"] == access["endpoint"]
  abort("client access node mismatch") unless client_access["node"] == access["node"]
  abort("client access client id mismatch") unless client_access.dig("client", "id") == access.dig("client", "id")
  abort("client access flow mismatch") unless client_access.dig("client", "flow") == access.dig("client", "flow")
  abort("rotated client access mismatch") unless rotated_client_access == client_access

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

  abort("initial token status mismatch") unless initial_status["status"] == "never_issued"
  abort("issued token status mismatch") unless issued_status["status"] == "active" && issued_status["generation"].to_i == 1
  abort("rotated token status mismatch") unless rotated_status["status"] == "active" && rotated_status["generation"].to_i == 2
  abort("revoked token status mismatch") unless revoked_status["status"] == "revoked" && revoked_status["generation"].to_i == 2
  abort("repeated revoke status mismatch") unless repeated_revoke_status["status"] == "revoked" && repeated_revoke_status["generation"].to_i == 2

  puts ""
  puts "Subscription handoff smoke summary"
  puts "  subscription_id: #{subscription_id}"
  puts "  selected_node_id: #{node_id}"
  puts "  selected_node_hostname: #{access.dig("node", "hostname")}"
  puts "  endpoint: #{access.dig("endpoint", "address")}:#{access.dig("endpoint", "port")}"
  puts "  protocol_path: #{access["protocol_path"]}"
  puts "  applied_revision: #{revision["revision_number"]}"
  puts "  lifecycle: #{initial_status["status"]} -> #{issued_status["status"]} -> #{rotated_status["status"]}(generation #{rotated_status["generation"]}) -> #{revoked_status["status"]}"
  puts "  client_read: ok"
  puts "  rotate_check: old token rejected, rotated token accepted"
  puts "  revoke_check: revoked token rejected, repeated revoke safe"
  puts "  client_payload_redacted: #{!client_access.key?("user_id") && !client_access.key?("plan_id")}"
  puts "  plaintext_token_printed: false"
'

log "subscription access handoff smoke passed"
