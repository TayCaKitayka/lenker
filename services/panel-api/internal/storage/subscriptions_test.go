package storage

import (
	"strings"
	"testing"

	"github.com/lenker/lenker/services/panel-api/internal/configrender"
)

func TestCreateSubscriptionSQLCastsStartParameter(t *testing.T) {
	if !strings.Contains(createSubscriptionSQL, "$3::timestamptz") {
		t.Fatalf("create subscription SQL must cast starts_at parameter: %s", createSubscriptionSQL)
	}
	if strings.Contains(createSubscriptionSQL, " $3, $3 + ") {
		t.Fatalf("create subscription SQL must not reuse untyped $3 in interval expression: %s", createSubscriptionSQL)
	}
}

func TestBuildVLESSRealityURIDeterministic(t *testing.T) {
	endpoint := SubscriptionAccessEndpoint{
		Address:     "node.example.com",
		Port:        configrender.DefaultVLESSPort,
		Network:     "tcp",
		Security:    "reality",
		SNI:         configrender.DefaultRealitySNI,
		PublicKey:   configrender.DefaultRealityPublic,
		ShortID:     configrender.DefaultRealityShortID,
		Fingerprint: configrender.DefaultFingerprint,
		SpiderX:     configrender.DefaultSpiderX,
	}
	client := SubscriptionAccessClient{
		ID:    "sub-1",
		Email: "subscription:sub-1",
		Flow:  configrender.DefaultVLESSFlow,
		Level: 0,
	}

	first := buildVLESSRealityURI(endpoint, client, "Lenker eu Basic")
	second := buildVLESSRealityURI(endpoint, client, "Lenker eu Basic")

	if first != second {
		t.Fatalf("expected deterministic URI")
	}
	expected := "vless://sub-1@node.example.com:443?encryption=none&flow=xtls-rprx-vision&fp=chrome&pbk=lenker-placeholder-reality-public-key&security=reality&sid=lenker00&sni=www.cloudflare.com&spx=%2F&type=tcp#Lenker%20eu%20Basic"
	if first != expected {
		t.Fatalf("unexpected URI:\n%s", first)
	}
}

func TestHashSubscriptionAccessTokenStableAndNotRaw(t *testing.T) {
	token := "lnksa_example"
	first := HashSubscriptionAccessToken(token)
	second := HashSubscriptionAccessToken(token)
	if first != second {
		t.Fatalf("expected stable access token hash")
	}
	if first == token || !strings.Contains("0123456789abcdef", first[:1]) || len(first) != 64 {
		t.Fatalf("expected sha256 hex hash, got %q", first)
	}
}

func TestHashSubscriptionHandoffTokenStableAndNotRaw(t *testing.T) {
	token := "lnkhi_example"
	first := HashSubscriptionHandoffToken(token)
	second := HashSubscriptionHandoffToken(token)
	if first != second {
		t.Fatalf("expected stable handoff token hash")
	}
	if first == token || !strings.Contains("0123456789abcdef", first[:1]) || len(first) != 64 {
		t.Fatalf("expected sha256 hex hash, got %q", first)
	}
}
