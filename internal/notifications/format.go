package notifications

import (
	"fmt"
	"strings"
)

// formatPlainMessage formats an event as a plain text message suitable
// for providers that don't support rich formatting.
func formatPlainMessage(event Event) string {
	var sb strings.Builder
	sb.WriteString(event.Message)

	if event.AppType != "" || event.Instance != "" {
		sb.WriteString("\n")
		if event.AppType != "" {
			fmt.Fprintf(&sb, "\nApp: %s", event.AppType)
		}
		if event.Instance != "" {
			fmt.Fprintf(&sb, "\nInstance: %s", event.Instance)
		}
	}

	if len(event.Fields) > 0 {
		sb.WriteString("\n")
		for k, v := range event.Fields {
			fmt.Fprintf(&sb, "\n%s: %s", k, v)
		}
	}

	return sb.String()
}
