package singbox

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildConfigConvertsVLESSAndStats(t *testing.T) {
	template := []byte(`{"log":{"level":"info"},"outbounds":[{"type":"direct","tag":"direct"}]}`)
	config, err := BuildConfig(template, []PanelInbound{{
		Enable:   true,
		Port:     443,
		Protocol: "vless",
		Tag:      "vless-443",
		Settings: []byte(`{"clients":[
			{"email":"active@example.com","id":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision","enable":true},
			{"email":"disabled@example.com","id":"22222222-2222-2222-2222-222222222222","enable":true}
		]}`),
		StreamSettings: []byte(`{"network":"ws","security":"tls","wsSettings":{"path":"/vpn","headers":{"Host":"vpn.example.com"}},"tlsSettings":{"serverName":"vpn.example.com","certificates":[{"certificateFile":"/etc/max-ui/cert/fullchain.pem","keyFile":"/etc/max-ui/cert/private.key"}]}}`),
		Sniffing:       []byte(`{"enabled":true,"routeOnly":false}`),
		EnabledByEmail: map[string]bool{"active@example.com": true, "disabled@example.com": false},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var root map[string]any
	if err := json.Unmarshal(config, &root); err != nil {
		t.Fatal(err)
	}
	inbounds := root["inbounds"].([]any)
	if len(inbounds) != 1 {
		t.Fatalf("expected one generated inbound, got %d", len(inbounds))
	}
	inbound := inbounds[0].(map[string]any)
	users := inbound["users"].([]any)
	if len(users) != 1 || users[0].(map[string]any)["name"] != "active@example.com" {
		t.Fatalf("disabled client leaked into generated config: %#v", users)
	}
	transport := inbound["transport"].(map[string]any)
	if transport["type"] != "ws" || transport["path"] != "/vpn" {
		t.Fatalf("websocket transport was not converted: %#v", transport)
	}
	tlsConfig := inbound["tls"].(map[string]any)
	if tlsConfig["certificate_path"] != "/etc/max-ui/cert/fullchain.pem" {
		t.Fatalf("TLS certificate path was not converted: %#v", tlsConfig)
	}
	experimental := root["experimental"].(map[string]any)
	api := experimental["v2ray_api"].(map[string]any)
	stats := api["stats"].(map[string]any)
	if stats["enabled"] != true || len(stats["users"].([]any)) != 1 {
		t.Fatalf("statistics API was not generated: %#v", stats)
	}
}

func TestBuildConfigRejectsUnsupportedEnabledInbound(t *testing.T) {
	_, err := BuildConfig([]byte(`{"outbounds":[]}`), []PanelInbound{{
		Enable: true, Port: 51820, Protocol: "wireguard", Tag: "wg",
		Settings: []byte(`{}`),
	}})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected an actionable unsupported-protocol error, got %v", err)
	}
}

func TestBuildConfigIgnoresDisabledUnsupportedInbound(t *testing.T) {
	config, err := BuildConfig([]byte(`{"outbounds":[]}`), []PanelInbound{{
		Enable: false, Port: 51820, Protocol: "wireguard", Tag: "wg",
		Settings: []byte(`{}`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(config), `"tag": "wg"`) {
		t.Fatalf("disabled inbound appeared in generated config: %s", config)
	}
}
