// Command ai is the sitemon ai service.
//
// Phase 0: stub. It exposes GET /health so the container is orchestratable and
// the monorepo builds end to end. Real logic lands in later phases (see
// PLAN.md).
package main

import (
	"log"

	"github.com/mralaminahamed/sitemon/packages/shared/health"
)

func main() {
	addr := health.AddrFromEnv(":8090")
	log.Printf("sitemon ai (phase 0 stub) listening on %s", addr)
	log.Fatal(health.Serve("ai", addr))
}
