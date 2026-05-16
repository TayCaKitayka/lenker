package configrender

import (
	"encoding/json"
	"testing"

	"github.com/lenker/lenker/services/panel-api/internal/configbundle"
)

func TestRenderVLESSRealityPayloadDeterministic(t *testing.T) {
	input := RenderInput{
		NodeID:         "node-1",
		RevisionNumber: 7,
		Hostname:       "node-1.example.com",
		Region:         "eu",
		CountryCode:    "FI",
	}

	first := RenderVLESSRealityPayload(input)
	second := RenderVLESSRealityPayload(input)

	firstJSON := mustJSON(t, first)
	secondJSON := mustJSON(t, second)
	if firstJSON != secondJSON {
		t.Fatalf("expected deterministic render:\n%s\n---\n%s", firstJSON, secondJSON)
	}

	expected := `{"config":{"inbounds":[{"listen":"0.0.0.0","port":443,"protocol":"vless","settings":{"clients":[],"decryption":"none"},"sniffing":{"destOverride":["http","tls","quic"],"enabled":true},"streamSettings":{"network":"tcp","realitySettings":{"dest":"www.cloudflare.com:443","privateKey":"lenker-placeholder-reality-private-key","serverNames":["www.cloudflare.com"],"shortIds":["lenker00"],"show":false},"security":"reality"},"tag":"vless-reality-in"}],"log":{"loglevel":"warning"},"outbounds":[{"protocol":"freedom","tag":"direct"}],"routing":{"domainStrategy":"AsIs","rules":[{"inboundTag":["vless-reality-in"],"outboundTag":"direct","type":"field"}]}},"config_kind":"xray-config-skeleton","config_text":"lenker xray vless reality skeleton node=node-1 revision=7 protocol=vless-reality-xtls-vision","core_type":"xray","generated_by":"panel-api","node":{"country_code":"FI","hostname":"node-1.example.com","id":"node-1","region":"eu"},"protocol":"vless-reality-xtls-vision","revision_number":7,"schema_version":"config-bundle.v1alpha1","transport":{"network":"tcp","security":"reality","xtls":"vision"}}`
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

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("expected json: %v", err)
	}
	return string(body)
}
