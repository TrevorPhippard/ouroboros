package server

import "strings"

// originAllowed reports whether a browser Origin may open a WebSocket. It
// mirrors the CORS allow-list so the two can never disagree. Requests without
// an Origin header (non-browser clients) are allowed, as they are by CORS.
func originAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return true
	}
	for _, candidate := range allowed {
		if candidate == "*" || strings.EqualFold(candidate, origin) {
			return true
		}
	}
	return false
}
