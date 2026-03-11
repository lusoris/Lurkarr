package notifications

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// ProviderConfig holds the raw DB fields needed to construct a Provider.
type ProviderConfig struct {
	Type   string
	Config json.RawMessage
	Events []string
}

// BuildProvider constructs a notification Provider from a type and raw JSON config.
func BuildProvider(cfg ProviderConfig) (Provider, ProviderType, []EventType, error) {
	pt := ProviderType(cfg.Type)

	var raw map[string]any
	if err := json.Unmarshal(cfg.Config, &raw); err != nil {
		return nil, "", nil, fmt.Errorf("unmarshal provider config: %w", err)
	}

	str := func(key string) string {
		v, _ := raw[key].(string)
		return v
	}
	num := func(key string) int {
		v, _ := raw[key].(float64)
		return int(v)
	}
	boolean := func(key string) bool {
		v, _ := raw[key].(bool)
		return v
	}

	var p Provider
	switch pt {
	case ProviderDiscord:
		p = NewDiscord(str("webhook_url"), str("username"), str("avatar_url"))
	case ProviderTelegram:
		p = NewTelegram(str("bot_token"), str("chat_id"))
	case ProviderPushover:
		p = NewPushover(str("api_token"), str("user_key"), str("device"), num("priority"))
	case ProviderGotify:
		p = NewGotify(str("server_url"), str("app_token"), num("priority"))
	case ProviderNtfy:
		p = NewNtfy(str("server_url"), str("topic"), str("token"), num("priority"))
	case ProviderApprise:
		var urls []string
		if rawURLs, ok := raw["urls"].([]any); ok {
			for _, u := range rawURLs {
				if s, ok := u.(string); ok {
					urls = append(urls, s)
				}
			}
		}
		p = NewApprise(str("server_url"), urls, str("tag"))
	case ProviderEmail:
		var to []string
		if rawTo, ok := raw["to"].([]any); ok {
			for _, t := range rawTo {
				if s, ok := t.(string); ok {
					to = append(to, s)
				}
			}
		}
		p = NewEmail(str("host"), num("port"), str("username"), str("password"), str("from"), to, boolean("starttls"), boolean("skip_verify"))
	case ProviderWebhook:
		headers := make(map[string]string)
		if rawHeaders, ok := raw["headers"].(map[string]any); ok {
			for k, v := range rawHeaders {
				if s, ok := v.(string); ok {
					headers[k] = s
				}
			}
		}
		p = NewWebhook(str("url"), headers)
	default:
		return nil, "", nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}

	var events []EventType
	for _, e := range cfg.Events {
		events = append(events, EventType(e))
	}

	return p, pt, events, nil
}

// LoadProviders registers all providers from the given configs, replacing any
// previously registered providers. This is typically called at startup and
// after CRUD operations on notification providers.
func (m *Manager) LoadProviders(configs []ProviderConfig) error {
	m.mu.Lock()
	// Clear existing providers.
	m.providers = make(map[ProviderType]providerEntry)
	m.mu.Unlock()

	var firstErr error
	for _, cfg := range configs {
		p, pt, events, err := BuildProvider(cfg)
		if err != nil {
			slog.Error("failed to build notification provider", "type", cfg.Type, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		m.Register(pt, p, events)
		slog.Info("registered notification provider", "type", pt, "events", len(events))
	}
	return firstErr
}
