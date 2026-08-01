package singbox

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const DefaultAPIPort = 62791

// PanelInbound is the core-neutral part of a Max-Ui inbound needed to build a
// sing-box configuration. Keeping the database model out of this package makes
// the conversion deterministic and easy to test.
type PanelInbound struct {
	Enable          bool
	Listen          string
	Port            int
	Protocol        string
	Tag             string
	Settings        []byte
	StreamSettings  []byte
	Sniffing        []byte
	EnabledByEmail  map[string]bool
}

func SupportsProtocol(protocol string) bool {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "vless", "vmess", "trojan", "shadowsocks", "mixed", "http":
		return true
	default:
		return false
	}
}

func ValidateInbound(inbound PanelInbound) error {
	if !inbound.Enable {
		return nil
	}
	_, _, err := convertInbound(inbound)
	return err
}

// BuildConfig combines the operator's sing-box base template with the inbounds
// and users managed by Max-Ui. Template inbounds are deliberately replaced: the
// database is the single source of truth, matching S-UI's configuration model.
func BuildConfig(template []byte, panelInbounds []PanelInbound) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(template, &root); err != nil {
		return nil, fmt.Errorf("invalid sing-box template: %w", err)
	}

	inbounds := make([]any, 0, len(panelInbounds))
	inboundTags := make([]string, 0, len(panelInbounds))
	users := make([]string, 0)
	seenUsers := make(map[string]struct{})
	for _, inbound := range panelInbounds {
		if !inbound.Enable {
			continue
		}
		converted, names, err := convertInbound(inbound)
		if err != nil {
			return nil, fmt.Errorf("inbound %q (%s): %w", inbound.Tag, inbound.Protocol, err)
		}
		inbounds = append(inbounds, converted)
		inboundTags = append(inboundTags, inbound.Tag)
		for _, name := range names {
			if _, exists := seenUsers[name]; exists {
				continue
			}
			seenUsers[name] = struct{}{}
			users = append(users, name)
		}
	}
	root["inbounds"] = inbounds

	outboundTags := make([]string, 0)
	if outbounds, ok := root["outbounds"].([]any); ok {
		for _, raw := range outbounds {
			if outbound, ok := raw.(map[string]any); ok {
				if tag, _ := outbound["tag"].(string); tag != "" {
					outboundTags = append(outboundTags, tag)
				}
			}
		}
	}

	experimental, _ := root["experimental"].(map[string]any)
	if experimental == nil {
		experimental = make(map[string]any)
	}
	experimental["v2ray_api"] = map[string]any{
		"listen": net.JoinHostPort("127.0.0.1", strconv.Itoa(DefaultAPIPort)),
		"stats": map[string]any{
			"enabled":   true,
			"inbounds":  inboundTags,
			"outbounds": outboundTags,
			"users":     users,
		},
	}
	root["experimental"] = experimental

	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal generated sing-box config: %w", err)
	}
	return data, nil
}

func convertInbound(source PanelInbound) (map[string]any, []string, error) {
	protocol := strings.ToLower(strings.TrimSpace(source.Protocol))
	if !SupportsProtocol(protocol) {
		return nil, nil, fmt.Errorf("protocol is not supported by the Max-Ui sing-box adapter; disable it or use Xray")
	}
	if source.Port < 1 || source.Port > 65535 {
		return nil, nil, fmt.Errorf("invalid listen port %d", source.Port)
	}

	var settings map[string]any
	if err := json.Unmarshal(source.Settings, &settings); err != nil {
		return nil, nil, fmt.Errorf("invalid settings JSON: %w", err)
	}
	result := map[string]any{
		"type":        protocol,
		"tag":         source.Tag,
		"listen":      defaultListen(source.Listen),
		"listen_port": source.Port,
	}

	var names []string
	var err error
	switch protocol {
	case "vless", "vmess", "trojan":
		result["users"], names, err = convertUsers(protocol, settings["clients"], source.EnabledByEmail)
		if err != nil {
			return nil, nil, err
		}
		if fallbacks, ok := settings["fallbacks"].([]any); ok && len(fallbacks) > 0 {
			return nil, nil, fmt.Errorf("Xray fallback rules cannot be translated safely")
		}
	case "shadowsocks":
		method, _ := settings["method"].(string)
		password, _ := settings["password"].(string)
		if method == "" || password == "" {
			return nil, nil, fmt.Errorf("method and server password are required")
		}
		result["method"] = normalizeSSMethod(method)
		result["password"] = password
		if network, _ := settings["network"].(string); network != "" && network != "tcp,udp" {
			result["network"] = network
		}
		clients, found := settings["clients"].([]any)
		if found && len(clients) > 0 {
			if !strings.HasPrefix(method, "2022-") {
				return nil, nil, fmt.Errorf("multi-user Shadowsocks requires a 2022 method in sing-box")
			}
			result["users"], names, err = convertUsers(protocol, clients, source.EnabledByEmail)
			if err != nil {
				return nil, nil, err
			}
		}
	case "mixed", "http":
		accounts, _ := settings["accounts"].([]any)
		result["users"], names, err = convertUsers(protocol, accounts, source.EnabledByEmail)
		if err != nil {
			return nil, nil, err
		}
	}

	if err := applyStream(result, source.StreamSettings); err != nil {
		return nil, nil, err
	}
	applySniffing(result, source.Sniffing)
	return result, names, nil
}

