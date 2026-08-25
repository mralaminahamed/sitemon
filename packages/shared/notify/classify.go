package notify

// Classify maps a status transition to an alert type. An empty prev (first
// sighting) yields no alert; UP -> not-UP is "down"; not-UP -> UP is "recovery".
// It is the single source of truth for the transition rule, shared by the
// in-process AlertManager and the distributed notifier service.
func Classify(prev, status string) string {
	switch {
	case prev == "":
		return ""
	case prev == "UP" && status != "UP":
		return "down"
	case prev != "UP" && status == "UP":
		return "recovery"
	default:
		return ""
	}
}
