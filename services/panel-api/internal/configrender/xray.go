package configrender

import "fmt"

const (
	SchemaVersion = "config-bundle.v1alpha1"
	GeneratedBy   = "panel-api"
	ProtocolVLESS = "vless-reality-xtls-vision"
	CoreTypeXray  = "xray"
)

type RenderInput struct {
	NodeID         string
	RevisionNumber int
	Hostname       string
	Region         string
	CountryCode    string
}

func RenderVLESSRealityPayload(input RenderInput) map[string]any {
	inboundTag := "vless-reality-in"
	outboundTag := "direct"

	return map[string]any{
		"schema_version":  SchemaVersion,
		"generated_by":    GeneratedBy,
		"protocol":        ProtocolVLESS,
		"revision_number": input.RevisionNumber,
		"node": map[string]any{
			"id":           input.NodeID,
			"hostname":     input.Hostname,
			"region":       input.Region,
			"country_code": input.CountryCode,
		},
		"core_type": CoreTypeXray,
		"transport": map[string]any{
			"network":  "tcp",
			"security": "reality",
			"xtls":     "vision",
		},
		"config_kind": "xray-config-skeleton",
		"config": map[string]any{
			"log": map[string]any{
				"loglevel": "warning",
			},
			"inbounds": []any{
				map[string]any{
					"tag":      inboundTag,
					"listen":   "0.0.0.0",
					"port":     443,
					"protocol": "vless",
					"settings": map[string]any{
						"clients":    []any{},
						"decryption": "none",
					},
					"streamSettings": map[string]any{
						"network":  "tcp",
						"security": "reality",
						"realitySettings": map[string]any{
							"show":        false,
							"dest":        "www.cloudflare.com:443",
							"serverNames": []any{"www.cloudflare.com"},
							"privateKey":  "lenker-placeholder-reality-private-key",
							"shortIds":    []any{"lenker00"},
						},
					},
					"sniffing": map[string]any{
						"enabled":      true,
						"destOverride": []any{"http", "tls", "quic"},
					},
				},
			},
			"outbounds": []any{
				map[string]any{
					"tag":      outboundTag,
					"protocol": "freedom",
				},
			},
			"routing": map[string]any{
				"domainStrategy": "AsIs",
				"rules": []any{
					map[string]any{
						"type":        "field",
						"inboundTag":  []any{inboundTag},
						"outboundTag": outboundTag,
					},
				},
			},
		},
		"config_text": fmt.Sprintf(
			"lenker xray vless reality skeleton node=%s revision=%d protocol=%s",
			input.NodeID,
			input.RevisionNumber,
			ProtocolVLESS,
		),
	}
}