func convertUsers(protocol string, raw any, enabled map[string]bool) ([]any, []string, error) {
	clients, _ := raw.([]any)
	users := make([]any, 0, len(clients))
	names := make([]string, 0, len(clients))
	for _, item := range clients {
		client, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := client["email"].(string)
		if name == "" {
			name, _ = client["user"].(string)
		}
		if name == "" {
			return nil, nil, fmt.Errorf("client is missing an email/name")
		}
		if value, exists := client["enable"].(bool); exists && !value {
			continue
		}
		if value, exists := enabled[name]; exists && !value {
			continue
		}

		user := map[string]any{"name": name}
		switch protocol {
		case "vless", "vmess":
			uuid, _ := client["id"].(string)
			if uuid == "" {
				return nil, nil, fmt.Errorf("client %q is missing a UUID", name)
			}
			user["uuid"] = uuid
			if protocol == "vless" {
				if flow, _ := client["flow"].(string); flow != "" {
					if flow == "xtls-rprx-vision-udp443" {
						flow = "xtls-rprx-vision"
					}
					user["flow"] = flow
				}
			} else {
				user["alterId"] = 0
			}
		case "trojan", "shadowsocks":
			password, _ := client["password"].(string)
			if password == "" {
				return nil, nil, fmt.Errorf("client %q is missing a password", name)
			}
			user["password"] = password
		case "mixed", "http":
			password, _ := client["pass"].(string)
			if password == "" {
				return nil, nil, fmt.Errorf("client %q is missing a password", name)
			}
			user["password"] = password
		}
		users = append(users, user)
		names = append(names, name)
	}
	return users, names, nil
}

func applyStream(inbound map[string]any, raw []byte) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil
	}
	var stream map[string]any
	if err := json.Unmarshal(raw, &stream); err != nil {
		return fmt.Errorf("invalid stream settings JSON: %w", err)
	}

	security, _ := stream["security"].(string)
	switch security {
	case "", "none":
	case "tls":
		tlsSettings, _ := stream["tlsSettings"].(map[string]any)
		tls, err := convertTLS(tlsSettings)
		if err != nil {
			return err
		}
		inbound["tls"] = tls
	case "reality":
		realitySettings, _ := stream["realitySettings"].(map[string]any)
		tls, err := convertReality(realitySettings)
		if err != nil {
			return err
		}
		inbound["tls"] = tls
	default:
		return fmt.Errorf("stream security %q is not supported by sing-box", security)
	}

	network, _ := stream["network"].(string)
	switch network {
	case "", "tcp", "raw":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		if kind, _ := header["type"].(string); kind != "" && kind != "none" {
			return fmt.Errorf("TCP header type %q is not supported by sing-box", kind)
		}
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		transport := map[string]any{"type": "ws"}
		copyString(transport, "path", ws, "path")
		if headers, ok := ws["headers"].(map[string]any); ok && len(headers) > 0 {
			transport["headers"] = headers
		}
		inbound["transport"] = transport
	case "grpc":
		grpcSettings, _ := stream["grpcSettings"].(map[string]any)
		transport := map[string]any{"type": "grpc"}
		copyString(transport, "service_name", grpcSettings, "serviceName")
		inbound["transport"] = transport
	case "httpupgrade":
		settings, _ := stream["httpupgradeSettings"].(map[string]any)
		transport := map[string]any{"type": "httpupgrade"}
		copyString(transport, "path", settings, "path")
		copyString(transport, "host", settings, "host")
		inbound["transport"] = transport
	case "xhttp", "splithttp":
		settings, _ := stream["xhttpSettings"].(map[string]any)
		transport := map[string]any{"type": "http"}
		copyString(transport, "path", settings, "path")
		if host, _ := settings["host"].(string); host != "" {
			transport["host"] = []string{host}
		}
		inbound["transport"] = transport
	default:
		return fmt.Errorf("transport %q is not supported by the Max-Ui sing-box adapter", network)
	}
	return nil
}

