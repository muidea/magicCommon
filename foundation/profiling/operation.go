package profiling

import (
	"net/url"
	"strings"
)

// HTTPOperation retains only allowlisted route words. Hosts, query strings,
// application/tenant IDs and credentials never become metric labels.
func HTTPOperation(method, target string) string {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
	default:
		method = "OTHER"
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return method + " invalid"
	}
	allowed := map[string]bool{"api": true, "v1": true, "application": true, "applications": true, "entity": true, "entities": true, "value": true, "values": true, "filter": true, "query": true, "create": true, "insert": true, "update": true, "delete": true, "private": true, "guarded-update": true, "changeset": true, "oauth": true, "token": true, "userinfo": true, "introspect": true, "authorization": true, "namespaces": true, "namespace": true, "runtime": true, "runtimes": true, "snapshot": true, "notifications": true, "role": true, "roles": true, "account": true, "accounts": true, "health": true, "ready": true, "todo": true, "todos": true, "system": true, "config": true}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) > 8 {
		parts = parts[:8]
	}
	for i, p := range parts {
		if !allowed[p] {
			parts[i] = "_"
		}
	}
	return method + " /" + strings.Join(parts, "/")
}
