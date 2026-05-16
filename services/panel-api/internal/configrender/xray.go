package configrender

import (
	"fmt"
	"sort"
)

const (
	SchemaVersion = "config-bundle.v1alpha1"
	GeneratedBy   = "panel-api"
	ProtocolVLESS = "vless-reality-xtls-vision"
	CoreTypeXray  = "xray"
)

type RenderInput struct {
	NodeID                 string
	RevisionNumber         int
	Hostname               string
	Region                 string
	CountryCode            string
	SubscriptionInputs     []SubscriptionInput
	RollbackTargetRevision int
}

type SubscriptionInput struct {
	SubscriptionID     string
	UserID             string
	PlanID             string
	UserStatus         string
	SubscriptionStatus string
	PreferredRegion    string
	PlanName           string
	DeviceLimit        int
	TrafficLimitBytes  *int64
	StartsAt           string
	ExpiresAt          string
}

func RenderVLESSRealityPayload(input RenderInput) map[string]any {
	inboundTag := "vless-reality-in"
	outboundTag := "direct"
	subscriptionInputs := sortedSubscriptionInputs(input.SubscriptionInputs)
	accessEntries := renderAccessEntries(subscriptionInputs)
	subscriptionSummary := renderSubscriptionSummary(subscriptionInputs)

	return map[string]any{
		"schema_version":           SchemaVersion,
		"generated_by":             GeneratedBy,
		"protocol":                 ProtocolVLESS,
		"revision_number":          input.RevisionNumber,
		"rollback_target_revision": input.RollbackTargetRevision,
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
						"clients":    renderClients(accessEntries),
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
		"subscription_inputs": subscriptionSummary,
		"access_entries":      accessEntries,
		"config_text": fmt.Sprintf(
			"lenker xray vless reality skeleton node=%s revision=%d protocol=%s subscriptions=%d",
			input.NodeID,
			input.RevisionNumber,
			ProtocolVLESS,
			len(subscriptionInputs),
		),
	}
}

func sortedSubscriptionInputs(inputs []SubscriptionInput) []SubscriptionInput {
	result := append([]SubscriptionInput(nil), inputs...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SubscriptionID != result[j].SubscriptionID {
			return result[i].SubscriptionID < result[j].SubscriptionID
		}
		if result[i].UserID != result[j].UserID {
			return result[i].UserID < result[j].UserID
		}
		return result[i].PlanID < result[j].PlanID
	})
	return result
}

func renderSubscriptionSummary(inputs []SubscriptionInput) []any {
	result := make([]any, 0, len(inputs))
	for _, input := range inputs {
		entry := map[string]any{
			"subscription_id":     input.SubscriptionID,
			"user_id":             input.UserID,
			"plan_id":             input.PlanID,
			"user_status":         input.UserStatus,
			"subscription_status": input.SubscriptionStatus,
			"preferred_region":    input.PreferredRegion,
			"plan_name":           input.PlanName,
			"device_limit":        input.DeviceLimit,
			"starts_at":           input.StartsAt,
			"expires_at":          input.ExpiresAt,
		}
		if input.TrafficLimitBytes != nil {
			entry["traffic_limit_bytes"] = *input.TrafficLimitBytes
		} else {
			entry["traffic_limit_bytes"] = nil
		}
		result = append(result, entry)
	}
	return result
}

func renderAccessEntries(inputs []SubscriptionInput) []any {
	result := make([]any, 0, len(inputs))
	for _, input := range inputs {
		entry := map[string]any{
			"subscription_id": input.SubscriptionID,
			"user_id":         input.UserID,
			"plan_id":         input.PlanID,
			"vless_client_id": input.SubscriptionID,
			"email":           fmt.Sprintf("subscription:%s", input.SubscriptionID),
			"flow":            "xtls-rprx-vision",
			"device_limit":    input.DeviceLimit,
			"expires_at":      input.ExpiresAt,
		}
		if input.TrafficLimitBytes != nil {
			entry["traffic_limit_bytes"] = *input.TrafficLimitBytes
		} else {
			entry["traffic_limit_bytes"] = nil
		}
		result = append(result, entry)
	}
	return result
}

func renderClients(accessEntries []any) []any {
	result := make([]any, 0, len(accessEntries))
	for _, raw := range accessEntries {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, map[string]any{
			"id":    entry["vless_client_id"],
			"email": entry["email"],
			"flow":  entry["flow"],
		})
	}
	return result
}
