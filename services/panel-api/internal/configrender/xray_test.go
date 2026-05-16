package configrender

import (
	"encoding/json"
	"testing"

	"github.com/lenker/lenker/services/panel-api/internal/configbundle"
)

func TestRenderVLESSRealityPayloadDeterministic(t *testing.T) {
	input := RenderInput{
		NodeID:                 "node-1",
		RevisionNumber:         7,
		Hostname:               "node-1.example.com",
		Region:                 "eu",
		CountryCode:            "FI",
		RollbackTargetRevision: 6,
	}

	first := RenderVLESSRealityPayload(input)
	second := RenderVLESSRealityPayload(input)

	firstJSON := mustJSON(t, first)
	secondJSON := mustJSON(t, second)
	if firstJSON != secondJSON {
		t.Fatalf("expected deterministic render:\n%s\n---\n%s", firstJSON, secondJSON)
	}

	expected := `{"access_entries":[],"config":{"inbounds":[{"listen":"0.0.0.0","port":443,"protocol":"vless","settings":{"clients":[],"decryption":"none"},"sniffing":{"destOverride":["http","tls","quic"],"enabled":true},"streamSettings":{"network":"tcp","realitySettings":{"dest":"www.cloudflare.com:443","privateKey":"lenker-placeholder-reality-private-key","serverNames":["www.cloudflare.com"],"shortIds":["lenker00"],"show":false},"security":"reality"},"tag":"vless-reality-in"}],"log":{"loglevel":"warning"},"outbounds":[{"protocol":"freedom","tag":"direct"}],"routing":{"domainStrategy":"AsIs","rules":[{"inboundTag":["vless-reality-in"],"outboundTag":"direct","type":"field"}]}},"config_kind":"xray-config-skeleton","config_text":"lenker xray vless reality skeleton node=node-1 revision=7 protocol=vless-reality-xtls-vision subscriptions=0","core_type":"xray","generated_by":"panel-api","node":{"country_code":"FI","hostname":"node-1.example.com","id":"node-1","region":"eu"},"protocol":"vless-reality-xtls-vision","revision_number":7,"rollback_target_revision":6,"schema_version":"config-bundle.v1alpha1","subscription_inputs":[],"transport":{"network":"tcp","security":"reality","xtls":"vision"}}`
	if firstJSON != expected {
		t.Fatalf("unexpected render:\n%s", firstJSON)
	}
}

func TestRenderVLESSRealityPayloadHashStable(t *testing.T) {
	payload := RenderVLESSRealityPayload(RenderInput{NodeID: "node-1", RevisionNumber: 7})
	firstHash, err := configbundle.HashPayload(payload)
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}
	secondHash, err := configbundle.HashPayload(payload)
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}
	if firstHash != secondHash {
		t.Fatalf("expected stable hash")
	}
}

func TestRenderVLESSRealityPayloadOrdersSubscriptionInputs(t *testing.T) {
	limit := int64(1024)
	payload := RenderVLESSRealityPayload(RenderInput{
		NodeID:         "node-1",
		RevisionNumber: 7,
		SubscriptionInputs: []SubscriptionInput{
			{SubscriptionID: "sub-b", UserID: "user-b", PlanID: "plan-b", UserStatus: "active", SubscriptionStatus: "active", PreferredRegion: "eu", PlanName: "Pro", DeviceLimit: 2, TrafficLimitBytes: &limit, StartsAt: "2026-05-01T00:00:00Z", ExpiresAt: "2026-06-01T00:00:00Z"},
			{SubscriptionID: "sub-a", UserID: "user-a", PlanID: "plan-a", UserStatus: "active", SubscriptionStatus: "active", PreferredRegion: "", PlanName: "Basic", DeviceLimit: 1, StartsAt: "2026-05-01T00:00:00Z", ExpiresAt: "2026-06-01T00:00:00Z"},
		},
	})

	subscriptions := payload["subscription_inputs"].([]any)
	first := subscriptions[0].(map[string]any)
	if first["subscription_id"] != "sub-a" {
		t.Fatalf("expected sorted subscription inputs, got %#v", subscriptions)
	}
	accessEntries := payload["access_entries"].([]any)
	firstAccess := accessEntries[0].(map[string]any)
	if firstAccess["vless_client_id"] != "sub-a" {
		t.Fatalf("expected sorted access entries, got %#v", accessEntries)
	}
	config := payload["config"].(map[string]any)
	inbound := config["inbounds"].([]any)[0].(map[string]any)
	settings := inbound["settings"].(map[string]any)
	clients := settings["clients"].([]any)
	if len(clients) != 2 {
		t.Fatalf("expected two rendered clients, got %#v", clients)
	}
	firstClient := clients[0].(map[string]any)
	if firstClient["id"] != "sub-a" || firstClient["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected first client: %#v", firstClient)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("expected json: %v", err)
	}
	return string(body)
}
