package main

import (
	"testing"

	"github.com/mralaminahamed/sitemon/packages/shared/cache"
)

func TestStep(t *testing.T) {
	// threshold 2: one blip does not page.
	s := cache.AlertState{}
	s, a := step(s, "DOWN", 2)
	if a != "" || s.Fails != 1 || s.Alerted {
		t.Fatalf("first DOWN: got alert=%q state=%+v", a, s)
	}
	s, a = step(s, "DOWN", 2)
	if a != "down" || !s.Alerted || s.Fails != 2 {
		t.Fatalf("second DOWN: got alert=%q state=%+v", a, s)
	}
	// stays down: no repeat alert.
	s, a = step(s, "DOWN", 2)
	if a != "" || !s.Alerted {
		t.Fatalf("third DOWN: got alert=%q state=%+v", a, s)
	}
	// recovery fires once, resets.
	s, a = step(s, "UP", 2)
	if a != "recovery" || s.Alerted || s.Fails != 0 {
		t.Fatalf("recovery: got alert=%q state=%+v", a, s)
	}
	// steady up: nothing.
	if _, a = step(s, "UP", 2); a != "" {
		t.Fatalf("steady up: got alert=%q", a)
	}
}

func TestStepThresholdOne(t *testing.T) {
	// threshold 1: first DOWN pages immediately.
	if s, a := step(cache.AlertState{}, "DOWN", 1); a != "down" || !s.Alerted {
		t.Fatalf("threshold 1 first DOWN: got alert=%q state=%+v", a, s)
	}
}

func TestStepFirstUpNoRecovery(t *testing.T) {
	if s, a := step(cache.AlertState{}, "UP", 2); a != "" || s.Alerted {
		t.Fatalf("first UP should not recover: got alert=%q state=%+v", a, s)
	}
}
