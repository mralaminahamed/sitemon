// Package health provides a minimal HTTP health endpoint shared by every
// service. Phase 0 services are otherwise stubs; this gives Docker
// healthchecks something real to hit and a uniform shape to grow on.
package health

import (
	"encoding/json"
	"net/http"
	"os"
)

// Serve starts an HTTP server exposing GET /health for the named service and
// blocks until it errors.
func Serve(service, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": service,
		})
	})
	return http.ListenAndServe(addr, mux)
}

// AddrFromEnv returns ":$PORT" when PORT is set, otherwise fallback. Lets
// compose/k8s override the listen port without a rebuild.
func AddrFromEnv(fallback string) string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return fallback
}