func convertTLS(source map[string]any) (map[string]any, error) {
	result := map[string]any{"enabled": true}
	copyString(result, "server_name", source, "serverName")
	copyString(result, "min_version", source, "minVersion")
	copyString(result, "max_version", source, "maxVersion")
	if alpn, ok := source["alpn"].([]any); ok && len(alpn) > 0 {
		result["alpn"] = alpn
	}
	certificates, _ := source["certificates"].([]any)
	if len(certificates) == 0 {
		return nil, fmt.Errorf("TLS requires a certificate")
	}
	certificate, _ := certificates[0].(map[string]any)
	if path, _ := certificate["certificateFile"].(string); path != "" {
		keyPath, _ := certificate["keyFile"].(string)
		if keyPath == "" {
			return nil, fmt.Errorf("TLS private key file is missing")
		}
		result["certificate_path"] = path
		result["key_path"] = keyPath
		return result, nil
	}
	cert := joinPEMLines(certificate["certificate"])
	key := joinPEMLines(certificate["key"])
	if cert == "" || key == "" {
		return nil, fmt.Errorf("TLS certificate or private key is empty")
	}
	result["certificate"] = cert
	result["key"] = key
	return result, nil
}

func convertReality(source map[string]any) (map[string]any, error) {
	privateKey, _ := source["privateKey"].(string)
	target, _ := source["target"].(string)
	if privateKey == "" || target == "" {
		return nil, fmt.Errorf("Reality private key and target are required")
	}
	host, portText, err := net.SplitHostPort(target)
	if err != nil {
		return nil, fmt.Errorf("invalid Reality target %q: %w", target, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return nil, fmt.Errorf("invalid Reality target port: %w", err)
	}
	reality := map[string]any{
		"enabled":     true,
		"private_key": privateKey,
		"handshake": map[string]any{
			"server":      host,
			"server_port": port,
		},
	}
	if shortIDs, ok := source["shortIds"].([]any); ok {
		reality["short_id"] = shortIDs
	}
	return map[string]any{"enabled": true, "reality": reality}, nil
}

func applySniffing(inbound map[string]any, raw []byte) {
	var sniffing map[string]any
	if json.Unmarshal(raw, &sniffing) != nil {
		return
	}
	if enabled, _ := sniffing["enabled"].(bool); enabled {
		inbound["sniff"] = true
		if routeOnly, _ := sniffing["routeOnly"].(bool); !routeOnly {
			inbound["sniff_override_destination"] = true
		}
	}
}

func defaultListen(value string) string {
	if strings.TrimSpace(value) == "" {
		return "0.0.0.0"
	}
	return value
}

func normalizeSSMethod(value string) string {
	if value == "chacha20-poly1305" {
		return "chacha20-ietf-poly1305"
	}
	return value
}

func copyString(destination map[string]any, destinationKey string, source map[string]any, sourceKey string) {
	if value, _ := source[sourceKey].(string); value != "" {
		destination[destinationKey] = value
	}
}

func joinPEMLines(value any) string {
	switch lines := value.(type) {
	case string:
		return lines
	case []any:
		parts := make([]string, 0, len(lines))
		for _, line := range lines {
			if text, ok := line.(string); ok {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}
